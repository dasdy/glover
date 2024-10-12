package ports

import (
	"bufio"
	"io"
	"log"
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

	port.SetReadTimeout(10 * time.Second)
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
