package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"encoding/gob"
	"os"
	"os/signal"
	"syscall"
)

type KVstore struct {
	mutex sync.RWMutex
	kvmap map[string]valueWithExpiry
}

type valueWithExpiry struct {
	Expiry time.Time
	KvString string
}

func main() {
	r := KVstore{kvmap:make(map[string]valueWithExpiry)}

	lo, err := net.Listen("tcp", ":6380")
	if err != nil {log.Fatal(err)}
	defer lo.Close()
	r.loadMap()
	go r.intervalSnapshot()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			conn, err := lo.Accept()
			if err != nil {log.Print(err); continue}
			go r.RESPParse(conn)
			
		}
	}()
	<-signalChannel
	r.saveMap()
	fmt.Println("Exiting Program")
}



func (r *KVstore) saveMap() {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	KVfile, err := os.Create("Map")
	if err != nil {log.Print(err); return}
	defer KVfile.Close()
	encoder := gob.NewEncoder(KVfile)
	err = encoder.Encode(r.kvmap)
	if err != nil {log.Print(err); return}
}

func (r *KVstore) loadMap() {
	KVfile, err := os.Open("Map")
	if os.IsNotExist(err) {
		return
	}
	if err != nil {log.Print(err); return}
	defer KVfile.Close()

	decoder := gob.NewDecoder(KVfile)
	err = decoder.Decode(&r.kvmap)
	if err != nil {log.Print(err); return}

}

func (r *KVstore) intervalSnapshot() {
	const tickSpeed = 10
	ticker := time.NewTicker(time.Duration(tickSpeed) * time.Second)
	for range ticker.C {
		r.saveMap()
	}
}

func writeResponse(writer *bufio.Writer, response string) error {
	_,err := writer.Write([]byte(response))
	if err != nil {return err}
	return writer.Flush()
}

func (r *KVstore) RESPParse(conn net.Conn) {
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

			if elementQuantity < 0 {
				if err = writeResponse(bufioWriter,"-ERR invalid array length\r\n"); err != nil {
					log.Print(err)
				}
				return
			}
			commandStore := make([]string, elementQuantity)

			for i := 0; i < elementQuantity; i++ {

				readInputBytes, err := bufioReader.ReadBytes('\n')
				if err != nil {log.Print(err); return}

				readStringConversion := string(readInputBytes)
				
				if !strings.HasPrefix(readStringConversion,"$") {
					if err = writeResponse(bufioWriter,"-ERR expected bulk string\r\n"); err != nil {
						log.Print(err)
					}
					return
			}

				trimmedString = strings.TrimPrefix(readStringConversion, "$")
				trimmedString = strings.TrimSuffix(trimmedString, "\r\n")

				byteQuantity, err := strconv.Atoi(trimmedString)
				if err != nil {log.Print(err); return}
				
				const maxBulkStringSize = 512 * 1024 * 1024
				if byteQuantity < 0 || byteQuantity > maxBulkStringSize {
					if err = writeResponse(bufioWriter,"-ERR invalid bulk string length\r\n"); err != nil {
					log.Print(err)
					}
				return
				}
			

				parsedBuffer := make([]byte, byteQuantity)

				_, err = io.ReadFull(bufioReader, parsedBuffer)
				if err != nil {log.Print(err); return}
				
				_, err = bufioReader.Discard(2)
				if err != nil {log.Print(err); return}

				
				
				commandStore[i] = string(parsedBuffer)
			}
			
			dispatchReturn := r.runDispatcher(commandStore)

			err = writeResponse(bufioWriter, dispatchReturn)
			if err != nil {log.Print(err); return}

		} else {
			err = writeResponse(bufioWriter, "-ERR invalid bulk string syntax\r\n")
		 	if err != nil {log.Print(err)}
			return
		}
 
		
		 
	}	
}

func (r *KVstore) getLive(key string) (string, bool) {
	r.mutex.RLock()
	

	value, exists := r.kvmap[key]
	if !exists {
		r.mutex.RUnlock()
		return "", false
	}

	if value.Expiry.IsZero() || !time.Now().After(value.Expiry) {
		r.mutex.RUnlock()
		return value.KvString, true
	}

	r.mutex.RUnlock()
	r.mutex.Lock()
	defer r.mutex.Unlock()

	value, exists = r.kvmap[key]
	if !exists {
		return "", false
	}

	if value.Expiry.IsZero() || !time.Now().After(value.Expiry) {
		return value.KvString, true
	}

	delete(r.kvmap, key)
	return "", false

}

func (r *KVstore) runDispatcher(commands []string) string {
	
	
	
	if len(commands) == 0 {
		return "-ERR no command ...\r\n"
	}
	
	switch command := commands[0]; command {
	case "SET":
		if len(commands) != 3 {
			return "-ERR wrong number of arguments for 'set' command\r\n"
		}
		
		r.mutex.Lock()
		defer r.mutex.Unlock()

		r.kvmap[commands[1]] = valueWithExpiry{KvString: commands[2]}
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

		value, exists := r.getLive(commands[1])
		if !exists {
			return "$-1\r\n"
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)

	case "DEL":
		if len(commands) != 2 {
			return "-ERR wrong number of arguments for 'del' command\r\n"
		}

		r.mutex.Lock()
		defer r.mutex.Unlock()

		_, exists := r.kvmap[commands[1]]
		if !exists {
			return ":0\r\n"
		}
		delete(r.kvmap,commands[1])
		return ":1\r\n"

	case "EXPIRE":
		if len(commands) != 3 {
			return "-ERR wrong number of arguments for 'expire' command\r\n"
		}

		r.mutex.Lock()
		defer r.mutex.Unlock()

		value, exists := r.kvmap[commands[1]]
		if !exists {
			return ":0\r\n"
		}

		previousValue := value.KvString
		
		expiryTime, err := strconv.Atoi(commands[2])
		if err != nil {return "-ERR value is not an integer or out of range\r\n"}

		r.kvmap[commands[1]] = valueWithExpiry{KvString: previousValue, Expiry: time.Now().Add(time.Duration(expiryTime) * time.Second)}
		return ":1\r\n"

	default:
		return "-ERR unknown command " + strings.ToLower(commands[0]) +"\r\n"
	}
	
}
