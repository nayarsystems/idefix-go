package messages

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

type mapstructureHooksTestStruct struct {
	Date    time.Time     `mapstructure:"date"`
	Uptime  time.Duration `mapstructure:"uptime"`
	RawData []byte        `mapstructure:"rawData"`
}

func Test_MapstructureDecodeHooks(t *testing.T) {
	tt := time.UnixMilli(time.Now().UnixMilli())
	dd := time.Second * 123
	rawData := []byte("hello world")
	rawDataB64 := base64.StdEncoding.EncodeToString(rawData)
	input := map[string]any{
		"date":    tt.UnixMilli(),
		"uptime":  dd.Seconds(),
		"rawData": rawDataB64,
	}
	out := mapstructureHooksTestStruct{}
	err := ParseMsiGeneric(input, &out,
		mapstructure.ComposeDecodeHookFunc(
			// DecodeBase64ToSliceHookFunc() hook is always added
			DecodeNumberToDurationHookFunc(time.Second),
			DecodeUnixMilliToTimeHookFunc()))
	require.NoError(t, err)
	require.Equal(t, tt, out.Date)
	require.Equal(t, dd, out.Uptime)
	require.Equal(t, rawData, out.RawData)
}

func Test_MapstructureEncodeHooks(t *testing.T) {
	tt := time.UnixMilli(time.Now().UnixMilli())
	dd := time.Second * 123
	rawData := []byte("hello world")
	rawDataB64 := base64.StdEncoding.EncodeToString(rawData)
	eOutput := map[string]any{
		"date":    tt.UnixMilli(),
		"uptime":  int64(dd.Seconds()),
		"rawData": rawDataB64,
	}
	input := mapstructureHooksTestStruct{
		Date:    tt,
		Uptime:  dd,
		RawData: rawData,
	}
	output, err := ToMsiGeneric(input,
		mapstructure.ComposeEncodeFieldMapHookFunc(
			// EncodeSliceToBase64Hook() hook is always added
			EncodeDurationToSecondsInt64Hook(),
			EncodeTimeToUnixMilliHook()))
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

// type WithTimeAndDurations struct {
// 	Since   time.Time     `json:"since" msgpack:"since" mapstructure:"since,omitempty"`
// 	Timeout time.Duration `json:"timeout" msgpack:"timeout" mapstructure:"timeout,omitempty"`
// }

type WithByteSlice struct {
	Buffer []byte `mapstructure:"buffer"`
}

func Test_ParseByteSlice_ParseMsi(t *testing.T) {
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
