package gopc

import (
	"archive/zip"
	"compress/flate"
	"errors"
	"io"
	"time"
)

// Writer implements a OPC file writer.
type Writer struct {
	Package *Package
	w       *zip.Writer
	last    *Part
}

// NewWriter returns a new Writer writing an OPC file to w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{Package: newPackage(), w: zip.NewWriter(w)}
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
	if w.last != nil {
		w.last.writeRelationships(nil)
		w.last = nil
	}
	return w.w.Close()
}

// Create adds a new Part to the Package. The part's contents must be written before the next call to Create, Copy or Close.
// The part URI shall be a valid part name, one can use NormalizePartName before calling Create to normalize the URI as a part name.
func (w *Writer) Create(uri, contentType string, compressionOption CompressionOption) (*Part, error) {
	part, err := w.Package.create(uri, contentType, compressionOption)
	if err != nil {
		return nil, err
	}

	if err = w.add(part); err != nil {
		return nil, err
	}
	return part, nil
}

// Copy an existing readable Part to the Package.
// The Part will be read until EOF and it won't be readable again.
func (w *Writer) Copy(part *Part) error {
	if part.r == nil {
		return errors.New("OPC: part cannot be copied because it does not have read access")
	}
	if err := w.Package.add(part); err != nil {
		return err
	}
	if err := w.add(part); err != nil {
		return err
	}
	if _, err := io.Copy(part.w, part.r); err != nil {
		return err
	}
	part.r = nil
	return nil
}

func (w *Writer) add(part *Part) error {
	if w.last != nil {
		w.last.writeRelationships(nil)
	}

	fh := &zip.FileHeader{
		Name:     part.uri[1:], // remove first "/"
		Modified: time.Now(),
	}
	w.setCompressor(fh, part.compressionOption)
	pw, err := w.w.CreateHeader(fh)
	if err != nil {
		w.Package.deletePart(part.uri)
		return err
	}
	part.w = pw
	w.last = part
	return nil
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
