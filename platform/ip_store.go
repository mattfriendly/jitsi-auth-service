package platform

import (
    "log"
    "sync"
)

type IPStore struct {
    IPs map[string]struct{}
    mu  sync.Mutex
}

func NewIPStore() *IPStore {
    return &IPStore{
        IPs: make(map[string]struct{}),
    }
}

func (store *IPStore) Add(ip string) {
    store.mu.Lock()
    defer store.mu.Unlock()
    store.IPs[ip] = struct{}{}
    log.Printf("Added IP to store: %s", ip)  // Log when an IP is added
}

func (store *IPStore) Exists(ip string) bool {
    store.mu.Lock()
    defer store.mu.Unlock()
    _, exists := store.IPs[ip]
    log.Printf("Checked IP existence: %s, Found: %v", ip, exists)  // Log the check and result
    return exists
}
