package main

import (
	"fmt"
	"glover/db"
	"glover/keylog/parser"
	"glover/keylog/ports"
	"glover/server"
	"log"
	"net/http"
	"os"

	"go.bug.st/serial"
)

func main() {
	portNames, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}

	for _, port := range portNames {
		fmt.Printf("Port:%v\n", port)
	}

	args := os.Args

	var fname1, fname2 string
	if len(args) < 3 {
		fname1 = "/dev/tty.usbmodem12301"
		fname2 = "/dev/tty.usbmodem12401"
	} else {
		fname1 = args[1]
		fname2 = args[2]

	}

	reader1, closer1, err := ports.Open(fname1)
	if err != nil {
		log.Fatalf("Could not open port 1: %s: %s", fname1, err.Error())
	}
	reader2, closer2, err := ports.Open(fname2)
	if err != nil {
		log.Fatalf("Could not open port 2: %s: %s", fname2, err.Error())
	}
	defer closer1()
	defer closer2()

	ch, done := ports.ReadTwoFiles(reader1, reader2)

	storage, err := db.ConnectDB("./keypresses.sqlite")
	if err != nil {
		log.Fatalf("Could not open port 2: %s: %s", fname2, err.Error())
	}
	defer storage.Close()

	handler := server.ServerHandler{storage}
	http.Handle("/", http.HandlerFunc(handler.StatsHandle))

	fmt.Println("Listening on :3000")
	go http.ListenAndServe(":3000", nil)
out:
	for {
		select {
		case line := <-ch:
			parsed, err := parser.ParseLine(line)
			if err != nil {
				log.Printf("Got warning: %s\nline: '%s'", err.Error(), line)
			}

			if parsed != nil {
				log.Printf("Event! %v", *parsed)
				storage.Store(parsed)
			}
		case <-done:
			break out
		}
	}
	fmt.Println("Read ended")
}
