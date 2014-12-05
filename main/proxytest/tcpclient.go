// +build tcpclient

package main

import (
	c "crypto/rand"
	"flag"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/mdevilliers/redishappy/services/logger"
)

var server string

func init() {
	flag.StringVar(&server, "server", "127.0.0.1:9999", "Server address")
}

func main() {

	flag.Parse()
	rand.Seed(time.Now().Unix())
	clients := rand.Intn(64)

	var wg sync.WaitGroup
	wg.Add(clients)

	for i := 0; i < clients; i++ {
		go NewClient(server)
	}

	wg.Wait()
}

func rand_str(size int) []byte {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, size)
	c.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return bytes
}

func NewClient(server string) {

	interval := rand.Intn(1000)
	message := rand_str(rand.Intn(1000))

	logger.Info.Printf("Client Info - Interval : %d, Message : %s, MessageSize : %d", interval, string(message), len(message))

	tcpAddr, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		logger.Info.Printf("ResolveTCPAddr failed: %s", err.Error())
		wg.Done()
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		logger.Info.Printf("DialTCP failed: %s", err.Error())
		wg.Done()
		return
	}

	go func() {

		reply := make([]byte, len(message))

		for {
			_, err = conn.Write(message)
			if err != nil {
				logger.Info.Printf("Write to server failed:  %s", err.Error())
				wg.Done()
				return
			}

			_, err = conn.Read(reply)
			if err != nil {
				logger.Info.Printf("Read from server failed: %s", err.Error())
				wg.Done()
				return
			}

			if string(message) != string(reply) {
				logger.Info.Print("Read isn't the same as I wrote!")
				logger.Info.Print(string(message))
				logger.Info.Print(string(reply))
				wg.Done()
				return
			}

			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}()
}
