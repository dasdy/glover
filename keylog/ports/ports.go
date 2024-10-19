package ports

import (
	"bufio"
	"io"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

func Open(path string) (r io.Reader, closer func(), err error) {
	port, err := serial.Open(path, &serial.Mode{
		BaudRate: 9600,
	})
	if err != nil {
		return nil, nil, err
	}

	c := func() {
		if port == nil {
			log.Fatal("Port is nil :(")
		}
		if err := port.Close(); err != nil {
			log.Fatal(err)
		}
	}

	// TODO make this configurable.
	port.SetReadTimeout(10 * time.Hour)
	return port, c, nil
}

// Read from two files at the same time line-by-line. Done channel sends a message
// when both files were closed.
func ReadTwoFiles(f1, f2 io.Reader) (<-chan string, <-chan bool) {
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

func OpenTwoFiles(fname1, fname2 string) (<-chan string, <-chan bool, func()) {
	reader1, closer1, err := Open(fname1)
	if err != nil {
		log.Fatalf("Could not open port 1: %s: %s", fname1, err.Error())
	}
	reader2, closer2, err := Open(fname2)
	if err != nil {
		log.Fatalf("Could not open port 2: %s: %s", fname2, err.Error())
	}

	closer := func() {
		closer1()
		closer2()
	}

	ch, done := ReadTwoFiles(reader1, reader2)

	return ch, done, closer
}

func GetAvailableDevices() ([]string, error) {
	names, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)

	for _, n := range names {
		if strings.Contains(n, "tty.usbmodem") {
			result = append(result, n)
		}
	}

	if len(names) != 0 {
		return result, nil
	} else {
		return names, nil
	}
}
