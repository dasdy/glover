package ports

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"sync"
	"time"

	"go.bug.st/serial"
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
	slog.Info("Closing device", "path", devicePath)

	r.lock.Lock()
	defer r.lock.Unlock()

	if device, exists := r.devicesList[devicePath]; exists {
		if err := device.Close(); err != nil {
			return fmt.Errorf("error closing device %s: %w", devicePath, err)
		}

		delete(r.devicesList, devicePath)
		slog.Info("Device closed and removed from list", "path", devicePath)
	} else {
		slog.Info("Device not found in list", "path", devicePath)
	}

	slog.Info("Device closing completed", "path", devicePath)

	return nil
}

func (r *MonitoringDeviceReader) AddDevice(devicePath string, out chan string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, exists := r.devicesList[devicePath]; exists {
		slog.Debug("Device already exists, skipping", "path", devicePath)

		return nil
	}

	device, err := r.opener.Open(devicePath)
	if err != nil {
		return fmt.Errorf("error opening device %s: %w", devicePath, err)
	}

	r.devicesList[devicePath] = device

	go func() {
		// TODO: is repeat-closing ok?
		slog.Info("Device loop started", "path", devicePath)

		defer device.Close()

		for line := range device.Channel() {
			out <- line
		}

		slog.Info("Device closed", "path", devicePath)

		err := r.CloseDevice(devicePath)
		if err != nil {
			slog.Error("Could not close device", "path", devicePath, "error", err)
		}
	}()

	return nil
}

func (r *MonitoringDeviceReader) FindDevices() ([]string, error) {
	slog.Debug("Finding devices in path:", "pathToLookup", r.pathToLookup)

	serialDevices, err := serial.GetPortsList()
	if err != nil {
		return nil, fmt.Errorf("could not get list of serial ports: %w", err)
	}

	slog.Debug("serial devices", "names", serialDevices)

	entries, err := os.ReadDir(r.pathToLookup)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %w", r.pathToLookup, err)
	}

	newDevices := make(map[string]bool)

	for _, devicePath := range serialDevices {
		shouldOpen := r.shouldOpenDevice(devicePath)
		if shouldOpen {
			newDevices[devicePath] = true
		}
	}

	for _, entry := range entries {
		shouldOpen, devicePath := r.shouldOpenFile(entry)
		if !shouldOpen {
			continue
		}

		slog.Info("Found device", "path", devicePath)

		newDevices[devicePath] = true
	}

	keys := make([]string, len(newDevices))

	i := 0
	for k := range newDevices {
		keys[i] = k
		i++
	}

	return keys, nil
}

func (r *MonitoringDeviceReader) Channel() (<-chan string, error) {
	slog.Info("Starting monitoring", "path", r.pathToLookup)

	outputChan := make(chan string, 5)

	go func() {
		slog.Info("Monitoring started", "path", r.pathToLookup)

		defer slog.Info("End monitoring", "path", r.pathToLookup)

		for {
			// log.Printf("Polling for devices in path: %s", r.pathToLookup)
			devices, err := r.FindDevices()
			if err != nil {
				slog.Error("Error finding devices", "error", err)

				continue
			}

			time.Sleep(r.pollingInterval)

			for _, devicePath := range devices {
				slog.Info("Processing device", "path", devicePath)

				err := r.AddDevice(devicePath, outputChan)
				if err != nil {
					slog.Error("Could not add device", "path", devicePath, "error", err)
				}
			}
		}
	}()

	slog.Info("Returning monitoring channel")

	return outputChan, nil
}

func (r *MonitoringDeviceReader) shouldOpenFile(entry os.DirEntry) (bool, string) {
	if entry.IsDir() || entry.Type()&os.ModeDevice == 0 {
		return false, ""
	}

	devicePath := path.Join(r.pathToLookup, entry.Name())

	resp := r.shouldOpenDevice(devicePath)
	if resp {
		return resp, devicePath
	} else {
		return false, ""
	}
}

func (r *MonitoringDeviceReader) shouldOpenDevice(devicePath string) bool {
	if !LooksLikeZMKDevice(devicePath) {
		return false
	}

	r.lock.RLock()
	defer r.lock.RUnlock()

	if _, ok := r.devicesList[devicePath]; ok {
		return false
	}

	return true
}
