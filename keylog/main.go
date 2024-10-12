package main

import (
	"bufio"
	"fmt"
	"glover/keylog/parser"
	"io"
	"log"
	"os"
	"time"

	"go.bug.st/serial"
)

func open(path string) (io.Reader, func(), error) {
	port, err := serial.Open(path, &serial.Mode{
		BaudRate: 9600,
	})
	if err != nil {
		return nil, nil, err
	}
	closer := func() {
		if port == nil {
			log.Fatal("Port is nil :(")
		}
		if err := port.Close(); err != nil {
			log.Fatal(err)
		}
	}

	port.SetReadTimeout(10 * time.Second)
	return port, closer, nil
}

func readTwoFiles(f1, f2 io.Reader) (<-chan string, <-chan bool) {
	ch1Done := make(chan bool)
	ch1 := make(chan string)
	go func() {
		scanner := bufio.NewScanner(f1)
		for scanner.Scan() {
			ch1 <- scanner.Text()
		}
		ch1Done <- true
	}()

	ch2Done := make(chan bool)
	ch2 := make(chan string)
	go func() {
		scanner := bufio.NewScanner(f2)
		for scanner.Scan() {
			ch2 <- scanner.Text()
		}
		ch2Done <- true
	}()

	outputChan := make(chan string)
	doneChan := make(chan bool)

	go func() {
		var ch1Closed, ch2Closed bool
		for !ch1Closed || !ch2Closed {
			select {
			case msg := <-ch1:
				outputChan <- msg
			case msg := <-ch2:
				outputChan <- msg
			case <-ch1Done:
				ch1Closed = true
			case <-ch2Done:
				ch2Closed = true
			}
		}

		doneChan <- true
	}()

	return outputChan, doneChan
}

func main() {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}

	for _, port := range ports {
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

	reader1, closer1, err := open(fname1)
	if err != nil {
		log.Fatalf("Could not open port 1: %s: %s", fname1, err.Error())
	}
	reader2, closer2, err := open(fname2)
	if err != nil {
		log.Fatalf("Could not open port 2: %s: %s", fname2, err.Error())
	}

	defer closer1()
	defer closer2()

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

	ch, done := readTwoFiles(reader1, reader2)
	for {
		select {
		case line := <-ch:
			parsed, err := parser.ParseLine(line)
			if err != nil {
				log.Printf("Got warning: %s\nline: '%s'", err.Error(), line)
			}

			if parsed != nil {
				log.Printf("Event! %v", *parsed)
			}
		case <-done:
			break
		}
	}
	fmt.Println("Read ended")
}
