package gopc

import (
	"archive/zip"
	"compress/flate"
	"io"
	"time"
)

// Writer implements a OPC file writer.
type Writer struct {
	p *Package
	w *zip.Writer
}

// NewWriter returns a new Writer writing an OPC file to w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{p: newPackage(), w: zip.NewWriter(w)}
}

// Flush flushes any buffered data to the underlying writer.
// Part metadata, relationships, content types and other OPC related files won't be flushed.
// Calling Flush is not normally necessary; calling Close is sufficient.
// Useful to do simultaneos writing and reading.
func (w *Writer) Flush() error {
	return w.w.Flush()
}

// Close finishes writing the opc file.
// It does not close the underlying writer.
func (w *Writer) Close() error {
	return w.w.Close()
}

// Create adds a new Part to the Package. The part's contents must be written before the next call to Create or Close.
// The part URI shall be a valid part name, one can use NormalizePartName before calling Create to normalize the URI as a part name.
func (w *Writer) Create(uri, contentType string, compression CompressionOption) (io.Writer, error) {
	part := &Part{Name: uri, ContentType: contentType}
	return w.add(part, compression)
}

func (w *Writer) add(part *Part, compression CompressionOption) (io.Writer, error) {
	if err := w.p.add(part); err != nil {
		return nil, err
	}

	fh := &zip.FileHeader{
		Name:     part.Name[1:], // remove first "/"
		Modified: time.Now(),
	}
	w.setCompressor(fh, compression)
	pw, err := w.w.CreateHeader(fh)
	if err != nil {
		w.p.deletePart(part.Name)
		return nil, err
	}
	return pw, nil
}

func (w *Writer) setCompressor(fh *zip.FileHeader, compression CompressionOption) {
	var comp int
	switch compression {
	case CompressionNormal:
		comp = flate.DefaultCompression
	case CompressionMaximum:
		comp = flate.BestCompression
		fh.Flags |= 0x2
	case CompressionFast:
		comp = flate.BestSpeed
		fh.Flags |= 0x4
	case CompressionSuperFast:
		comp = flate.BestSpeed
		fh.Flags |= 0x6
	case CompressionNone:
		fh.Method = zip.Store
		return
	}

	fh.Method = zip.Deflate
	w.w.RegisterCompressor(zip.Deflate, compressionFunc(comp))
}

func compressionFunc(comp int) func(out io.Writer) (io.WriteCloser, error) {
	return func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, comp)
	}
}
