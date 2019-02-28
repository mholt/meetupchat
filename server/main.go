package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/mholt/meetupchat"
)

func main() {
	ln, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// TODO: POTENTIAL BUG: A defer close here would never be run
		go func(conn net.Conn) {
			// TODO: POTENTIAL BUG: conn should be passed in
			err = handle(conn)
			if err != nil {
				log.Printf("[ERROR] %v", err)
			}
		}(conn)
	}
}

func handle(conn net.Conn) error {
	defer conn.Close()

	// TODO: discuss choice of gob encoding

	dec := gob.NewDecoder(conn)

	var hdr connHeader
	err := dec.Decode(&hdr.Name)
	if err != nil {
		// TODO: discuss wrapping errors
		return fmt.Errorf("decoding header: %v", err)
	}

	// TODO: could inspect header and reject connection

	hdr.encoder = gob.NewEncoder(conn)
	hdr.decoder = dec

	connsMu.Lock()
	conns[conn] = hdr
	connsMu.Unlock()

	log.Printf("[CONNECTED] %s", hdr.Name)

	for {
		var msg meetupchat.Message
		err := hdr.decoder.Decode(&msg)
		if err != nil {
			break
		}

		// in case rogue or buggy clients want to impersonate someone else
		msg.From = hdr.Name

		log.Printf("%s: %s", msg.From, msg.Body)

		connsMu.RLock()
		for c, hdr := range conns {
			if c == conn {
				continue
			}
			go hdr.encoder.Encode(msg)
		}
		connsMu.RUnlock()
	}

	connsMu.Lock()
	delete(conns, conn)
	connsMu.Unlock()

	log.Printf("[DISCONNECTED] %s", hdr.Name)

	return nil
}

type connHeader struct {
	Name string

	decoder *gob.Decoder
	encoder *gob.Encoder
}

var conns = make(map[net.Conn]connHeader) // TODO: POTENTIAL BUG: not initializing map
var connsMu sync.RWMutex
