package personality

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

// Loader composes the system prompt and keeps it current when hostDir files change.
type Loader struct {
	hostDir      string
	constitution string

	mu      sync.RWMutex
	current string

	watcher *fsnotify.Watcher
	done    chan struct{}
}

// NewLoader creates a Loader for hostDir, loading the constitution from DefaultConstitutionPath.
func NewLoader(hostDir string) (*Loader, error) {
	return NewLoaderWith(hostDir, DefaultConstitutionPath)
}

// NewLoaderWith creates a Loader with an explicit constitution path (testing + custom installs).
func NewLoaderWith(hostDir, constitutionPath string) (*Loader, error) {
	constitution, err := LoadConstitution(constitutionPath)
	if err != nil {
		return nil, fmt.Errorf("load constitution: %w", err)
	}

	l := &Loader{
		hostDir:      hostDir,
		constitution: constitution,
		done:         make(chan struct{}),
	}

	if err := l.recompose(); err != nil {
		return nil, fmt.Errorf("initial compose: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create watcher: %w", err)
	}
	if err := watcher.Add(hostDir); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("watch %s: %w", hostDir, err)
	}
	l.watcher = watcher

	go l.watch()
	return l, nil
}

// Current returns the most recently composed system prompt.
func (l *Loader) Current() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.current
}

// Close stops the watcher goroutine.
func (l *Loader) Close() {
	close(l.done)
	if l.watcher != nil {
		l.watcher.Close()
	}
}

func (l *Loader) recompose() error {
	prompt, err := compose(l.hostDir, l.constitution)
	if err != nil {
		return err
	}
	l.mu.Lock()
	l.current = prompt
	l.mu.Unlock()
	return nil
}

func (l *Loader) watch() {
	for {
		select {
		case <-l.done:
			return
		case event, ok := <-l.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if err := l.recompose(); err != nil {
					log.Error().Err(err).Str("file", event.Name).Msg("personality reload failed")
				} else {
					log.Info().Str("file", event.Name).Msg("personality reloaded")
				}
			}
		case err, ok := <-l.watcher.Errors:
			if !ok {
				return
			}
			log.Error().Err(err).Msg("watcher error")
		}
	}
}
