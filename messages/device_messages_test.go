package messages

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_SysInfoResMsg(t *testing.T) {
	input := SysInfoResMsg{
		SysInfo: SysInfo{
			DeviceInfo: DeviceInfo{
				Address: "random-address",
			},
			LastRunExitIssuedAt: time.UnixMilli(time.Now().UnixMilli()),
			Uptime:              time.Second * 123,
			LastRunUptime:       time.Second * 12345,
		},
		Report: map[string]map[string]interface{}{
			"module1": {
				"prop1": "prop1-value",
			},
		},
	}
	inputRaw, err := ToMsi(input)
	require.NoError(t, err)

	output := SysInfoResMsg{}
	err = ParseMsi(inputRaw, &output)
	require.NoError(t, err)

	require.Equal(t, input, output)
}
