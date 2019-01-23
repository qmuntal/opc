package gopc

import (
	"errors"
	"io"
)

type archiveFile interface {
	Open() (io.ReadCloser, error)
	Name() string
}

type archive interface {
	Files() []archiveFile
}

// Reader implements a OPC file reader.
type Reader struct {
	Parts []*Part
	p     *Package
	r     archive
}

// NewReader returns a new Reader reading an OPC file to r.
func NewReader(r io.ReaderAt, size int64) (*Reader, error) {
	zr, err := newZipReader(r, size)
	if err != nil {
		return nil, err
	}
	return newReader(zr)
}

// newReader returns a new Reader reading an OPC file to r.
func newReader(a archive) (*Reader, error) {
	r := &Reader{p: newPackage(), r: a}
	if err := r.loadPackage(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Reader) loadPackage() error {
	files := r.r.Files()
	r.Parts = make([]*Part, len(files)-1) // -1 is for [Content_Types].xml
	ct, err := r.loadContentType()
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.Name() == "[Content_Types].xml" {
			continue
		}
		n := "/" + file.Name()
		cType, err := ct.findType(n)
		if err != nil {
			return err
		}
		part := &Part{Name: n, ContentType: cType}
		r.p.add(part)
	}
	r.p.contentTypes = *ct
	return nil
}

//isRelationshipURI(uri string) bool

func (r *Reader) loadContentType() (*contentTypes, error) {
	// Process descrived in ISO/IEC 29500-2 §10.1.2.4
	files := r.r.Files()
	for _, file := range files {
		if file.Name() != "[Content_Types].xml" {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		return decodeContentTypes(reader)
	}
	return nil, errors.New("OPC: the file content type must exist in the package")
}
