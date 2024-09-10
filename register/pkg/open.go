package register

import (
	"compress/gzip"
	"io"
	"bufio"
	"regexp"
	"os"
)

var gzRe = regexp.MustCompile(`\.gz$`)

// A buffered reader for a file
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

// A reader for reading from a gzipped file
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

// Open a path and gunzip the input stream
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

// Open a path and gunzip the input stream if the path ends in .gz
func OpenMaybeGz(path string) (io.ReadCloser, error) {
	if gzRe.MatchString(path) {
		return OpenGz(path)
	}
	return Open(path)
}

// A buffered writer to a file
type Writer struct {
	fp *os.File
	*bufio.Writer
}

// A gzipped buffered writer to a file
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

// Create a file and gzip the output stream
func CreateGz(path string) (*GzWriter, error) {
	w, e := Create(path)
	if e != nil {
		return nil, e
	}
	gw := gzip.NewWriter(w)
	return &GzWriter{w, gw}, nil
}

// Create a file and automatically gzip the output stream if the path name ends in .gz
func CreateMaybeGz(path string) (io.WriteCloser, error) {
	if gzRe.MatchString(path) {
		return CreateGz(path)
	}
	return Create(path)
}
