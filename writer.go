package gopc

import (
	"archive/zip"
	"compress/flate"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"time"
)

// Writer implements a OPC file writer.
type Writer struct {
	p    *Package
	w    *zip.Writer
	last *Part
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

type Relationships struct {
	XMLName xml.Name       `xml:"Relationships"`
	XML     string         `xml:"xml"`
	Rels    []relationship `xml:"relations"`
}

type relationship struct {
	XMLName xml.Name `xml:"Relationship"`
	Target  string   `xml:"type,Id,Target"`
}

func (w *Writer) add(part *Part, compression CompressionOption) (io.Writer, error) {
	if w.last != nil && len(w.last.Relationships) != 0 {
		filepath.Dir(w.last.Name)
		relWriter, err := w.w.Create(fmt.Sprintf("%s/_rels/%s.rels", filepath.Dir(w.last.Name)[1:], filepath.Base(w.last.Name)))
		if err != nil {
			return nil, err
		}

		v := &Relationships{XML: "http://schemas.openxmlformats.org/package/2006/relationships"}
		v.Svs = append(v.Svs, server{"Shanghai_VPN", "127.0.0.1"})
		/*e := xml.NewEncoder(relWriter)
		e.EncodeElement(`xml version="1.0" encoding="UTF-8"`, xml.StartElement{})
		e.EncodeElement(`ashd`, xml.StartElement{Name: xml.Name{Space: "", Local: "Relationships"}})

		for _, r := range w.last.Relationships {
			if err = r.writeToXML(e); err != nil {
				return nil, err
			}
		}*/
	}

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
