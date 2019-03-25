package rlog

import (
	"compress/gzip"
	"io"
)

func GzipWrapper(w io.Writer) io.WriteCloser {
	return gzip.NewWriter(w)
}
