package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
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
		if err == io.EOF {return}
		if err != nil {log.Print(err); return}

		readStringConversion := string(readInputBytes)

		if strings.HasPrefix(readStringConversion, "*") {
			trimmedString := strings.TrimPrefix(readStringConversion, "*")
			trimmedString = strings.TrimSuffix(trimmedString, "\r\n")

			elementQuantity, err := strconv.Atoi(trimmedString)
			if err != nil {log.Print(err); return}

			for i := 0; i < elementQuantity; i++ {

				readInputBytes, err := bufioReader.ReadBytes('\n')
				if err != nil {log.Print(err); return}

				readStringConversion := string(readInputBytes)
				
				trimmedString = strings.TrimPrefix(readStringConversion, "$")
				trimmedString = strings.TrimSuffix(trimmedString, "\r\n")

				byteQuantity, err := strconv.Atoi(trimmedString)
				if err != nil {log.Print(err); return}
				
				parsedBuffer := make([]byte, byteQuantity)

				_, err = io.ReadFull(bufioReader, parsedBuffer)
				if err != nil {log.Print(err); return}
				
				_, err = bufioReader.Discard(2)
				if err != nil {log.Print(err); return}

				_ , err = bufioWriter.Write(parsedBuffer)
				if err != nil {log.Print(err); return}
				
				err = bufioWriter.Flush()

				if err != nil {log.Print(err); return}
			}
		}
	}	
}
