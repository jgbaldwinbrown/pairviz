package register

import (
	"compress/gzip"
	"io"
	"bufio"
	"regexp"
	"os"
)

var gzRe = regexp.MustCompile(`\.gz$`)

type Reader struct {
	fp *os.File
	*bufio.Reader
}

func (r *Reader) Close() error {
	return r.fp.Close()
}

func Open(path string) (*Reader, error) {
	fp, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	return &Reader{ fp, bufio.NewReader(fp) }, nil
}

type GzReader struct {
	r *Reader
	*gzip.Reader
}

func (r *GzReader) Close() error {
	var err error
	if e := r.Reader.Close(); err == nil {
		err = e
	}
	if e := r.r.Close(); err == nil {
		err = e
	}
	return err
}

func OpenGz(path string) (*GzReader, error) {
	r, e := Open(path)
	if e != nil {
		return nil, e
	}
	gr, e := gzip.NewReader(r)
	if e != nil {
		return nil, e
	}
	return &GzReader{r, gr}, nil
}

func OpenMaybeGz(path string) (io.ReadCloser, error) {
	if gzRe.MatchString(path) {
		return OpenGz(path)
	}
	return Open(path)
}

type Writer struct {
	fp *os.File
	*bufio.Writer
}

type GzWriter struct {
	w *Writer
	*gzip.Writer
}

func (w *Writer) Close() error {
	var err error
	if e := w.Flush(); err == nil {
		err = e
	}
	if e := w.fp.Close(); err == nil {
		err = e
	}
	return err
}

func Create(path string) (*Writer, error) {
	fp, e := os.Create(path)
	if e != nil {
		return nil, e
	}
	return &Writer{ fp, bufio.NewWriter(fp) }, nil
}

func (w *GzWriter) Close() error {
	var err error
	if e := w.Writer.Close(); err == nil {
		err = e
	}
	if e := w.w.Close(); err == nil {
		err = e
	}
	return err
}

func CreateGz(path string) (*GzWriter, error) {
	w, e := Create(path)
	if e != nil {
		return nil, e
	}
	gw := gzip.NewWriter(w)
	return &GzWriter{w, gw}, nil
}

func CreateMaybeGz(path string) (io.WriteCloser, error) {
	if gzRe.MatchString(path) {
		return CreateGz(path)
	}
	return Create(path)
}
