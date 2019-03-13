package rlog

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	// this is used to help debug
	Now        = time.Now
	closedFile os.File
)

// NewRotator returns a new Rotator with the given path.
// It uses time.Time{}.Format, ex: /var/log/2006/01/02.log to split by year/month
func NewRotator(basePath, pathWithTimeFmt string) *Rotator {
	return &Rotator{
		base: basePath,
		path: pathWithTimeFmt,
	}
}

type Rotator struct {
	mu   sync.RWMutex
	f    *os.File
	base string
	path string
}

func (r *Rotator) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	if err = r.tryRotate(); err == nil {
		n, err = r.f.Write(p)
	}
	r.mu.Unlock()
	return
}

func (r *Rotator) WriteString(s string) (n int, err error) {
	r.mu.Lock()
	if err = r.tryRotate(); err == nil {
		n, err = r.f.WriteString(s)
	}
	r.mu.Unlock()
	return
}

func (r *Rotator) Name() (fpath string) {
	r.mu.RLock()
	if r.f != nil {
		fpath = r.f.Name()
	}
	r.mu.RUnlock()

	return
}

func (r *Rotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.f == nil || r.f == &closedFile {
		return os.ErrClosed
	}

	err := r.f.Close()
	r.f = &closedFile
	return err
}

func (r *Rotator) open(fn string) (err error) {
	if err = os.MkdirAll(filepath.Dir(fn), 0755); err != nil {
		return
	}
	r.f, err = os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	return
}

func (r *Rotator) tryRotate() (err error) {
	fn := filepath.Join(r.base, Now().UTC().Format(r.path))

	switch {
	case r.f == &closedFile:
		return os.ErrClosed

	case r.f == nil:
		err = r.open(fn)

	case r.f.Name() != fn:
		if err = r.f.Close(); err != nil {
			r.f = &closedFile
			return
		}
		err = r.open(fn)
	}

	return
}
