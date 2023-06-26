package messages

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/mitchellh/mapstructure"
)

// https://glucn.medium.com/golang-an-interface-holding-a-nil-value-is-not-nil-bb151f472cc7
// https://stackoverflow.com/questions/13476349/check-for-nil-and-nil-interface-in-go
func InterfaceIsNil(b interface{}) bool {
	return b == nil || (reflect.ValueOf(b).Kind() == reflect.Ptr && reflect.ValueOf(b).IsNil())
}

/*
A convention in idefix is that all messages must be published using the map[string]interface{} type (msi).
In this way it facilitates all messages to be compatible with json/msgpack (de)serializers to allow them to be forwarded outside via a transport module (e.g. MQTT)
Making all messages of type msi also allows the message's fields normalization by the transport module (which uses the "normalize" package)
For this purpose this package offers the ToMsi() and ParseMsi() functions to to transform struct based types to msi and viceversa.
*/
type msi = map[string]interface{}

// A type struct can implement the Msiable interfaces to offer an alternative implementation to ToMsi()
type Msiable interface {
	ToMsi() (msi, error)
}

// A type struct can implement the MsiParser interfaces to offer an alternative implementation to ParseMsi()
type MsiParser interface {
	ParseMsi(msi msi) error
}

// Outputs a msi from struct or msi. It uses mapstructure by default.
func ToMsi(input any) (msi, error) {
	if InterfaceIsNil(input) {
		return msi{}, nil
	}
	msiable := getMsiable(input)
	if msiable != nil {
		return msiable.ToMsi()
	}
	return ToMsiGeneric(input, nil)
}

// Returns msiable object from input. Nil if input does not implement Msiable interface
func getMsiable(input any) (msiable Msiable) {
	// If the input implements the Msiable interface using a value receiver
	// the "type assertion" will work for both cases of input (pointer or value)
	msiable, ok := input.(Msiable)
	if !ok {
		// In case the input implements Msiable interface using a pointer receiver
		// we need to create a pointer to the value before the type assertion.
		iType := reflect.TypeOf(input)
		iKind := iType.Kind()
		if iKind != reflect.Ptr {
			// Let's create a pointer to the value and retry the type assertion
			x := reflect.New(iType)
			x.Elem().Set(reflect.ValueOf(input))
			input2 := x.Interface()
			msiable, _ = input2.(Msiable)
		}
	}
	return
}

// Outputs a msi from struct or msi using mapstructure (optionally with a EncodeFieldMapHookFunc)
func ToMsiGeneric(input any, inputHook mapstructure.EncodeFieldMapHookFunc) (msi, error) {
	output := msi{}
	var hook mapstructure.EncodeFieldMapHookFunc
	persistentHooks := mapstructure.ComposeEncodeFieldMapHookFunc(EncodeSliceToBase64Hook(), EncodeMsiableToMsiHook())
	if inputHook == nil {
		hook = persistentHooks
	} else {
		hook = mapstructure.ComposeEncodeFieldMapHookFunc(inputHook, persistentHooks)
	}
	cfg := mapstructure.DecoderConfig{
		Result:             &output,
		EncodeFieldMapHook: hook,
	}
	decoder, err := mapstructure.NewDecoder(&cfg)
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(input)
	if err != nil {
		return nil, err
	}
	return output, err
}

// Fills a struct (given by reference) or msi from a msi. It uses mapstructure by default.
func ParseMsi(input msi, output any) error {
	outputValue := reflect.ValueOf(output)
	if outputValue.Kind() != reflect.Pointer || outputValue.IsNil() {
		return fmt.Errorf("output not a pointer")
	}
	outputMsiParser, ok := output.(MsiParser)
	if ok {
		return outputMsiParser.ParseMsi(input)
	}
	return ParseMsiGeneric(input, output, nil)
}

func ParseMsiGeneric(input msi, output any, hookFunc mapstructure.DecodeHookFunc) error {
	persistentHook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToTimeHookFunc(time.RFC3339),
		DecodeBase64ToSliceHookFunc())
	var hooks any
	if hookFunc == nil {
		hooks = persistentHook
	} else {
		hooks = mapstructure.ComposeDecodeHookFunc(hookFunc, persistentHook)
	}
	cfg := mapstructure.DecoderConfig{
		Result:     output,
		DecodeHook: hooks,
	}
	decoder, err := mapstructure.NewDecoder(&cfg)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

func ParseMsg(input any, output any) error {
	inputMsi, err := MsiCast(input)
	if err != nil {
		return err
	}
	return ParseMsi(inputMsi, output)
}

func MsiCast(input any) (msi, error) {
	if inputMsi, ok := input.(msi); ok {
		return inputMsi, nil
	}
	return nil, fmt.Errorf("%v (%T) is not a msi", input, input)
}

// Gets the schema Id from a bstates based event's type field.
// Example:
//
// - input: "application/vnd.nayar.bstates; id=\"oEM5eJzBBGbyT9CLrSKrQwdnP2C+CVM8JHjfA0g3MAB=\""
//
// - oputput: "oEM5eJzBBGbyT9CLrSKrQwdnP2C+CVM8JHjfA0g3MAB="
func BstatesParseSchemaIdFromType(evtype string) (string, error) {
	r := regexp.MustCompile(`^application/vnd.nayar.bstates; id=([a-zA-Z0-9+/=]+)|"([a-zA-Z0-9+/=]+)"$`)

	matches := r.FindStringSubmatch(evtype)
	if matches == nil {
		return "", fmt.Errorf("no bstates type")
	}

	return matches[1] + matches[2], nil
}

func IsMsi(input any) bool {
	_, ok := input.(msi)
	return ok
}

func TimeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}

func getActualValue(old reflect.Value) reflect.Value {
	if old.Kind() == reflect.Ptr || old.Kind() == reflect.Interface {
		old = old.Elem()
	}
	return old
}
