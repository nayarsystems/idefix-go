package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/jaracil/ei"
	"github.com/nayarsystems/idefix-go/messages"
	"github.com/songgao/water"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
)

func init() {
	cmdTun.Flags().StringP("address", "a", "", "Device address")
	cmdTun.Flags().StringP("ip", "i", "172.16.1.2/24", "IP address")
	cmdTun.MarkFlagRequired("address")

	rootCmd.AddCommand(cmdTun)
}

var cmdTun = &cobra.Command{
	Use:   "tun",
	Short: "Create TUN device to connect to the device",
	RunE:  cmdTunRunE,
}

func cmdTunRunE(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	ip, err := cmd.Flags().GetString("ip")
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	iface, err := setupInterface("ifx0", ip)
	if err != nil {
		return err
	}
	defer iface.Close()

	fmt.Println("Created interface", iface.Name())

	ic, err := getConnectedClient()
	if err != nil {
		return err
	}
	defer ic.Disconnect()

	stream, err := ic.NewStream(addr, "tun.evt.frame", 100, true, time.Second*30)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		b := make([]byte, 1500)
		for {
			n, err := iface.Read(b)
			if err != nil {
				if ic.Context().Err() == nil {
					fmt.Println("Error reading from interface", err)
				}
				return
			}
			data := map[string]interface{}{"frame": b[:n]}
			ic.Publish(addr, &messages.Message{To: "tun.cmd.frame", Data: data})
		}
	}()

	fmt.Println("Connected")
	for {
		select {
		case frame, ok := <-stream.Channel():
			if !ok {
				return fmt.Errorf("stream closed")
			}

			framebytes, err := ei.N(frame.Data).M("frame").Bytes()
			if err != nil {
				fmt.Println("Error getting frame bytes", err)
				continue
			}

			iface.Write(framebytes)

		case <-ic.Context().Done():
			return nil
		}
	}
}

func setupInterface(ifaceName string, ipAddr string) (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}

	config.Name = ifaceName

	iface, err := water.New(config)
	if err != nil {
		return nil, err
	}

	tap, err := netlink.LinkByName(iface.Name())
	if err != nil {
		return nil, err
	}

	addr, err := netlink.ParseAddr(ipAddr)
	if err != nil {
		return nil, err
	}

	if err := netlink.LinkSetUp(tap); err != nil {
		return nil, err
	}

	if err := netlink.AddrAdd(tap, addr); err != nil {
		return nil, err
	}

	return iface, nil
}
