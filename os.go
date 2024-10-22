package idefixgo

import (
	"encoding/base64"
	"os"
	"time"

	m "github.com/nayarsystems/idefix-go/messages"
)

// The following constants define command topics for inter-system communication regarding file operations
// and other commands in the operating system context. Each command topic is a string that specifies
// the type of command being executed, allowing for structured messaging and organization.
//
// TopicCmd serves as the base topic for OS commands, while the other constants extend this base to
// specify particular file operations, such as reading, writing, copying, and managing files. This
// organization helps in routing messages correctly within the system.
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

// FileWrite uploads data to a file at the specified path on the remote system. It sends a request
// using the client's Call2 method, including the file's path, content, and mode for the file's
// permissions. Upon successful upload, the function returns the base64-encoded SHA256 hash of
// the data or an error if the write operation fails.
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

// FileRead retrieves the contents of a file located at the specified path on the remote system.
// It sends a request using the client's Call2 method to read the file data. The function returns
// the file's content as a byte slice or an error if the read operation fails.
func FileRead(ic *Client, address, path string, tout time.Duration) (data []byte, err error) {
	msg := &m.FileReadMsg{
		Path: path,
	}
	resp := &m.FileReadResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdFileRead, Data: msg}, resp, tout)
	data = resp.Data
	return
}

// FileSHA256 computes the SHA256 hash of a file located at the specified path on the remote system.
// It sends a request using the client's Call2 method to retrieve the hash. The function returns the
// computed hash as a byte slice or an error if the request fails.
func FileSHA256(ic *Client, address, path string, tout time.Duration) (hash []byte, err error) {
	msg := &m.FileSHA256Msg{
		Path: path,
	}
	resp := &m.FileSHA256ResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdFileSHA256, Data: msg}, resp, tout)
	hash = resp.Hash
	return
}

// FileSHA256b64 computes the SHA256 hash of a file located at the specified path on the remote system
// and returns it as a base64-encoded string. It first calls [FileSHA256] to obtain the raw hash bytes.
// If successful, the raw hash is then encoded into a base64 string. The function returns the base64
// string representation of the hash or an error if the hash computation fails.
func FileSHA256b64(ic *Client, address, path string, tout time.Duration) (hash string, err error) {
	hashRaw, err := FileSHA256(ic, address, path, tout)
	if err != nil {
		return "", err
	}
	hash = base64.StdEncoding.EncodeToString(hashRaw)
	return
}

// FileCopy sends a request to copy a file from srcPath to dstPath on the remote system.
// It uses the client's Call2 method to execute the file copy operation at the specified address.
// The function returns an error if the file copy operation fails.
func FileCopy(ic *Client, address, srcPath, dstPath string, tout time.Duration) (err error) {
	msg := m.FileCopyMsg{
		SrcPath: srcPath,
		DstPath: dstPath,
	}
	resp := &m.FileCopyResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdFileCopy, Data: msg}, resp, tout)
	return
}

// Remove sends a request to delete a file or directory at the specified path on the remote system.
// It uses the client's Call2 method to perform the removal operation at the given address.
// The function returns an error if the removal operation fails.
func Remove(ic *Client, address, path string, tout time.Duration) (err error) {
	msg := m.RemoveMsg{
		Path: path,
	}
	resp := &m.RemoveResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdRemove, Data: msg}, resp, tout)
	return
}

// Move sends a request to move a file or directory from srcPath to dstPath on the remote system.
// It uses the client's Call2 method to perform the move operation at the specified address.
// The function returns an error if the move operation fails.
func Move(ic *Client, address, srcPath, dstPath string, tout time.Duration) (err error) {
	msg := m.MoveMsg{
		SrcPath: srcPath,
		DstPath: dstPath,
	}
	resp := &m.MoveResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdMove, Data: msg}, resp, tout)
	return
}

// GetFree requests the available free space for a given path from the remote address.
// The function sends a request using the client's Call2 method, asking for the free disk space at the specified path.
// It returns the amount of free space in bytes or an error if the request fails.
func GetFree(ic *Client, address, path string, tout time.Duration) (freeSpace uint64, err error) {
	msg := m.FreeSpaceMsg{
		Path: path,
	}
	resp := &m.FreeSpaceResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdFree, Data: msg}, resp, tout)
	freeSpace = resp.Free
	return
}

// ListDir sends a request to the specified address to list the contents of the directory at the given path.
// It uses the client's Call2 method to send the request and retrieve a response containing the file information.
// The function returns a slice of FileInfo pointers representing the files in the directory or an error if
// the request fails.
func ListDir(ic *Client, address, path string, tout time.Duration) (files []*m.FileInfo, err error) {
	msg := m.ListDirMsg{
		Path: path,
	}
	resp := &m.ListDirResMsg{}
	err = ic.Call2(address, &m.Message{To: TopicCmdListDir, Data: msg}, resp, tout)
	files = resp.Files
	return
}

// ExitToUpdate sends an update command to the given address to initiate a graceful exit and update process.
// The update command includes the update type, cause, stop delay, and wait halt delay, allowing for
// customizable control over the update behavior. It sends the command via the client's Call2 method and
// waits for the response.
//
// The updateType defines the type of update being performed, and cause provides a reason for the update.
// stopDelay and waitHaltDelay specify the durations to wait before stopping and halting operations, respectively.
// tout sets the timeout for the update request.
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
