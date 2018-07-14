package gling

import (
	"net"
	"os"
	"syscall"
)

func ReceiveFileDescriptor(conn *net.UnixConn, count int, filenames []string) ([]*os.File, error) {
	if count < 1 {
		return nil, nil
	}
	socketFile, err := conn.File()
	if err != nil {
		return nil, err
	}
	socketFd := int(socketFile.Fd())
	defer socketFile.Close()
	buffers := make([]byte, syscall.CmsgSpace(count*4))
	_, _, _, _, err = syscall.Recvmsg(socketFd, nil, buffers, 0)
	if err != nil {
		return nil, err
	}
	var controlMessages []syscall.SocketControlMessage
	controlMessages, err = syscall.ParseSocketControlMessage(buffers)
	files := make([]*os.File, 0, len(controlMessages))
	for i := 0; i < len(controlMessages) && err == nil; i++ {
		var fds []int
		fds, err = syscall.ParseUnixRights(&controlMessages[i])
		for j, fd := range fds {
			var filename string
			if j < len(filenames) {
				filename = filenames[j]
			}

			files = append(files, os.NewFile(uintptr(fd), filename))
		}
	}
	return files, err
}

func SendFileDescriptor(conn *net.UnixConn, files ...*os.File) error {
	if len(files) == 0 {
		return nil
	}
	file, err := conn.File()
	if err != nil {
		return err
	}
	socketFd := int(file.Fd())
	defer file.Close()
	fds := make([]int, len(files))
	for i := range files {
		fds[i] = int(files[i].Fd())
	}
	// Encode file descriptors into a socket control message
	rights := syscall.UnixRights(fds...)
	return syscall.Sendmsg(socketFd, nil, rights, nil, 0)
}
