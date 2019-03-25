package rlog

import (
	"io"
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
// if pathWithTimeFmt ends with .gz, it'll enable gzip compression.
func NewRotator(basePath, pathWithTimeFmt string) *Rotator {
	var wc func(w io.Writer) io.WriteCloser
	if filepath.Ext(pathWithTimeFmt) == ".gz" {
		wc = GzipWrapper
	}

	return &Rotator{
		base:    basePath,
		path:    pathWithTimeFmt,
		Wrapper: wc,
	}
}

type Rotator struct {
	mu   sync.RWMutex
	f    *os.File
	wc   io.WriteCloser
	base string
	path string

	Wrapper func(w io.Writer) io.WriteCloser
}

func (r *Rotator) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	if err = r.tryRotate(); err == nil {
		n, err = r.writer().Write(p)
	}
	r.mu.Unlock()
	return
}

func (r *Rotator) WriteString(s string) (n int, err error) {
	r.mu.Lock()
	if err = r.tryRotate(); err == nil {
		n, err = io.WriteString(r.writer(), s)
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

func (r *Rotator) WithWriter(fn func(w io.Writer) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.tryRotate(); err != nil {
		return err
	}

	return fn(r.writer())
}

func (r *Rotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.f == nil || r.f == &closedFile {
		return os.ErrClosed
	}

	if r.wc != nil {
		r.wc.Close()
	}

	err := r.f.Close()
	r.f = &closedFile
	return err
}

func (r *Rotator) open(fn string) (err error) {
	if err = os.MkdirAll(filepath.Dir(fn), 0755); err != nil {
		return
	}

	if r.f, err = os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil || r.Wrapper == nil {
		return
	}

	if rwc, ok := r.wc.(resetter); ok {
		rwc.Reset(r.f)
	} else {
		r.wc = r.Wrapper(r.f)
	}

	return
}

func (r *Rotator) writer() io.Writer {
	if r.wc != nil {
		return r.wc
	}

	return r.f
}

func (r *Rotator) tryRotate() (err error) {
	// this is probably slow, should change it
	fn := filepath.Join(r.base, Now().UTC().Format(r.path))

	switch {
	case r.f == &closedFile:
		return os.ErrClosed

	case r.f == nil:
		err = r.open(fn)

	case r.f.Name() != fn:
		if r.wc != nil {
			r.wc.Close()
		}
		if err = r.f.Close(); err != nil {
			r.f = &closedFile
			return
		}
		err = r.open(fn)
	}

	return
}

type resetter interface {
	Reset(w io.Writer)
}
