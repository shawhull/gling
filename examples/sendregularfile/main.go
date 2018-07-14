package main

import (
	"os"
	"net"
	"sync"
	"fmt"

	"github.com/shawhull/gling"
)

var sockFilename = "/tmp/gling_socket"

func main() {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	os.Remove(sockFilename)
	listener, err := net.Listen("unix", sockFilename)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go getFileDescriptor(&waitGroup)
	var conn net.Conn
	conn, err = listener.Accept()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	listenConn := conn.(*net.UnixConn)
	if err = gling.SendFileDescriptor(listenConn, f); err != nil {
		panic(err)
	}
	waitGroup.Wait()
}

func getFileDescriptor(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	c, err := net.Dial("unix", sockFilename)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	sendFdConn := c.(*net.UnixConn)

	var files []*os.File
	files, err = gling.ReceiveFileDescriptor(sendFdConn, 1, []string{"sentFile"})
	if err != nil {
		panic(err)
	}
	file := files[0]
	defer file.Close()
	bytes := make([]byte, 64)
	var n int
	n, err = file.Read(bytes)
	if err != nil {
		panic(err)
	}
	if n < 1 {
		panic("failed to read the data")
	}
	fmt.Println(bytes)
	fmt.Println(n)
}
