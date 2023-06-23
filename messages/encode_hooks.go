package messages

import (
	"encoding/base64"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

func EncodeMsiableToMsiHook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		input := old.Interface()
		msiable := getMsiable(input)
		new = old
		var inputMap msi
		if msiable != nil {
			inputMap, err = msiable.ToMsi()
			if err != nil {
				return
			}
			new = reflect.ValueOf(inputMap)
			handled = true
		}
		return
	}
}

func EncodeDurationToSecondsHook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		t := old.Type()
		if t != reflect.TypeOf(time.Duration(0)) {
			new = old
		} else {
			d := old.Interface().(time.Duration)
			secs := d.Seconds()
			new = reflect.ValueOf(secs)
			handled = true
		}
		return
	}
}

func EncodeDurationToStringHook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		t := old.Type()
		if t != reflect.TypeOf(time.Duration(0)) {
			new = old
		} else {
			d := old.Interface().(time.Duration)
			secs := d.String()
			new = reflect.ValueOf(secs)
			handled = true
		}
		return
	}
}

func EncodeTimeToUnixMilliHook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		t := old.Type()
		new = old
		if t == reflect.TypeOf(time.Time{}) {
			tt := old.Interface().(time.Time)
			new = reflect.ValueOf(tt.UnixMilli())
			handled = true
		}
		return
	}
}

// This hook avoids the attempt to encode time.Time as a map[string]interface{} with the fields
func EncodeTimeToTimeHook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		t := old.Type()
		new = old
		if t == reflect.TypeOf(time.Time{}) {
			handled = true
		}
		return
	}
}

func EncodeSliceToBase64Hook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		t := old.Type()
		new = old
		if t == reflect.TypeOf([]byte{}) {
			data := old.Interface().([]byte)
			newValue := base64.RawStdEncoding.EncodeToString(data)
			new = reflect.ValueOf(newValue)
			handled = true
		}
		return
	}
}
