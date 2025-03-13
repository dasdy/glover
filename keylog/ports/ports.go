package ports

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
)

type RealDeviceReader struct {
	ports []io.ReadCloser
}

type DeviceReader interface {
	Close() error
	Channel() <-chan string
}

type DeviceOpener interface {
	OpenMultiple(paths ...string) (*RealDeviceReader, error)
	GetAvailableDevices() ([]string, error)
}

type RealDeviceOpener struct{}

func NewDeviceReader(devices ...io.ReadCloser) *RealDeviceReader {
	return &RealDeviceReader{ports: devices}
}

func (r *RealDeviceReader) Close() error {
	es := make([]error, 0)

	for _, p := range r.ports {
		err := p.Close()
		if err != nil {
			es = append(es, err)
		}
	}

	if len(es) > 0 {
		return errors.Join(es...)
	}

	return nil
}

func (r *RealDeviceReader) Channel() <-chan string {
	outputChan := make(chan string, 5)

	var wg sync.WaitGroup

	wg.Add(len(r.ports))

	for i, p := range r.ports {
		ch := ReadFile(p)
		go func() {
			for v := range ch {
				outputChan <- v
			}

			wg.Done()
			log.Printf("Read channel %d routine fin", i)
		}()
	}

	go func() {
		wg.Wait()
		log.Print("All files marked as closed")
		close(outputChan)
	}()

	return outputChan
}

func (r *RealDeviceOpener) Open(path string) (*RealDeviceReader, error) {
	port, err := serial.Open(path, &serial.Mode{
		BaudRate: 9600,
	})
	if err != nil {
		return nil, err
	}

	// TODO make this configurable.
	if err := port.SetReadTimeout(10 * time.Hour); err != nil {
		if innerErr := port.Close(); innerErr != nil {
			return nil, fmt.Errorf("error during closing of port: %w, outer error: %w", innerErr, err)
		}

		return nil, err
	}

	return NewDeviceReader(port), nil
}

func CloseReaders(outerError error, itemsToClose []io.ReadCloser) error {
	es := []error{outerError}

	for i, item := range itemsToClose {
		err := item.Close()
		if err != nil {
			es = append(es, fmt.Errorf("error on item %d: %w", i, err))
		}
	}

	if len(es) > 1 {
		return errors.Join(es...)
	}

	return outerError
}

func (r *RealDeviceOpener) OpenMultiple(paths ...string) (*RealDeviceReader, error) {
	ports := make([]io.ReadCloser, len(paths))

	for i, p := range paths {
		reader, err := r.Open(p)
		if err != nil {
			outerError := fmt.Errorf("error on opening path %s: %w", p, err)

			return nil, CloseReaders(outerError, ports[:i])
		}

		if len(reader.ports) != 1 {
			outerError := fmt.Errorf("should not be here: got %d ports on file %s", len(reader.ports), p)

			return nil, CloseReaders(outerError, ports[:i])
		}

		ports[i] = reader.ports[0]
	}

	return NewDeviceReader(ports...), nil
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

func (r *RealDeviceOpener) GetAvailableDevices() ([]string, error) {
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
