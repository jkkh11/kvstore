package main

import (
	"bufio"
	//"fmt"
	"log"
	"net"
)

func main() {
	lo, err := net.Listen("tcp", ":6380")
	if err != nil {
		log.Fatal(err)

	}

	conn, err := lo.Accept()
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	bufioReader := bufio.NewReader(conn)
	readInputBytes, err := bufioReader.ReadBytes('\n')
	if err != nil {
		log.Fatal(err)
	}

	bufioWriter := bufio.NewWriter(conn)
	bufioWriter.Write(readInputBytes)
	err = bufioWriter.Flush()

	if err != nil {
		log.Fatal(err)
	}
}
