package gopc

import (
	"archive/zip"
	"compress/flate"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"
)

const opcCtx = "opc"

// Writer implements a OPC file writer.
type Writer struct {
	p                    *Package
	w                    *zip.Writer
	last                 *Part
	testRelationshipFail bool // Only true for testing
	testContentTypesFail bool // Only true for testing
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
	if err := w.createContentTypes(); err != nil {
		w.w.Close()
		return err
	}
	if err := w.createRelationships(); err != nil {
		w.w.Close()
		return err
	}
	return w.w.Close()
}

// Create adds a file to the OPC archive using the provided name and content type.
// The file contents will be compressed using the Deflate default method.
// The name shall be a valid part name, one can use NormalizePartName before calling Create to normalize it
//
// This returns a Writer to which the file contents should be written.
// The file's contents must be written to the io.Writer before the next call to Create, CreatePart, or Close.
func (w *Writer) Create(name, contentType string) (io.Writer, error) {
	part := &Part{Name: name, ContentType: contentType}
	return w.add(part, CompressionNormal)
}

// CreatePart adds a file to the OPC archive using the provided part.
// The name shall be a valid part name, one can use NormalizePartName before calling CreatePart to normalize it.
// Writer takes ownership of part and may mutate all its fields except the Relationships,
// which can be modified until the next call to Create, CreatePart or Close.
// The caller must not modify part after calling CreatePart, except the Relationships.
//
// This returns a Writer to which the file contents should be written.
// The file's contents must be written to the io.Writer before the next call to Create, CreatePart, or Close.
func (w *Writer) CreatePart(part *Part, compression CompressionOption) (io.Writer, error) {
	return w.add(part, compression)
}

func (w *Writer) createContentTypes() error {
	cw, err := w.w.Create("[Content_Types].xml")
	if w.testContentTypesFail {
		err = errors.New("")
	}
	if err != nil {
		return err
	}
	return w.p.encodeContentTypes(cw)
}

func (w *Writer) createRelationships() error {
	if w.last == nil || len(w.last.Relationships) == 0 {
		return nil
	}
	if err := validateRelationships(w.last.Relationships); err != nil {
		return err
	}
	filepath.Dir(w.last.Name)
	rw, err := w.w.Create(fmt.Sprintf("%s/_rels/%s.rels", filepath.Dir(w.last.Name)[1:], filepath.Base(w.last.Name)))
	if w.testRelationshipFail {
		err = errors.New("")
	}
	if err != nil {
		return err
	}
	return encodeRelationships(rw, w.last.Relationships)
}

func (w *Writer) add(part *Part, compression CompressionOption) (io.Writer, error) {
	if err := w.createRelationships(); err != nil {
		return nil, err
	}

	// Validate name and check for duplicated names ISO/IEC 29500-2 M3.3
	if err := w.p.add(part); err != nil {
		return nil, err
	}

	// ISO/IEC 29500-2 M1.4
	fh := &zip.FileHeader{
		Name:     zipName(part.Name),
		Modified: time.Now(),
	}
	w.setCompressor(fh, compression)
	pw, err := w.w.CreateHeader(fh)
	if err != nil {
		w.p.deletePart(part.Name)
		return nil, err
	}
	w.last = part
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
		comp = flate.NoCompression
	default:
		comp = -1000 // write will failt
	}

	fh.Method = zip.Deflate
	w.w.RegisterCompressor(zip.Deflate, compressionFunc(comp))
}

func compressionFunc(comp int) func(out io.Writer) (io.WriteCloser, error) {
	return func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, comp)
	}
}

func zipName(partName string) string {
	// ISO/IEC 29500-2 M3.4
	return partName[1:] // remove first slash
}
