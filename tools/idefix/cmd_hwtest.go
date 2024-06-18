package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jaracil/ei"
	"github.com/montanaflynn/stats"
	idefixgo "github.com/nayarsystems/idefix-go"
	m "github.com/nayarsystems/idefix-go/messages"
	"github.com/spf13/cobra"
)

type device struct {
	presence   bool
	product    string
	lot        string
	version    string
	flashmode  bool
	updating   bool
	testResult result
}

type result struct {
	input       []map[string]bool
	temperature []int
	output      []map[string]bool
	ok          bool
	report      []error
}

func init() {
	cmdHWTest.Flags().StringP("address", "a", "", "Device address")
	cmdHWTest.Flags().StringP("port", "p", "P0", "Port to test (P0: GSR2; P1: Specialist board)")
	cmdHWTest.MarkFlagRequired("address")

	rootCmd.AddCommand(cmdHWTest)
}

var cmdHWTest = &cobra.Command{
	Use:     "hwtest",
	Aliases: []string{"test", "hw"},
	Short:   "Hardware test: GSR2 or specialist board connected to P1",
	RunE:    cmdHWTestRunE,
}

func cmdHWTestRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	port, err := cmd.Flags().GetString("port")
	if err != nil {
		return err
	}

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	usb, err := ic.NewSubscriberStream(addr, "edev.evt.usb.tag."+port, 10, false, getTimeout(cmd))
	if err != nil {
		log.Fatalln("Cannot open stream", err)
	}

	var testDevice device

	for {
		topicRead := false
		select {
		case k := <-usb.Channel():
			topicRead = true
			testDevice.presence = ei.N(k.Data).M("PRESENT").BoolZ()
			product := strings.Split(ei.N(k.Data).M("PRODUCT").StringZ(), "-")
			testDevice.product = product[1]
			testDevice.lot = product[2]
			testDevice.version = ei.N(k.Data).M("VERSION").StringZ()
			testDevice.flashmode = ei.N(k.Data).M("FLASHMODE").BoolZ()
			testDevice.updating = ei.N(k.Data).M("UPDATING").BoolZ()
		case <-usb.Context().Done():
			return usb.Context().Err()
		case <-ic.Context().Done():
			return nil
		}
		if topicRead {
			usb.Close()
			break
		}
	}
	testDevice.testResult.hwTest(addr, ic, testDevice)

	return nil
}

func (res *result) hwTest(address string, client *idefixgo.Client, dev device) error {
	var topic string
	switch dev.product {
	case "multivoltage":
		topic = "multitension1"
		res.testPBM(topic, client, address)
		fmt.Println(res.ok)
		for _, i := range res.report {
			fmt.Printf("Error: %v\n", i)
		}

		// Loop for visual inspection
		toggle := true
		for {
			toggle = !toggle
			_, err := client.Call(address, &m.Message{To: fmt.Sprintf("%s.cmd.relay.1", topic), Data: toggle}, time.Second)
			if err != nil {
				return err
			}

			_, err = client.Call(address, &m.Message{To: fmt.Sprintf("%s.cmd.relay.2", topic), Data: !toggle}, time.Second)
			if err != nil {
				return err
			}
			time.Sleep(time.Second)
		}
	case "serial":
		topic = "serial1" // TO DO: This topic is not real
	case "core":
		topic = "coreboard"
	default:
		return fmt.Errorf("product not set: %s", topic)
	}

	return nil
}

