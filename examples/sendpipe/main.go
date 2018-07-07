package main

import (
	"os"
	"net"
	"sync"
	"fmt"
	"io/ioutil"

	"github.com/shawhull/gling"
)

//var sockFilename = "@actord"	// allso possible
var sockFilename = "/tmp/gling_socket"

func main() {
	// create unnamed pipe
	f, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	// clean up any left-over file-based unix socket
	os.Remove(sockFilename)
	// liston on given socket filename
	listener, err := net.Listen("unix", sockFilename)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	// set up waitgroup = do not exit immediately
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go getFileDescriptor(&waitGroup)
	// accept first client connection
	var conn net.Conn
	conn, err = listener.Accept()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// send file descriptor to client
	listenConn := conn.(*net.UnixConn)
	if err = gling.SendFileDescriptor(listenConn, f); err != nil {
		panic(err)
	}
	// send data through the pipe to the client
	w.Write([]byte("Hello pipe!"))
	w.Close()
	// wait for getFileDescriptor() to finish
	waitGroup.Wait()
}

// getFileDescriptor connects to the unix socket server, receives a file descriptors and drains it.
func getFileDescriptor(waitGroup *sync.WaitGroup) {
	// send done signal to waitgroup on function return
	defer waitGroup.Done()
	// connect to unix socket server
	c, err := net.Dial("unix", sockFilename)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	// convert to UnixConn
	sendFdConn := c.(*net.UnixConn)
	// receive a file description with some name
	var files []*os.File
	files, err = gling.ReceiveFileDescriptor(sendFdConn, 1, []string{"sentFile"})
	if err != nil {
		panic(err)
	}
	file := files[0]
	defer file.Close()
	// read all from that file, which is a unnamed pipe
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	// print contents
	fmt.Println("read", len(bytes), "bytes:", string(bytes))
}
