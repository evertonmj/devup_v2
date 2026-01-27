//go:build ignore
package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	fmt.Println(l.Addr().(*net.TCPAddr).Port)
	for {
		time.Sleep(time.Hour)
	}
}
