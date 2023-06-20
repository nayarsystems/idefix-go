package messages

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/mitchellh/mapstructure"
)

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
	inputMsiable, ok := input.(Msiable)
	if ok {
		return inputMsiable.ToMsi()
	}
	return ToMsiGeneric(input, nil)
}

// Outputs a msi from struct or msi. It uses mapstructure by default.
func ToMsiGeneric(input any, hookFunc mapstructure.DecodeHookFunc) (msi, error) {
	output := msi{}

	cfg := mapstructure.DecoderConfig{
		Result:     &output,
		DecodeHook: hookFunc,
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
	var hooks any
	if hookFunc == nil {
		hooks = Base64ToSliceHookFunc()
	} else {
		hooks = mapstructure.ComposeDecodeHookFunc(Base64ToSliceHookFunc(), hookFunc)
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

// TODO: investigate the use of mapstructure hooks to minimize custom Msiable/MsiParser
// cfg := mapstructure.DecoderConfig{
// 	Result:     output,
// 	DecodeHook: mapstructure.StringToTimeHookFunc(time.RFC1123),
// }
// decoder, err := mapstructure.NewDecoder(&cfg)
// if err != nil {
// 	return err
// }
// return decoder.Decode(input)

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
