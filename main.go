package main

import (
	"fmt"
	"net"
)

func main() {
	lo, err := net.Listen("tcp",":6380")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(lo,"test")
}