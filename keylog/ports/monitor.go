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
		pathToLookup:    "/dev/",
		devicesList:     make(map[string]*RealDeviceReader),
		lock:            sync.RWMutex{},
		opener:          &RealDeviceOpener{},
		pollingInterval: 5 * time.Second,
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
	log.Printf("[1] Closing device %s", devicePath)
	r.lock.Lock()
	defer r.lock.Unlock()

	if device, exists := r.devicesList[devicePath]; exists {
		if err := device.Close(); err != nil {
			return fmt.Errorf("error closing device %s: %w", devicePath, err)
		}

		delete(r.devicesList, devicePath)
		log.Printf("[2] Device %s closed and removed from list", devicePath)
	} else {
		log.Printf("[2] Device %s not found in list, nothing to close", devicePath)
	}

	log.Printf("[3] Closing device %s done", devicePath)

	return nil
}

func (r *MonitoringDeviceReader) AddDevice(devicePath string, out chan string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, exists := r.devicesList[devicePath]; exists {
		log.Printf("[Opener] Device %s already exists, skipping", devicePath)

		return nil
	}

	device, err := r.opener.Open(devicePath)
	if err != nil {
		return fmt.Errorf("error opening device %s: %w", devicePath, err)
	}

	r.devicesList[devicePath] = device

	go func() {
		log.Printf("[Monitor loop] Device %s loop started", devicePath)
		// TODO: is repeat-closing ok?
		defer device.Close()

		for line := range device.Channel() {
			out <- line
		}

		log.Printf("[Monitor loop] Device %s closed", devicePath)

		err := r.CloseDevice(devicePath)
		if err != nil {
			log.Printf("[Monitor loop] Could not close device %s: %s", devicePath, err.Error())
		}
	}()

	return nil
}

func (r *MonitoringDeviceReader) FindDevices() ([]string, error) {
	// log.Printf("Finding devices in path: %s", r.pathToLookup)
	entries, err := os.ReadDir(r.pathToLookup)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %w", r.pathToLookup, err)
	}

	result := make([]string, 0)

	for _, entry := range entries {
		shouldOpen, devicePath := r.shouldOpen(entry)
		if !shouldOpen {
			continue
		}

		log.Printf("[Finder] Found device: %s", devicePath)

		result = append(result, devicePath)
	}

	return result, nil
}

func (r *MonitoringDeviceReader) Channel() (<-chan string, error) {
	log.Printf("[Monitor-main-launcher] Starting monitoring on path: %s", r.pathToLookup)

	outputChan := make(chan string, 5)

	go func() {
		log.Printf("[Monitor-main] Monitoring started on path: %s", r.pathToLookup)

		defer log.Printf("[Monitor-main] End monitoring on path: %s", r.pathToLookup)

		for {
			// log.Printf("Polling for devices in path: %s", r.pathToLookup)
			devices, err := r.FindDevices()
			if err != nil {
				log.Printf("[Monitor-main] Error finding devices: %v", err)

				continue
			}

			time.Sleep(r.pollingInterval)

			for _, devicePath := range devices {
				log.Printf("[Monitor-main] Processing device: %s", devicePath)

				err := r.AddDevice(devicePath, outputChan)
				if err != nil {
					log.Printf("[Monitor-main] could not add device %s: %s", devicePath, err.Error())
				}
			}
		}
	}()

	log.Printf("[Monitor-main-launcher] Returning a channel")

	return outputChan, nil
}

func (r *MonitoringDeviceReader) shouldOpen(entry os.DirEntry) (bool, string) {
	if entry.IsDir() || entry.Type()&os.ModeDevice == 0 {
		return false, ""
	}

	devicePath := path.Join(r.pathToLookup, entry.Name())

	if !LooksLikeZMKDevice(devicePath) {
		return false, ""
	}

	r.lock.RLock()
	defer r.lock.RUnlock()

	if _, ok := r.devicesList[devicePath]; ok {
		return false, ""
	}

	return true, devicePath
}
