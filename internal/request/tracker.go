package request

import (
	"sync"
)

// RequestInfo holds information about the current API request
type RequestInfo struct {
	Provider string
	Model    string
	URL      string
}

var (
	currentRequest RequestInfo
	mu             sync.RWMutex
)

// SetCurrent updates the current request information
func SetCurrent(provider, model, url string) {
	mu.Lock()
	defer mu.Unlock()
	currentRequest = RequestInfo{
		Provider: provider,
		Model:    model,
		URL:      url,
	}
}

// GetCurrent returns the current request information
func GetCurrent() RequestInfo {
	mu.RLock()
	defer mu.RUnlock()
	return currentRequest
}

// Clear clears the current request information
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	currentRequest = RequestInfo{}
}
