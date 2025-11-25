package runtime

import "sync"

var AppState = struct {
    ScanRunning   bool
    UpdateRunning bool
    Lock          sync.Mutex
}{}