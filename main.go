package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"fmt"
)

func main() {
	kvstore := make(map[string]string)

	lo, err := net.Listen("tcp", ":6380")
	if err != nil {log.Fatal(err)}

	for {
		conn, err := lo.Accept()
		if err != nil {log.Print(err); continue}
		go RESPParse(conn,kvstore)
	}

	
}

func RESPParse(conn net.Conn, kvstore map[string]string) {
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

			commandStore := make([]string, elementQuantity)

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

				
				
				commandStore[i] = string(parsedBuffer)
			}
			
			//fmt.Print(len(commandStore))
			dispatchReturn := runDispatcher(commandStore,kvstore)

			_ , err = bufioWriter.Write([]byte(dispatchReturn))
			if err != nil {log.Print(err); return}
			err = bufioWriter.Flush()
			if err != nil {log.Print(err); return}

		}
	}	
}

func runDispatcher(commands []string, kvstore map[string]string) string {
	if len(commands) == 0 {
		return "-ERR no command ...\r\n"
	}
	
	switch command := commands[0]; command {
	case "SET":
		if len(commands) != 3 {
			return "-ERR wrong number of arguments for 'set' command\r\n"
		}
		kvstore[commands[1]] = commands[2]
		return "+OK\r\n"
	case "PING":
		if len(commands) > 2 {
			return "-ERR wrong number of arguments for 'ping' command\r\n"
		}
		if len(commands) == 2 {
			return "+" + commands[1] + "\r\n"
		}
		return "+PONG\r\n"
	case "GET":
		if len(commands) != 2 {
			return "-ERR wrong number of arguments for 'get' command\r\n"
		}
		value, exists := kvstore[commands[1]]
		if !exists {
			return "$-1\r\n"
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)

	case "DEL":
		if len(commands) != 2 {
			return "-ERR wrong number of arguments for 'del' command\r\n"
		}
		value := kvstore[commands[1]]
		if value == "" {
			return ":0\r\n"
		}
		delete(kvstore,commands[1])
		return ":1\r\n"

	default:
		return "-ERR unknown command " + strings.ToLower(commands[0]) +"\r\n"
	}
	
}
