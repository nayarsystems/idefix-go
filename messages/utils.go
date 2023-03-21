package messages

import (
	"fmt"
	"reflect"
	"regexp"

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
	ParseMsi(input any) error
}

// Outputs a msi from struct or msi. It uses mapstructure by default.
func ToMsi(input any) (msi, error) {
	msiableType := reflect.TypeOf((*Msiable)(nil)).Elem()
	inputValue := reflect.ValueOf(input)
	inputType := inputValue.Type()
	if inputType.Implements(msiableType) {
		rvalues := inputValue.MethodByName("ToMsi").Call([]reflect.Value{})
		if len(rvalues) != 2 {
			return nil, fmt.Errorf("fix me: ToMsi must return 2 values")
		}
		e := rvalues[1].Interface()
		if e != nil {
			return nil, e.(error)
		}
		v := rvalues[0].Interface()
		if v != nil {
			return v.(msi), nil
		}
		return nil, nil
	}
	output := msi{}
	err := mapstructure.Decode(input, &output)
	return output, err
}

// Fills a struct (given by reference) or msi from a msi. It uses mapstructure by default.
func ParseMsi(input any, output any) error {
	if !IsMsi(input) {
		return fmt.Errorf("input is not a map")
	}
	msiParserType := reflect.TypeOf((*MsiParser)(nil)).Elem()
	outputValue := reflect.ValueOf(output)

	if outputValue.Kind() != reflect.Pointer || outputValue.IsNil() {
		return fmt.Errorf("output not a pointer")
	}
	outputType := outputValue.Type()

	if outputType.Implements(msiParserType) {
		rvalues := outputValue.MethodByName("ParseMsi").Call([]reflect.Value{reflect.ValueOf(input)})
		if len(rvalues) != 1 {
			return fmt.Errorf("fix me: ParseMsi must returns a value of type error")
		}
		ret := rvalues[0].Interface()
		if ret != nil {
			return ret.(error)
		}
		return nil
	}
	return mapstructure.Decode(input, output)
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
