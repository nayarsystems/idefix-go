package messages

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/nayarsystems/mapstructure"
	"github.com/stretchr/testify/require"
)

func Test_ToMsi_FromMsi(t *testing.T) {
	in := msi{
		"param0": 1,
	}
	out, err := ToMsi(in)
	require.NoError(t, err)
	require.Equal(t, in, out)
}

type testType struct {
	Field0 string `mapstructure:"field0"`
}

func Test_ToMsi_FromStruct(t *testing.T) {
	in := &testType{
		Field0: "hello world",
	}
	eout := msi{
		"field0": in.Field0,
	}
	out, err := ToMsi(in)
	require.NoError(t, err)
	require.Equal(t, eout, out)
}

type testTypeMsiable struct {
	Field0 string `mapstructure:"field0"`
}

func (tt *testTypeMsiable) ToMsi() (msi, error) {
	return msi{
		"field0": tt.Field0,
		"field1": "this field does not exist",
	}, nil
}

func Test_ToMsi_FromMsiable(t *testing.T) {
	in := &testTypeMsiable{
		Field0: "hello world",
	}
	eout := msi{
		"field0": in.Field0,
		"field1": "this field does not exist",
	}
	out, err := ToMsi(in)
	require.NoError(t, err)
	require.Equal(t, eout, out)
}

func Test_ParseMsi_ToStruct(t *testing.T) {
	in := msi{
		"field0": "hello world",
	}
	tt := &testType{}
	err := ParseMsi(in, tt)
	require.NoError(t, err)
	eo := &testType{
		Field0: "hello world",
	}
	require.Equal(t, eo, tt)
}

type testTypeMsiParser struct {
	Field0 string `mapstructure:"field0"`
}

func (tt *testTypeMsiParser) ParseMsi(input msi) error {
	mapstructure.Decode(input, tt)
	tt.Field0 += " (modified)"
	return nil
}

func Test_ParseMsi_ToMsiParser(t *testing.T) {
	in := msi{
		"field0": "hello world",
	}
	tt := &testTypeMsiParser{}
	err := ParseMsi(in, tt)
	require.NoError(t, err)
	eo := &testTypeMsiParser{
		Field0: "hello world (modified)",
	}
	require.Equal(t, eo, tt)
}

func Test_ParseMsi_ToMsi(t *testing.T) {
	in := msi{
		"field0": "hello world",
	}
	tt := msi{}
	err := ParseMsi(in, &tt)
	require.NoError(t, err)
	require.Equal(t, in, tt)
}

func Test_ParseByteSlice_ParseMsi(t *testing.T) {
	type WithByteSlice struct {
		Buffer []byte `mapstructure:"buffer"`
	}
	in := msi{
		"buffer": "CgsM",
	}
	out := WithByteSlice{}
	err := ParseMsi(in, &out)
	require.NoError(t, err)
	eout := WithByteSlice{
		Buffer: []byte{0xa, 0xb, 0xc},
	}
	require.Equal(t, eout, out)
}

func Test_MapstructureDecodeNilTest(t *testing.T) {
	type mapstructureDecodeNilSliceTestStruct struct {
		RawData []byte `mapstructure:"rawData"`
	}
	input := map[string]any{
		"rawData": nil,
	}
	out := mapstructureDecodeNilSliceTestStruct{}
	err := ParseMsi(input, &out)
	require.NoError(t, err)
}

func Test_MapstructureEncodeNilInterfaceTest(t *testing.T) {
	type mapstructureEncodeNilInterfaceTestStruct struct {
		RawData any `mapstructure:"rawData"`
	}
	input := mapstructureEncodeNilInterfaceTestStruct{
		RawData: nil,
	}
	raw, err := ToMsi(input)
	require.NoError(t, err)

	out := mapstructureEncodeNilInterfaceTestStruct{}
	err = ParseMsi(raw, &out)
	require.NoError(t, err)
	require.Equal(t, input, out)
}

func Test_MapstructureEncodeNilPointerTest(t *testing.T) {
	type mapstructureEncodeNilPointerTestStruct struct {
		RawData *int `mapstructure:"rawData"`
	}
	input := mapstructureEncodeNilPointerTestStruct{
		RawData: nil,
	}
	raw, err := ToMsi(input)
	require.NoError(t, err)

	out := mapstructureEncodeNilPointerTestStruct{}
	err = ParseMsi(raw, &out)
	require.NoError(t, err)
	require.Equal(t, input, out)
}

type hooksTestStructA struct {
	Date   time.Time        `mapstructure:"date"`
	Uptime time.Duration    `mapstructure:"uptime"`
	Bytes  []byte           `mapstructure:"bytes"`
	Any    any              `mapstructure:"any"`
	Sub    hooksTestStructB `mapstructure:"sub"`
}

type hooksTestStructB struct {
	Date   time.Time     `mapstructure:"date"`
	Uptime time.Duration `mapstructure:"uptime"`
	Bytes  []byte        `mapstructure:"bytes"`
	Any    any           `mapstructure:"any"`
}

