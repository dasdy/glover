package ports

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

type MonitoringDeviceReader struct {
	pathToLookup string

	devicesList map[string]*RealDeviceReader
	lock        sync.RWMutex

	opener *RealDeviceOpener

	pollingInterval time.Duration
}

func DefaultMonitoringDeviceReader() *MonitoringDeviceReader {
	return &MonitoringDeviceReader{
		pathToLookup: "/dev/",
		devicesList:  make(map[string]*RealDeviceReader),
		lock:         sync.RWMutex{},
		opener:       &RealDeviceOpener{},
	}
}

func NewMonitoringDeviceReader(pathToLookup string) *MonitoringDeviceReader {
	return &MonitoringDeviceReader{
		pathToLookup:    pathToLookup,
		devicesList:     make(map[string]*RealDeviceReader),
		lock:            sync.RWMutex{},
		opener:          &RealDeviceOpener{},
		pollingInterval: 5 * time.Second,
	}
}

func (r *MonitoringDeviceReader) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	for i, device := range r.devicesList {
		if err := device.Close(); err != nil {
			return fmt.Errorf("error closing device %s: %w", i, err)
		}
	}

	return nil
}

func (r *MonitoringDeviceReader) CloseDevice(devicePath string) error {
	log.Printf("Closing device %s", devicePath)
	r.lock.Lock()
	defer r.lock.Unlock()

	if device, exists := r.devicesList[devicePath]; exists {
		if err := device.Close(); err != nil {
			return fmt.Errorf("error closing device %s: %w", devicePath, err)
		}

		delete(r.devicesList, devicePath)
		log.Printf("Device %s closed and removed from list", devicePath)
	} else {
		log.Printf("Device %s not found in list, nothing to close", devicePath)
	}

	log.Printf("Closing device %s", devicePath)

	return nil
}

func (r *MonitoringDeviceReader) AddDevice(devicePath string, out chan string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, exists := r.devicesList[devicePath]; exists {
		log.Printf("Device %s already exists, skipping", devicePath)

		return nil
	}

	device, err := r.opener.Open(devicePath)
	if err != nil {
		return fmt.Errorf("error opening device %s: %w", devicePath, err)
	}

	r.devicesList[devicePath] = device

	go func() {
		log.Printf("Device %s loop started", devicePath)
		// TODO: is repeat-closing ok?
		defer device.Close()

		for line := range device.Channel() {
			out <- line
		}

		log.Printf("Device %s closed", devicePath)

		err := r.CloseDevice(devicePath)
		if err != nil {
			log.Printf("Could not close device %s: %s", devicePath, err.Error())
		}
	}()

	return nil
}

func (r *MonitoringDeviceReader) FindDevices() ([]string, error) {
	log.Printf("Finding devices in path: %s", r.pathToLookup)

	entries, err := os.ReadDir(r.pathToLookup)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %w", r.pathToLookup, err)
	}

	result := make([]string, 0)

	for _, entry := range entries {
		if entry.IsDir() || entry.Type()&os.ModeDevice == 0 {
			continue
		}

		devicePath := path.Join(r.pathToLookup, entry.Name())

		if !LooksLikeZMKDevice(devicePath) {
			continue
		}

		if _, ok := r.devicesList[devicePath]; ok {
			continue
		}

		log.Printf("Found device: %s", devicePath)

		result = append(result, devicePath)
	}

	return result, nil
}

func (r *MonitoringDeviceReader) Channel() (<-chan string, error) {
	log.Printf("Starting monitoring on path: %s", r.pathToLookup)

	outputChan := make(chan string, 5)

	go func() {
		log.Printf("Monitoring started on path: %s", r.pathToLookup)

		defer log.Printf("End monitoring on path: %s", r.pathToLookup)

		for {
			log.Printf("Polling for devices in path: %s", r.pathToLookup)

			devices, err := r.FindDevices()
			if err != nil {
				log.Printf("Error finding devices: %v", err)

				continue
			}

			time.Sleep(r.pollingInterval)

			for _, devicePath := range devices {
				log.Printf("Processing device: %s", devicePath)

				err := r.AddDevice(devicePath, outputChan)
				if err != nil {
					log.Printf("could not add device %s: %s", devicePath, err.Error())
				}
			}
		}
	}()

	log.Printf("Returning a channel")

	return outputChan, nil
}
