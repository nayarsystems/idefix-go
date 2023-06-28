package messages

import (
	"encoding/base64"
	"reflect"
	"time"

	"github.com/jaracil/ei"
	"github.com/nayarsystems/mapstructure"
)

func DecodeBase64ToSliceHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Value,
		t reflect.Value) (interface{}, error) {

		if f.Kind() != reflect.String {
			return f.Interface(), nil
		}
		if t.Type() != reflect.TypeOf([]byte{}) {
			return f.Interface(), nil
		}
		// Convert it by parsing
		b64input := f.Interface().(string)
		res, err := base64.StdEncoding.DecodeString(b64input)
		return res, err
	}
}

func DecodeNumberToDurationHookFunc(unit time.Duration) mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		switch f.Kind() {
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64,
			reflect.Uintptr,
			reflect.Float32,
			reflect.Float64:
		default:
			return data, nil
		}
		if t != reflect.TypeOf(time.Duration(0)) {
			return data, nil
		}
		raw := ei.N(data).Int64Z()
		res := unit * time.Duration(raw)
		return res, nil
	}
}

func DecodeUnixMilliToTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		switch f.Kind() {
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64,
			reflect.Uintptr,
			reflect.Float32,
			reflect.Float64:
		default:
			return data, nil
		}
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}
		raw := ei.N(data).Int64Z()
		res := time.UnixMilli(raw)
		return res, nil
	}
}

func DecodeAnyTimeStringToTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {

		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}
		res, err := ei.N(data).Time()
		if err != nil {
			return data, nil
		}
		return res, nil
	}
}
