package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

// acquireLock takes an exclusive flock on <hostDir>/symmated.lock.
// Non-blocking: returns error immediately if another process holds the lock.
// The lock is held for the lifetime of the process; kernel releases on exit/crash.
func acquireLock(hostDir string) (*os.File, error) {
	lockPath := filepath.Join(hostDir, "symmated.lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("symmated already running for this host (lock held at %s)", lockPath)
	}
	// Write PID for observability only. Flock is the actual lock.
	f.Truncate(0)
	f.WriteString(strconv.Itoa(os.Getpid()))
	return f, nil
}
