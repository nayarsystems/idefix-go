package messages

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_EventsGetResponseMsg(t *testing.T) {
	t0, err := time.Parse(time.RFC3339, TimeToString(time.Now()))
	require.NoError(t, err)
	t1, err := time.Parse(time.RFC3339, TimeToString(time.Now()))
	require.NoError(t, err)
	input := EventsGetResponseMsg{
		Events: []*Event{
			{
				EventMsg: EventMsg{
					UID:     "uid1",
					Meta:    map[string]interface{}{"meta1": "meta1-value"},
					Type:    "type1",
					Payload: map[string]interface{}{"payload1": "payload1-value"},
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
					Payload: map[string]interface{}{"payload2": "payload2-value"},
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

	output := EventsGetResponseMsg{}
	err = ParseMsi(inputRaw, &output)
	require.NoError(t, err)

	require.Equal(t, input, output)
}
