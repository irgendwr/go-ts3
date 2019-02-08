package ts3

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

const avatarPrefix = "/avatar_"
const bufferSize = 16384 // 16 kB

// FileArg identifies a file or folder.
type FileArg struct {
	ChannelID int
	Password  string
	Path      string
}

// FileTransfer represents a file up- or download.
type FileTransfer struct {
	server       *ServerMethods
	ClientFTID   int    `ms:"clientftfid"`
	ServerFTID   int    `ms:"serverftfid"`
	Key          string `ms:"ftkey"`
	Port         int    `ms:"port"`
	SeekPosition int    `ms:"seekpos"`
	Size         int    `ms:"size"`
}

// FileTransferStatus contains information about a file up- or download.
type FileTransferStatus struct {
	ClientFTID   int     `ms:"clientftfid"`
	ServerFTID   int     `ms:"serverftfid"`
	ClientID     int     `ms:"clid"`
	Path         string  `ms:"path"`
	Name         string  `ms:"name"`
	Size         int     `ms:"size"`
	SizeDone     int     `ms:"sizeDone"`
	Sender       int     `ms:"sender"`
	Status       int     `ms:"status"`
	CurrentSpeed float64 `ms:"current_speed"`
	AverageSpeed float64 `ms:"average_speed"`
	Runtime      int     `ms:"runtime"` // TODO: use time.Duration
}

// FileInfo represents file metadata.
type FileInfo struct {
	ChannelID int       `ms:"cid"`
	Name      string    `ms:"name"`
	Size      int       `ms:"size"`
	Date      time.Time `ms:"datetime"`
	Type      int       `ms:"type"`
}

// ClientAvatar downloads the avatar of a client identified by the Base64HashClientUID.
func (s *ServerMethods) ClientAvatar(base64HashClientUID string) (interface{}, error) {
	if _, err := s.GetFileInfo(FileArg{Path: avatarPrefix + base64HashClientUID}); err != nil {
		return nil, err
	}
	//ftInitDownload(avatarPrefix + $auid, 0, '')
	//ftDownloadFile($init)
	return nil, errors.New("not implemented")
}

// InitUpload initializes a file transfer upload.
// clientFileTransferID is an arbitrary ID to identify the file transfer on client-side.
// On success, the server generates a new ftkey which is required to start uploading
// the file through TeamSpeak 3's file transfer interface.
func (s *ServerMethods) InitUpload(file FileArg, fileTransferID, size int, overwrite, resume bool) (*FileTransfer, error) {
	var ft *FileTransfer
	if _, err := s.ExecCmd(NewCmd("ftinitupload").WithArgs(
		NewArg("cid", file.ChannelID),
		NewArg("cpw", file.Password),
		NewArg("name", file.Path),
		NewArg("clientftfid", fileTransferID),
		NewArg("size", size),
		NewArg("overwrite", overwrite),
		NewArg("resume", resume),
	).WithResponse(&ft)); err != nil {
		return nil, err
	}
	ft.Size = size
	ft.server = s
	return ft, nil
}

// InitDownload initializes a file transfer download.
// clientFileTransferID is an arbitrary ID to identify the file transfer on client-side.
// On success, the server generates a new ftkey which is required to start downloading
// the file through TeamSpeak 3's file transfer interface.
func (s *ServerMethods) InitDownload(file FileArg, fileTransferID, seekPosition int) (*FileTransfer, error) {
	var ft *FileTransfer
	if _, err := s.ExecCmd(NewCmd("ftinitdownload").WithArgs(
		NewArg("cid", file.ChannelID),
		NewArg("cpw", file.Password),
		NewArg("name", file.Path),
		NewArg("clientftfid", fileTransferID),
		NewArg("seekpos", seekPosition),
	).WithResponse(&ft)); err != nil {
		return nil, err
	}
	ft.SeekPosition = seekPosition
	ft.server = s
	return ft, nil
}

// Stop stops the running file transfer.
func (ft *FileTransfer) Stop(delete bool) error {
	_, err := ft.server.ExecCmd(NewCmd("ftstop").WithArgs(
		NewArg("serverftfid", ft.ServerFTID),
		NewArg("delete", delete),
	))
	return err
}

