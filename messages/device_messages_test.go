package messages

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

//	"module1": {
//		"prop1": "prop1-value",
//	},
func Test_SysInfoResMsg(t *testing.T) {
	input := SysInfoResMsg{
		SysInfo: SysInfo{
			SysInfoVersion: 2,
			DeviceInfo: DeviceInfo{
				Address: "random-address",
			},
			LastRunExitIssuedAt: time.UnixMilli(time.Now().UnixMilli()),
			Uptime:              time.Second * 123,
			LastRunUptime:       time.Second * 12345,
		},
		Report: map[string]SysInfoReportMsg{
			"moduleA-1": {
				Class: "moduleA",
				State: map[string]any{
					"prop1": "prop1-value",
				},
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
