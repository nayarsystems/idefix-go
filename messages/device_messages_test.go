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

	require.Contains(t, inputRaw, "uptime")
	require.Equal(t, int64(123), inputRaw["uptime"])

	require.Contains(t, inputRaw, "lastRunUptime")
	require.Equal(t, int64(12345), inputRaw["lastRunUptime"])

	require.Contains(t, inputRaw, "devInfo")
	devInfoRaw := inputRaw["devInfo"]
	require.IsType(t, map[string]any{}, devInfoRaw)
	require.Contains(t, devInfoRaw, "address")

	output := SysInfoResMsg{}
	err = ParseMsi(inputRaw, &output)
	require.NoError(t, err)

	require.Equal(t, input, output)
}