// Upload uploads a file.
func (ft *FileTransfer) Upload() error {
	// TODO: implement
	return errors.New("not implemented")
}

// Download downloads a file.
func (ft *FileTransfer) Download() (*string, error) {
	/*conn, err := ft.newConnection()
	if err != nil {
		return nil, err
	}

	// TODO: implement
	connbuf := bufio.NewReaderSize(conn, bufferSize)
	for {
		str, err := connbuf.
		if len(str) > 0 {
			fmt.Println(str)
		}
		if err != nil {
			break
		}
	}*/
	return nil, errors.New("not implemented")
}

// newConnection opens a new FileTransfer connection
func (ft *FileTransfer) newConnection() (net.Conn, error) {
	// Get hostname from client connection.
	a := strings.Split(ft.server.conn.RemoteAddr().String(), ":")
	host := strings.Join(a[:len(a)-1], ":")

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, ft.Port), ft.server.timeout)
	if err != nil {
		return nil, err
	}

	if _, err := conn.Write([]byte(ft.Key)); err != nil {
		defer conn.Close()
		return nil, err
	}

	return conn, nil
}

// FileTransferList returns a list of running file transfers on the selected virtual server.
func (s *ServerMethods) FileTransferList() (*[]FileTransferStatus, error) {
	var ftList *[]FileTransferStatus
	if _, err := s.ExecCmd(NewCmd("ftlist").WithResponse(&ftList)); err != nil {
		return nil, err
	}
	return ftList, nil
}

// GetFileList displays a list of files and directories stored in the specified channels file repository.
func (s *ServerMethods) GetFileList(file FileArg) (*[]FileInfo, error) {
	var fileInfo *[]FileInfo
	if _, err := s.ExecCmd(NewCmd("ftgetfilelist").WithArgs(
		NewArg("cid", file.ChannelID),
		NewArg("cpw", file.Password),
		NewArg("path", file.Path),
	).WithResponse(&fileInfo)); err != nil {
		return nil, err
	}
	return fileInfo, nil
}

// GetFileInfo returns information about one or more specified files stored in a channels file repository.
func (s *ServerMethods) GetFileInfo(files ...FileArg) (*[]FileInfo, error) {
	fileArgGroups := make([]CmdArg, len(files))
	for i, file := range files {
		fileArgGroups[i] = NewArgGroup(
			NewArg("cid", file.ChannelID),
			NewArg("cpw", file.Password),
			NewArg("name", file.Path),
		)
	}

	var fileInfo *[]FileInfo
	if _, err := s.ExecCmd(NewCmd("ftgetfileinfo").WithArgs(fileArgGroups...).WithResponse(&fileInfo)); err != nil {
		return nil, err
	}
	return fileInfo, nil
}

// DeleteFile deletes one or more files stored in a channels file repository.
func (s *ServerMethods) DeleteFile(files ...FileArg) error {
	fileArgGroups := make([]CmdArg, len(files))
	for i, file := range files {
		fileArgGroups[i] = NewArgGroup(
			NewArg("cid", file.ChannelID),
			NewArg("cpw", file.Password),
			NewArg("name", file.Path),
		)
	}

	_, err := s.ExecCmd(NewCmd("ftdeletefile").WithArgs(fileArgGroups...))
	return err
}

// RenameFile renames a file in a channels file repository.
// If new.ChannelID is set, the file will be moved into another channels file repository.
func (s *ServerMethods) RenameFile(old, new FileArg) error {
	args := []CmdArg{
		NewArg("cid", old.ChannelID),
		NewArg("cpw", old.Password),
		NewArg("oldname", old.Path),
		NewArg("newname", new.Path),
	}

	if new.ChannelID != 0 && new.ChannelID != old.ChannelID {
		args = append(args,
			NewArg("tcid", new.ChannelID),
			NewArg("tcpw", new.Password),
		)
	}

	_, err := s.ExecCmd(NewCmd("ftcreatedir").WithArgs(args...))
	return err
}
