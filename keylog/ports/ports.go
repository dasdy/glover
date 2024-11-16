package ports

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
)

func noop() {}

func Open(path string) (io.Reader, func(), error) {
	port, err := serial.Open(path, &serial.Mode{
		BaudRate: 9600,
	})
	if err != nil {
		return nil, noop, err
	}

	closer := func() {
		if port == nil {
			log.Print("Port is nil :(")
			return
		}
		if err := port.Close(); err != nil {
			log.Print(err)
		}
	}

	// TODO make this configurable.
	err = port.SetReadTimeout(10 * time.Hour)
	if err != nil {
		closer()
		// Guarantee that closer is non-null, but close
		// file now because it does not make sense to keep it open.
		return nil, noop, err
	}
	return port, closer, nil
}

// Read from two files at the same time line-by-line. Done channel sends a message
// when both files were closed.
func ReadTwoFiles(f1, f2 io.Reader) <-chan string {
	ch1 := ReadFile(f1)
	ch2 := ReadFile(f2)

	outputChan := make(chan string, 5)
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		for v := range ch1 {
			outputChan <- v
		}
		wg.Done()
		log.Print("Read ch1 routine fin")
	}()

	go func() {
		for v := range ch2 {
			outputChan <- v
		}
		wg.Done()
		log.Print("Read ch2 routine fin")
	}()

	go func() {
		wg.Wait()
		log.Print("Both files marked as closed")
		close(outputChan)
	}()

	return outputChan
}

func OpenTwoFiles(fname1, fname2 string) (<-chan string, func(), error) {
	reader1, closer1, err1 := Open(fname1)
	reader2, closer2, err2 := Open(fname2)
	// Guarantee that closer is non-null and we can close connection if the other fails
	closer := func() {
		closer1()
		closer2()
	}

	if err1 != nil {
		return nil, closer, fmt.Errorf("Could not open port 1: %s: %s", fname1, err1.Error())
	}
	if err2 != nil {
		return nil, closer, fmt.Errorf("Could not open port 2: %s: %s", fname2, err2.Error())
	}

	ch := ReadTwoFiles(reader1, reader2)

	return ch, closer, nil
}

func ReadFile(r io.Reader) <-chan string {
	ch1 := make(chan string, 5)

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			ch1 <- scanner.Text()
		}

		close(ch1)
	}()

	return ch1
}

func LooksLikeZMKDevice(path string) bool {
	return strings.HasPrefix(filepath.Base(path), "tty.usbmodem")
}

func GetAvailableDevices() ([]string, error) {
	names, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)

	for _, n := range names {
		if LooksLikeZMKDevice(n) {
			result = append(result, n)
		}
	}

	return result, nil
}
