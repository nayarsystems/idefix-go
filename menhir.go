package idefixgo

import (
	"fmt"
	"time"

	"github.com/jaracil/ei"
	m "github.com/nayarsystems/idefix-go/messages"
)

const (
	SuffixMenhirCmd              = "cmd"
	SuffixMenhirCmdFrame         = SuffixMenhirCmd + ".frame"
	SuffixMenhirCmdStorageWrite  = SuffixMenhirCmd + ".storage.write"
	SuffixMenhirCmdStorageRead   = SuffixMenhirCmd + ".storage.read"
	SuffixMenhirCmdStorageRemove = SuffixMenhirCmd + ".storage.remove"
	SuffixMenhirCmdStorageStat   = SuffixMenhirCmd + ".storage.stat"
	SuffixMenhirCmdStorageStats  = SuffixMenhirCmd + ".storage.stats"
)

type MenhirStorageFileInfo struct {
	Filename string
	Type     uint8
	Size     uint32
}

const (
	MenhirStorageFileTypeRegular   = 1
	MenhirStorageFileTypeDirectory = 2
)

type MenhirStorageFSStats struct {
	BlockSize  uint32
	BlockCount uint32
	BlocksUsed uint32
}

func MenhirStorageWrite(ic *Client, address, menhirInstance, fname string, data []byte, tout time.Duration) (err error) {
	msg := map[string]any{
		"filename": fname,
		"data":     data,
	}
	_, err = ic.Call(address, &m.Message{To: fmt.Sprintf("%s.%s", menhirInstance, SuffixMenhirCmdStorageWrite), Data: msg}, tout)
	return
}

func MenhirStorageRead(ic *Client, address, menhirInstance, fname string, tout time.Duration) (data []byte, err error) {
	msg := map[string]any{
		"filename": fname,
	}
	res, err := ic.Call(address, &m.Message{To: fmt.Sprintf("%s.%s", menhirInstance, SuffixMenhirCmdStorageRead), Data: msg}, tout)
	if err != nil {
		return
	}
	data, err = ei.N(res.Data).M("data").Bytes()
	return
}

func MenhirStorageRemove(ic *Client, address, menhirInstance, fname string, tout time.Duration) (err error) {
	msg := map[string]any{
		"filename": fname,
	}
	_, err = ic.Call(address, &m.Message{To: fmt.Sprintf("%s.%s", menhirInstance, SuffixMenhirCmdStorageRemove), Data: msg}, tout)
	return
}

func MenhirStorageStat(ic *Client, address, menhirInstance, fname string, tout time.Duration) (finfo MenhirStorageFileInfo, err error) {
	msg := map[string]any{
		"filename": fname,
	}
	res, err := ic.Call(address, &m.Message{To: fmt.Sprintf("%s.%s", menhirInstance, SuffixMenhirCmdStorageStat), Data: msg}, tout)
	if err != nil {
		return
	}
	finfo.Filename = ei.N(res.Data).M("filename").StringZ()
	finfo.Type = ei.N(res.Data).M("type").Uint8Z()
	finfo.Size = ei.N(res.Data).M("size").Uint32Z()
	return
}

func MenhirStorageStats(ic *Client, address, menhirInstance string, tout time.Duration) (stats MenhirStorageFSStats, err error) {
	res, err := ic.Call(address, &m.Message{To: fmt.Sprintf("%s.%s", menhirInstance, SuffixMenhirCmdStorageStats), Data: nil}, tout)
	if err != nil {
		return
	}
	stats.BlockSize = ei.N(res.Data).M("block_size").Uint32Z()
	stats.BlockCount = ei.N(res.Data).M("block_count").Uint32Z()
	stats.BlocksUsed = ei.N(res.Data).M("blocks_used").Uint32Z()
	return
}