func (res *result) testPBM(topic string, client *idefixgo.Client, address string) {
	/* The testing process will proceed as follows:
	1.- Setting the initials conditions: Relay 1 ON and relay 2 OFF (Odd inputs to true and even to false)
	2.- Starting measurements:
		2.1.- Temperature measurements -> Some temperature measurements will be done and at the end of the process an average will be computed
		2.2.- Start monitoring inputs (subscribe to its path of events)
		2.3.- Toggle the relays (check that the relays has toggled) and get the inputs values
		2.4.- Toggle again the relays and get the new inputs values
		2.5.- Loop of continuously toggling the relays every 2s for the tester to visually check the LEDs
	*/
	res.temperature = make([]int, 0)
	res.input = make([]map[string]bool, 2)
	res.output = make([]map[string]bool, 2)
	toggleRel := true

	_, err := client.Call(address, &m.Message{To: fmt.Sprintf("%s.cmd.relay.1", topic), Data: toggleRel}, time.Second*10)
	if err != nil {
		return
	}

	_, err = client.Call(address, &m.Message{To: fmt.Sprintf("%s.cmd.relay.2", topic), Data: !toggleRel}, time.Second*10)
	if err != nil {
		return
	}

	temp, err := client.NewSubscriberStream(address, fmt.Sprintf("%s.evt.temperature", topic), 10, false, time.Minute*10)
	if err != nil {
		log.Fatalln("Cannot open stream:", err)
	}
	defer temp.Close()

	go func() {
		for t := range temp.Channel() {
			res.temperature = append(res.temperature, ei.N(t.Data).IntZ())
		}
	}()

	inputs, err := client.NewSubscriberStream(address, fmt.Sprintf("%s.evt.input", topic), 100, false, time.Minute*10)
	if err != nil {
		log.Fatalln("Cannot open stream:", err)
	}
	defer inputs.Close()

	rel, err := client.NewSubscriberStream(address, fmt.Sprintf("%s.evt.relay", topic), 10, false, time.Minute)
	if err != nil {
		log.Fatalln("Cannot open stream:", err)
	}
	defer rel.Close()

	res.output[0] = make(map[string]bool)
	res.input[0] = make(map[string]bool)
	res.output[1] = make(map[string]bool)
	res.input[1] = make(map[string]bool)

	j := 0

	go func() {
		for {

			select {
			case r := <-rel.Channel():
				out := fmt.Sprintf("%s%s", strings.Split(r.To, ".")[2], strings.Split(r.To, ".")[3])
				res.output[j][out] = ei.N(r.Data).BoolZ()
			case i := <-inputs.Channel():
				in := fmt.Sprintf("%s%s", strings.Split(i.To, ".")[2], strings.Split(i.To, ".")[3])
				res.input[j][in] = ei.N(i.Data).BoolZ()
			}
		}
	}()

	toggleRel = !toggleRel
	_, err = client.Call(address, &m.Message{To: fmt.Sprintf("%s.cmd.relay.1", topic), Data: toggleRel}, time.Second)
	if err != nil {
		return
	}

	_, err = client.Call(address, &m.Message{To: fmt.Sprintf("%s.cmd.relay.2", topic), Data: !toggleRel}, time.Second)
	if err != nil {
		return
	}

	time.Sleep(time.Second)

	j++

	toggleRel = !toggleRel
	_, err = client.Call(address, &m.Message{To: fmt.Sprintf("%s.cmd.relay.1", topic), Data: toggleRel}, time.Second)
	if err != nil {
		return
	}

	_, err = client.Call(address, &m.Message{To: fmt.Sprintf("%s.cmd.relay.2", topic), Data: !toggleRel}, time.Second)
	if err != nil {
		return
	}
	time.Sleep(time.Second)

	// Analyse results
	/* How to know if the results are correct:
	1.- Temperature -> Compute the average and the deviation of the measurements done
	2.- Inputs:
		2.1.- Check if there are 16 data inputs in each measurement
		2.2.- Check that every input is opposite to the following one
		2.3.- Check that every input has changed its value from one measurement to another
	3.- Outputs:
		3.1.- Check that there are 2 changes stored
		3.2.- Check that both states are opposite
	*/

	temperature, input, output := true, true, true
	// Temperature
	avgTemp, err := stats.Mean(stats.LoadRawData(res.temperature))
	if err != nil {
		fmt.Println(err)
	}
	sdTemp, err := stats.StandardDeviation(stats.LoadRawData(res.temperature))
	if err != nil {
		fmt.Println(err)
	}

	if avgTemp > 45 && avgTemp < 20 && sdTemp > 1 {
		res.report = append(res.report, fmt.Errorf("temperature is not ok (avg:%v sd:%v)", avgTemp, sdTemp))
		temperature = false
	}

	// Inputs
	for idx, in := range res.input {
		var noInput []int
		var meas string
		if idx == 0 {
			meas = "first"
		} else {
			meas = "second"
		}

		if len(in) != 16 {
			for i := 1; i <= 16; i++ {
				_, ok := in[fmt.Sprintf("input%d", i)]
				if !ok {
					noInput = append(noInput, i)
				}
			}
			res.report = append(res.report, fmt.Errorf("%v measurement is incomplete: inputs %v are not present", meas, noInput))
			input = false
		}

		var oddIn, evenIn bool
		for i := 1; i <= 16; i += 2 {
			oddIn = oddIn || in[fmt.Sprintf("input%d", i)]
			evenIn = evenIn || in[fmt.Sprintf("input%d", i+1)]
		}
		if oddIn == evenIn {
			res.report = append(res.report, fmt.Errorf("%v measurement: two consecutive inputs have the same value", meas))
			input = false
		}
	}

	for i := 1; i <= 16; i++ {
		if res.input[0][fmt.Sprintf("input%d", i)] == res.input[1][fmt.Sprintf("input%d", i)] {
			res.report = append(res.report, fmt.Errorf("input %v value has not changed", i))
			input = false
		}
	}

	// Output
	for idx, out := range res.output {
		if len(out) != 2 {
			res.report = append(res.report, fmt.Errorf("incorrect number of relays activation %v", idx))
			output = false
		}
	}

	for i := 1; i <= 2; i++ {
		if res.output[0][fmt.Sprintf("relay%d", i)] == res.output[1][fmt.Sprintf("relay%d", i)] {
			res.report = append(res.report, fmt.Errorf("relay %d has not changed", i))
			output = false
		}
	}

	if temperature && input && output {
		res.ok = true
	}

}