func (m *hooksTestStructB) ToMsi() (data msi, err error) {
	data, err = ToMsiGeneric(m,
		mapstructure.ComposeEncodeFieldMapHookFunc(
			EncodeTimeToStringHook(time.UnixDate),
			EncodeDurationToStringHook()))

	return data, err
}

func (m *hooksTestStructB) ParseMsi(input msi) (err error) {
	err = ParseMsiGeneric(input, m,
		mapstructure.ComposeDecodeHookFunc(
			DecodeAnyTimeStringToTimeHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
		))
	if err != nil {
		return err
	}
	return nil
}

func Test_MapstructureEncodeHooks(t *testing.T) {
	tt := time.Unix(time.Now().Unix(), 0)
	dd := time.Second * 123
	bytes := []byte("hello world")

	input := hooksTestStructA{
		Date:   tt,
		Uptime: dd,
		Bytes:  bytes,
		Any:    bytes,
		Sub: hooksTestStructB{
			Date:   tt,
			Uptime: dd,
			Bytes:  bytes,
			Any:    bytes,
		},
	}

	bytesB64 := base64.StdEncoding.EncodeToString(bytes)
	eOutput := map[string]any{
		"date":   tt.UnixMilli(),
		"uptime": int64(dd.Seconds()),
		"bytes":  bytesB64,
		"any":    bytesB64,
		"sub": map[string]any{
			"date":   tt.Format(time.UnixDate),
			"uptime": dd.String(),
			"bytes":  bytes,
			"any":    bytes,
		},
	}

	// These hooks will only be used to encode the top level struct's fields since
	// nested struct ("Sub" field) implements the Msiable interface to customize
	// its encoding to a msi. So that, time.Time and time.Duration fields will be
	// encoded differently for each struct.
	encodeHooks := mapstructure.ComposeEncodeFieldMapHookFunc(
		EncodeByteSliceToBase64Hook(),
		EncodeDurationToSecondsInt64Hook(),
		EncodeTimeToUnixMilliHook())

	output, err := ToMsiGeneric(input, encodeHooks)

	require.NoError(t, err)
	require.Equal(t, eOutput, output)
}

func Test_MapstructureDecodeHooks(t *testing.T) {
	tt := time.Unix(time.Now().Unix(), 0)
	dd := time.Second * 123
	bytes := []byte("hello world")
	bytesB64 := base64.StdEncoding.EncodeToString(bytes)

	input := map[string]any{
		"date":   tt.UnixMilli(),
		"uptime": dd.Seconds(),
		"bytes":  bytesB64,
		"any":    bytesB64,
		"sub": map[string]any{
			"date":   tt.Format(time.UnixDate),
			"uptime": dd.String(),
			"bytes":  bytesB64,
			"any":    bytesB64,
		},
	}

	eOutput := hooksTestStructA{
		Date:   tt,
		Uptime: dd,
		Bytes:  bytes,
		Any:    bytesB64,
		Sub: hooksTestStructB{
			Date:   tt,
			Uptime: dd,
			Bytes:  bytes,
			Any:    bytesB64,
		},
	}

	// These hooks will only be used to decode the top level struct's fields since
	// nested struct ("Sub" field) implements the Msiable interface to customize
	// its decoding from a msi. So that, time.Time and time.Duration fields will be
	// decoded differently for each struct.
	decodeHooks := mapstructure.ComposeDecodeHookFunc(
		// DecodeBase64ToSliceHookFunc() hook is added intrinsically (see ParseMsiGeneric func)
		DecodeNumberToDurationHookFunc(time.Second),
		DecodeUnixMilliToTimeHookFunc())
	output := hooksTestStructA{}

	err := ParseMsiGeneric(input, &output, decodeHooks)

	require.NoError(t, err)
	require.Equal(t, eOutput, output)
}

func Test_BstatesParseSchemaFromType(t *testing.T) {
	id0, err := BstatesParseSchemaIdFromType("application/vnd.nayar.bstates; id=8OTC92xYkW7CWPJGhRvqCR0U1CR6L8PhhpRGGxgW4Ts=")
	require.NoError(t, err)
	require.Equal(t, "8OTC92xYkW7CWPJGhRvqCR0U1CR6L8PhhpRGGxgW4Ts=", id0)

	id1, err := BstatesParseSchemaIdFromType("application/vnd.nayar.bstates; id=\"8OTC92xYkW7CWPJGhRvqCR0U1CR6L8PhhpRGGxgW4Ts=\"")
	require.NoError(t, err)
	require.Equal(t, id0, id1)

	_, err = BstatesParseSchemaIdFromType("application/vnd.nayar.bstates; id=\"\"")
	require.Error(t, err)

	_, err = BstatesParseSchemaIdFromType("application/vnd.nayar.bstates; id=\"8OTC92xYkW7CWPJGhRvqCR0U1CR6L8PhhpRGGxgW4Ts=")
	require.Error(t, err)
}
