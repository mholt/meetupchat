package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/mholt/meetupchat"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter your name: ")
	scanner.Scan()
	name := scanner.Text()
	if name == "" {
		return
	}

	conn, err := net.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// send connection header
	enc := gob.NewEncoder(conn)
	err = enc.Encode(name)
	if err != nil {
		log.Fatal(err)
	}

	// begin receiving messages (NOTE: potential goroutine leak)
	go recv(conn)

	// loop to send messages
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		msg := scanner.Text()

		if msg == "" {
			continue
		}

		err = enc.Encode(meetupchat.Message{
			From: name, // NOTE: this is optional, because the server already knows who we are
			Body: msg,
			Time: time.Now(),
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func recv(conn net.Conn) {
	dec := gob.NewDecoder(conn)

	for {
		var msg meetupchat.Message
		err := dec.Decode(&msg)
		if err != nil {
			break
		}

		fmt.Printf("\n[%s] %s: %s\n...> ",
			msg.Time.Format("Jan 2, 3:04:05 PM"), msg.From, msg.Body)
	}
}
