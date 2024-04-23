package idefixgo

import (
	"encoding/base64"
	"os"
	"time"

	m "github.com/nayarsystems/idefix-go/messages"
)

const (
	TopicCmd = "os.cmd"

	TopicCmdFile       = TopicCmd + ".file"
	TopicCmdFileRead   = TopicCmdFile + ".read"
	TopicCmdFileWrite  = TopicCmdFile + ".write"
	TopicCmdFileSize   = TopicCmdFile + ".size"
	TopicCmdFileCopy   = TopicCmdFile + ".copy"
	TopicCmdFileSHA256 = TopicCmdFile + ".sha256"
	TopicCmdExec       = TopicCmd + ".exec"
	TopicCmdMkdir      = TopicCmd + ".mkdir"
	TopicCmdPatch      = TopicCmd + ".patch"
	TopicCmdRemove     = TopicCmd + ".remove"
	TopicCmdMove       = TopicCmd + ".move"
	TopicCmdFree       = TopicCmd + ".free"
	TopicCmdListDir    = TopicCmd + ".listdir"
)

func FileWrite(ic *Client, address, path string, data []byte, mode os.FileMode, tout time.Duration) (hash string, err error) {
	msg := &m.FileWriteMsg{
		Path: path,
		Data: data,
		Mode: uint32(mode),
	}
	resp := &m.FileWriteResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdFileWrite, Data: msg}, resp, tout)
	hash = base64.StdEncoding.EncodeToString(resp.Hash)
	return
}

func FileRead(ic *Client, address, path string, tout time.Duration) (data []byte, err error) {
	msg := &m.FileReadMsg{
		Path: path,
	}
	resp := &m.FileReadResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdFileRead, Data: msg}, resp, tout)
	data = resp.Data
	return
}

func FileSHA256(ic *Client, address, path string, tout time.Duration) (hash []byte, err error) {
	msg := &m.FileSHA256Msg{
		Path: path,
	}
	resp := &m.FileSHA256ResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdFileSHA256, Data: msg}, resp, tout)
	hash = resp.Hash
	return
}

func FileSHA256b64(ic *Client, address, path string, tout time.Duration) (hash string, err error) {
	hashRaw, err := FileSHA256(ic, address, path, tout)
	if err != nil {
		return "", err
	}
	hash = base64.StdEncoding.EncodeToString(hashRaw)
	return
}

func FileCopy(ic *Client, address, srcPath, dstPath string, tout time.Duration) (err error) {
	msg := m.FileCopyMsg{
		SrcPath: srcPath,
		DstPath: dstPath,
	}
	resp := &m.FileCopyResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdFileCopy, Data: msg}, resp, tout)
	return
}

func Remove(ic *Client, address, path string, tout time.Duration) (err error) {
	msg := m.RemoveMsg{
		Path: path,
	}
	resp := &m.RemoveResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdRemove, Data: msg}, resp, tout)
	return
}

func Move(ic *Client, address, srcPath, dstPath string, tout time.Duration) (err error) {
	msg := m.MoveMsg{
		SrcPath: srcPath,
		DstPath: dstPath,
	}
	resp := &m.MoveResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdMove, Data: msg}, resp, tout)
	return
}

func GetFree(ic *Client, address, path string, tout time.Duration) (freeSpace uint64, err error) {
	msg := m.FreeSpaceMsg{
		Path: path,
	}
	resp := &m.FreeSpaceResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdFree, Data: msg}, resp, tout)
	freeSpace = resp.Free
	return
}

func ListDir(ic *Client, address, path string, tout time.Duration) (files []*m.FileInfo, err error) {
	msg := m.ListDirMsg{
		Path: path,
	}
	resp := &m.ListDirResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdListDir, Data: msg}, resp, tout)
	files = resp.Files
	return
}

func ExitToUpdate(ic *Client, address string, updateType int, cause string, stopDelay, waitHaltDelay, tout time.Duration) (resp *m.UpdateResMsg, err error) {
	msg := m.UpdateMsg{
		Type:          updateType,
		Cause:         cause,
		StopDelay:     stopDelay,
		WaitHaltDelay: waitHaltDelay,
	}
	resp = &m.UpdateResMsg{}
	err = ic.Call2(address, &m.Message{To: "updater.cmd.update", Data: msg}, resp, tout)
	return
}
