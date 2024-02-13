package messages

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"time"

	"github.com/nayarsystems/mapstructure"
)

func EncodeMsiableToMsiHook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		new = old
		msiValue, err := recursiveToMsi(old)
		if err != nil {
			return
		}
		new = msiValue
		handled = true
		return
	}
}

func recursiveToMsi(input reflect.Value) (output reflect.Value, err error) {
	var ivalue reflect.Value
	var valid bool
	var newObj any
	if ValueIsNil(input) {
		goto unhandled
	}
	ivalue, valid = getActualValue(input)
	if !valid {
		goto unhandled
	}
	switch ivalue.Kind() {
	case reflect.Struct:
		oldObj := ivalue.Interface()
		newObj, err = ToMsi(oldObj)
		if err != nil {
			return input, err
		}
	case reflect.Slice:
		switch ivalue.Type().Elem().Kind() {
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint,
			reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32,
			reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String:
			goto unhandled
		}
		newObj = []any{}
		for i := 0; i < ivalue.Len(); i++ {
			iv := ivalue.Index(i)
			ov, err := recursiveToMsi(iv)
			if err != nil {
				return input, err
			}
			newObj = append(newObj.([]any), ov.Interface())
		}
	case reflect.Map:
		newMap := msi{}
		if ivalue.Type().Key() != reflect.TypeOf("") {
			return input, fmt.Errorf("map key is not string")
		}
		for _, key := range ivalue.MapKeys() {
			iv := ivalue.MapIndex(key)
			ov, err := recursiveToMsi(iv)
			if err != nil {
				return input, err
			}
			keyStr := key.Interface().(string)
			newMap[keyStr] = ov.Interface()
		}
		newObj = newMap
	}
	if newObj != nil {
		return reflect.ValueOf(newObj), nil
	}
unhandled:
	return input, nil
}

func EncodeDurationToSecondsInt64Hook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		new = old
		actual, valid := getActualValue(old)
		if !valid {
			return
		}
		t := actual.Type()
		if t == reflect.TypeOf(time.Duration(0)) {
			d := actual.Interface().(time.Duration)
			secs := int64(d.Seconds())
			new = reflect.ValueOf(secs)
			handled = true
		}
		return
	}
}

func EncodeDurationToStringHook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		new = old
		actual, valid := getActualValue(old)
		if !valid {
			return
		}
		if actual.Type() != reflect.TypeOf(time.Duration(0)) {
			return
		}
		d := actual.Interface().(time.Duration)
		secs := d.String()
		new = reflect.ValueOf(secs)
		handled = true
		return
	}
}

func EncodeTimeToUnixMilliHook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		new = old
		actual, valid := getActualValue(old)
		if !valid {
			return
		}
		if actual.Type() != reflect.TypeOf(time.Time{}) {
			return
		}
		tt := actual.Interface().(time.Time)
		new = reflect.ValueOf(tt.UnixMilli())
		handled = true
		return
	}
}

func EncodeTimeToStringHook(layout string) mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		new = old
		actual, valid := getActualValue(old)
		if !valid {
			return
		}
		if actual.Type() != reflect.TypeOf(time.Time{}) {
			return
		}
		tt := actual.Interface().(time.Time)

		// RFC3339 representation of zero value
		// can't be parsed by golang's RFC3339 parser: https://github.com/golang/go/issues/20555
		if tt.IsZero() {
			tt = time.Unix(0, 0)
		}

		ttStr := tt.Format(layout)
		new = reflect.ValueOf(ttStr)
		handled = true
		return
	}
}

// This hook avoids the attempt to encode time.Time as a map[string]interface{} with the fields
func EncodeTimeToTimeHook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		new = old
		actual, valid := getActualValue(old)
		if !valid {
			return
		}
		if actual.Type() != reflect.TypeOf(time.Time{}) {
			return
		}
		handled = true
		return
	}
}

func EncodeByteSliceToBase64Hook() mapstructure.EncodeFieldMapHookFunc {
	return func(old reflect.Value) (new reflect.Value, handled bool, err error) {
		new = old
		actual, valid := getActualValue(old)
		if !valid {
			return
		}
		if actual.Type() != reflect.TypeOf([]byte{}) {
			return
		}
		data := actual.Interface().([]byte)
		newValue := base64.StdEncoding.EncodeToString(data)
		new = reflect.ValueOf(newValue)
		handled = true
		return
	}
}
