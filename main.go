package main

import (
	"bufio"
	//"fmt"
	"log"
	"net"
)

func main() {
	lo, err := net.Listen("tcp", ":6380")
	if err != nil {log.Fatal(err)}

	for {
		conn, err := lo.Accept()
		if err != nil {log.Print(err); continue}
		go rwConnection(conn)
	}

	
}

func rwConnection(conn net.Conn) {
	defer conn.Close()
	bufioWriter := bufio.NewWriter(conn)
	bufioReader := bufio.NewReader(conn)

	for {
		readInputBytes, err := bufioReader.ReadBytes('\n')
		if err != nil {log.Print(err); return}

		_ , err = bufioWriter.Write(readInputBytes)
		if err != nil {log.Print(err); return}
		
		err = bufioWriter.Flush()

		if err != nil {log.Print(err); return}
	}
}