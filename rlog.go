package rlog

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
)

const (
	metaKey = "_"
	tsKey   = "ts"
	funcKey = "func"
	lineKey = "line"
)

type M map[string]interface{}

func (m M) Set(key string, v interface{}) M {
	m[key] = v
	return m
}

func (m M) SetMeta(key string, v interface{}) M {
	mm, _ := m[metaKey].(M)
	if mm == nil {
		mm = M{}
		m[metaKey] = mm
	}
	mm[key] = v
	return m
}

func (m M) JSON() json.RawMessage {
	b, _ := json.Marshal(m)
	return json.RawMessage(b)
}

type Options struct {
	NoTimestamp bool
	CallerInfo  bool
}

func New(basePath, pathWithTimeFmt string, opts *Options) *Logger {
	r := NewRotator(basePath, pathWithTimeFmt)

	if opts == nil {
		opts = &Options{}
	}

	return &Logger{
		r:    r,
		enc:  json.NewEncoder(r),
		opts: *opts,
	}
}

type Logger struct {
	mu   sync.Mutex
	r    *Rotator
	enc  *json.Encoder
	opts Options
}

func (l *Logger) Log(msg M) error {
	if msg == nil {
		msg = M{}
	}

	for k, v := range msg {
		switch v := v.(type) {
		case error:
			msg[k] = v.Error()
		}
	}

	if !l.opts.NoTimestamp {
		msg.SetMeta(tsKey, Now().Unix())
	}

	if l.opts.CallerInfo {
		cfunc, line := callerFunc(1)
		msg.SetMeta(funcKey, cfunc)
		msg.SetMeta(lineKey, line)
	}

	rm := msg.JSON()

	l.mu.Lock()
	err := l.enc.Encode(rm)
	l.mu.Unlock()
	return err
}

func (l *Logger) SetIndent(v bool) {
	if l.enc == nil {
		return
	}
	l.mu.Lock()
	l.enc.SetIndent("", "\t")
	l.mu.Unlock()
}

func (l *Logger) Close() error {
	if l.r != nil {
		return l.r.Close()
	}

	return nil
}

func callers(idx int) *runtime.Frames {
	var pc [32]uintptr

	if n := runtime.Callers(3+idx, pc[:]); n > 0 {
		return runtime.CallersFrames(pc[:n])
	}

	return nil
}

var badCallers = map[string]bool{
	"runtime.goexit": true,
}

func callerFunc(idx int) (fn, line string) {
	frames := callers(idx)
	for {
		frame, more := frames.Next()
		fn = filepath.Base(frame.Function)
		line = frame.File + ":" + strconv.Itoa(frame.Line)
		if !more {
			return
		}

		if !badCallers[fn] {
			return
		}

	}
}
