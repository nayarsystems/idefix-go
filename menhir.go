package idefixgo

import (
	"fmt"
	"time"

	"github.com/jaracil/ei"
	m "github.com/nayarsystems/idefix-go/messages"
)

// These constants define topic suffixes used for sending various commands trough Idefix to Menhir.
// Each constant represents a specific command or action.
//
// SuffixMenhirCmd represents the base suffix for general commands,
// while the others specify commands related to storage operations.
const (
	SuffixMenhirCmd              = "cmd"
	SuffixMenhirCmdFrame         = SuffixMenhirCmd + ".frame"
	SuffixMenhirCmdStorageWrite  = SuffixMenhirCmd + ".storage.write"
	SuffixMenhirCmdStorageRead   = SuffixMenhirCmd + ".storage.read"
	SuffixMenhirCmdStorageRemove = SuffixMenhirCmd + ".storage.remove"
	SuffixMenhirCmdStorageStat   = SuffixMenhirCmd + ".storage.stat"
	SuffixMenhirCmdStorageStats  = SuffixMenhirCmd + ".storage.stats"
)

// MenhirStorageFileInfo represents metadata information about a file
// stored in the Menhir storage system.
type MenhirStorageFileInfo struct {
	Filename string
	Type     uint8
	Size     uint32
}

// These constants represent file types used in the Menhir storage system.
// Each constant categorizes a file as either a regular file or a directory.
const (
	MenhirStorageFileTypeRegular   = 1
	MenhirStorageFileTypeDirectory = 2
)

// MenhirStorageFSStats contains statistics about the file system
// in the Menhir storage system, related to block storage usage.
type MenhirStorageFSStats struct {
	BlockSize  uint32
	BlockCount uint32
	BlocksUsed uint32
}

// MenhirStorageWrite uploads a file to the Menhir storage system.
//
// It sends the provided data to the specified Menhir instance to be stored
// under the given filename. The function constructs a message containing the
// filename and data, then calls the remote system to execute the storage write command.
func MenhirStorageWrite(ic *Client, address, menhirInstance, fname string, data []byte, tout time.Duration) (err error) {
	msg := map[string]any{
		"filename": fname,
		"data":     data,
	}
	_, err = ic.Call(address, &m.Message{To: fmt.Sprintf("%s.%s", menhirInstance, SuffixMenhirCmdStorageWrite), Data: msg}, tout)
	return
}

// MenhirStorageRead retrieves a file's data from the Menhir storage system.
//
// It sends a request to the specified Menhir instance to read the file with the given filename.
// The function returns the file's data or an error if the read operation fails.
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

// MenhirStorageRemove removes a file from the Menhir storage system.
//
// The function sends a command to the specified Menhir instance to delete a file identified by its filename.
// If the operation is successful, the file is removed from storage.
func MenhirStorageRemove(ic *Client, address, menhirInstance, fname string, tout time.Duration) (err error) {
	msg := map[string]any{
		"filename": fname,
	}
	_, err = ic.Call(address, &m.Message{To: fmt.Sprintf("%s.%s", menhirInstance, SuffixMenhirCmdStorageRemove), Data: msg}, tout)
	return
}

// MenhirStorageStat retrieves the metadata of a file from the Menhir storage system.
//
// The function requests file information, such as filename, type, and size, from the specified Menhir instance.
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

// MenhirStorageStats retrieves filesystem statistics from the Menhir storage system.
//
// The function fetches details such as the block size, total block count, and the number of blocks used from the specified Menhir instance.
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
