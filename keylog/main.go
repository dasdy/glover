package main

import (
	"bufio"
	"fmt"
	"glover/keylog/parser"
	"log"
	"os"
	"time"

	"go.bug.st/serial"
)

func main() {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}

	for _, port := range ports {
		fmt.Printf("Port:%v\n", port)
	}

	args := os.Args

	var fname string
	if len(args) < 2 {
		fname = "/dev/tty.usbmodem12301"
	} else {
		fname = args[1]
	}

	port, err := serial.Open(fname, &serial.Mode{
		BaudRate: 9600,
	})
	defer func() {
		if port == nil {
			log.Fatal("Port is nil :(")
		}
		if err := port.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	port.SetReadTimeout(10 * time.Second)

	if err != nil {
		log.Fatal(err)
	}

	// print via a buffer
	// buff := make([]byte, 1024)
	// for {
	// 	n, err := port.Read(buff)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	if n == 0 {
	// 		break
	// 	}
	// 	fmt.Printf("%v", string(buff[:n]))
	// }

	// use a scanner!
	scanner := bufio.NewScanner(port)
	for scanner.Scan() {
		line := scanner.Text()
		parsed, err := parser.ParseLine(line)
		if err != nil {
			log.Printf("Got warning: %s\nline: '%s'", err.Error(), line)
		}

		if parsed != nil {
			log.Printf("Event! %v", *parsed)
		}
	}
	fmt.Println("Read ended")
}
