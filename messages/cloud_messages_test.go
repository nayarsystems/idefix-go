package messages

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_EventMsgMsg(t *testing.T) {
	input := EventMsg{
		UID:     "uid1",
		Meta:    map[string]interface{}{"meta1": "meta1-value"},
		Type:    "type1",
		Payload: "test payload",
	}
	inputRaw, err := ToMsi(input)
	require.NoError(t, err)

	output := EventMsg{}
	err = ParseMsi(inputRaw, &output)
	require.NoError(t, err)

	require.Equal(t, input, output)
}

func Test_EventsGetResponseMsg(t *testing.T) {
	t0, err := time.Parse(time.RFC3339, TimeToString(time.Now()))
	require.NoError(t, err)
	t1, err := time.Parse(time.RFC3339, TimeToString(time.Now()))
	require.NoError(t, err)
	payloadB64 := base64.StdEncoding.EncodeToString([]byte{1, 2, 3, 4})
	input := EventsGetResponseMsg{
		Events: []*Event{
			{
				EventMsg: EventMsg{
					UID:     "uid1",
					Meta:    map[string]interface{}{"meta1": "meta1-value"},
					Type:    "type1",
					Payload: payloadB64,
				},
				Domain:    "domain1",
				Address:   "address1",
				Timestamp: t0,
			},
			{
				EventMsg: EventMsg{
					UID:     "uid2",
					Meta:    map[string]interface{}{"meta2": "meta2-value"},
					Type:    "type2",
					Payload: payloadB64,
				},
				Domain:    "domain2",
				Address:   "address2",
				Timestamp: t1,
			},
		},
		ContinuationID: "continuation-id",
	}
	inputRaw, err := ToMsi(input)
	require.NoError(t, err)

	require.Contains(t, inputRaw, "events")
	events := inputRaw["events"]
	require.IsType(t, []any{}, events)
	require.Len(t, events, 2)
	require.IsType(t, msi{}, events.([]any)[0])

	output := EventsGetResponseMsg{}
	err = ParseMsi(inputRaw, &output)
	require.NoError(t, err)

	require.Equal(t, input, output)
}

func Test_EventsGetMsg(t *testing.T) {
	since, err := time.Parse(time.RFC3339, TimeToString(time.Now()))
	require.NoError(t, err)
	input := EventsGetMsg{
		UID:            "uid1",
		Domain:         "domain1",
		Address:        "address1",
		Since:          since,
		Limit:          10,
		Timeout:        time.Second * 10,
		ContinuationID: "continuation-id",
	}
	inputRaw, err := ToMsi(input)
	require.NoError(t, err)

	output := EventsGetMsg{}
	err = ParseMsi(inputRaw, &output)
	require.NoError(t, err)

	require.Equal(t, input, output)
}

func Test_EventMsgGetUID(t *testing.T) {
	input := EventsGetMsg{
		UID: "uid1",
	}
	inputRaw, err := ToMsi(input)
	fmt.Println(inputRaw)
	require.NoError(t, err)

	output := EventsGetMsg{}
	err = ParseMsi(inputRaw, &output)
	require.NoError(t, err)

	input.Since = time.Unix(0, 0)
	require.Equal(t, input, output)
}
