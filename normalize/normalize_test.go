package normalize

import (
	"testing"
	"time"

	"github.com/jaracil/ei"
	"github.com/stretchr/testify/require"
)

func TestDecodeTypes(t *testing.T) {
	d := msi{
		"key0":     10,
		"key1:hex": "eeaa",
		"key2:b64": "EUA=", // 0x1140
		"key3": msi{
			"key3.1:hex":             "7744",
			"key3.2:b64":             "FxE=",                                             // 0x1711
			"key3.3:string:gzip:b64": "H4sIAAAAAAAAA/PIz0lUyC3NS8lXSC4qTc0BALdHfuoQAAAA", // Hola mundo cruel
			"key3.4:bytes":           "Adios mundo cruel",
			"key3.5:dur":             "2s",
		},
	}
	err := DecodeTypes(d)
	require.NoError(t, err)
	require.Len(t, d, 4)
	require.Equal(t, ei.N(d).M("key0").IntZ(), 10)

	v, ok := d["key1"]
	require.True(t, ok)
	require.IsType(t, []byte{}, v)
	require.Equal(t, v.([]byte)[0], uint8(0xee))
	require.Equal(t, v.([]byte)[1], uint8(0xaa))

	v, ok = d["key2"]
	require.True(t, ok)
	require.IsType(t, []byte{}, v)
	require.Equal(t, v.([]byte)[0], uint8(0x11))
	require.Equal(t, v.([]byte)[1], uint8(0x40))

	v, ok = d["key3"]
	require.True(t, ok)

	d, ok = v.(msi)
	require.True(t, ok)
	require.Len(t, d, 5)

	v, ok = d["key3.1"]
	require.True(t, ok)
	require.Equal(t, v.([]byte)[0], uint8(0x77))
	require.Equal(t, v.([]byte)[1], uint8(0x44))

	v, ok = d["key3.2"]
	require.True(t, ok)
	require.Equal(t, v.([]byte)[0], uint8(0x17))
	require.Equal(t, v.([]byte)[1], uint8(0x11))

	v, ok = d["key3.3"]
	require.True(t, ok)
	require.Equal(t, v, "Hola mundo cruel")

	v, ok = d["key3.4"]
	require.True(t, ok)
	require.IsType(t, []byte{}, v)
	require.Equal(t, string(v.([]byte)), "Adios mundo cruel")

	v, ok = d["key3.5"]
	require.True(t, ok)
	require.IsType(t, time.Second, v)
	require.Equal(t, v, time.Second*2)
}

func TestEncodeTypes(t *testing.T) {
	duration := time.Hour + time.Minute + time.Second + time.Millisecond
	d := msi{
		"key0": make([]byte, 128),
		"key1": make([]byte, 63),
		"key2": time.Now(),
		"key3": duration,
	}

	opts := &EncodeTypesOpts{BytesToB64: true, Compress: true, CompThreshold: 64}
	err := EncodeTypes(d, opts)
	require.NoError(t, err)
	require.Len(t, d, 4)

	v, ok := d["key0:gzip:b64"]
	require.True(t, ok)
	require.IsType(t, v, "")

	v, ok = d["key1:b64"]
	require.True(t, ok)
	require.IsType(t, v, "")

	v, ok = d["key2:time"]
	require.True(t, ok)
	require.IsType(t, v, int64(0))

	v, ok = d["key3:dur"]
	require.True(t, ok)
	require.Equal(t, duration.String(), v)

	// Decode previous encoded types
	err = DecodeTypes(d)
	require.NoError(t, err)
	require.Len(t, d, 4)

	v, ok = d["key0"]
	require.True(t, ok)
	require.IsType(t, v, []byte{})
	require.Len(t, v, 128)

	v, ok = d["key1"]
	require.True(t, ok)
	require.IsType(t, v, []byte{})
	require.Len(t, v, 63)

	v, ok = d["key2"]
	require.True(t, ok)
	require.IsType(t, v, time.Time{})

	v, ok = d["key3"]
	require.True(t, ok)
	require.Equal(t, duration, v)
}

func normalizeOrDie(t *testing.T, d msi) {
	err := DecodeTypes(d)
	require.Error(t, err)
}

func TestNestedInvalid(t *testing.T) {
	normalizeOrDie(t, msi{
		"key3": msi{
			"key3.1:hex": "77g44",
		},
	})

	normalizeOrDie(t, msi{
		"key3": msi{
			"key3.1:b64": "77g44",
		},
	})

	normalizeOrDie(t, msi{
		"key3": msi{
			"key3.1:bytes": 33,
		},
	})

	normalizeOrDie(t, msi{
		"key3": msi{
			"key3.1:bytes:string": 33,
		},
	})

	normalizeOrDie(t, msi{
		"key3": msi{
			"key3.1:gzip": 33,
		},
	})

	normalizeOrDie(t, msi{
		"key3": msi{
			"key3.1:aaaaaaaaa": 33,
		},
	})
}
