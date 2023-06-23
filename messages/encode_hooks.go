package messages

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

func EncodeMsiableToMsiHook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		//fmt.Printf("oldkind: %v, old: %v\n", old.Kind(), old)
		oldObj := old.Interface()
		newObj, err := recursiveToMsi(oldObj)
		if err != nil {
			return
		}
		return reflect.ValueOf(newObj), true, nil
	}
}

func recursiveToMsi(input any) (output any, err error) {
	ivalue := reflect.ValueOf(input)
	if ivalue.Kind() == reflect.Ptr || ivalue.Kind() == reflect.Interface {
		ivalue = ivalue.Elem()
	}
	if ivalue.Kind() == reflect.Struct {
		oldObj := ivalue.Interface()
		output, err = ToMsi(oldObj)
		return
	}
	if ivalue.Kind() == reflect.Slice {
		output = []any{}
		for i := 0; i < ivalue.Len(); i++ {
			iv := ivalue.Index(i)
			ov, err := recursiveToMsi(iv.Interface())
			if err != nil {
				return nil, err
			}
			output = append(output.([]any), ov)
		}
		return
	}
	if ivalue.Kind() == reflect.Map {
		newMap := msi{}
		if ivalue.Type().Key() != reflect.TypeOf("") {
			return nil, fmt.Errorf("map key is not string")
		}
		for _, key := range ivalue.MapKeys() {
			iv := ivalue.MapIndex(key)
			ov, err := recursiveToMsi(iv.Interface())
			if err != nil {
				return nil, err
			}
			keyStr := key.Interface().(string)
			newMap[keyStr] = ov
		}
		output = newMap
		return
	}
	return input, nil
}

func EncodeDurationToSecondsInt64Hook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		t := old.Type()
		if t != reflect.TypeOf(time.Duration(0)) {
			new = old
		} else {
			d := old.Interface().(time.Duration)
			secs := int64(d.Seconds())
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

func EncodeTimeToStringHook(layout string) mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		t := old.Type()
		new = old
		if t == reflect.TypeOf(time.Time{}) {
			tt := old.Interface().(time.Time)
			ttStr := tt.Format(layout)
			new = reflect.ValueOf(ttStr)
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
