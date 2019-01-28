package gopc

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
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
	p     *pkg
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
	rels, err := r.loadRelationships()
	if err != nil {
		return err
	}
	for _, file := range files {
		fileName := file.Name()
		if fileName == "[Content_Types].xml" || isRelationshipURI(fileName) {
			continue
		}
		name := "/" + fileName
		cType, err := ct.findType(name)
		if err != nil {
			return err
		}
		part := &Part{Name: name, ContentType: cType, Relationships: rels.findRelationship(name)}
		r.p.add(part)
	}
	r.p.contentTypes = *ct
	return nil
}

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
	// ISO/IEC 29500-2 M2.4
	return nil, errors.New("OPC: the file content type must exist in the package")
}

func (r *Reader) loadRelationships() (*relationshipsPart, error) {
	files := r.r.Files()
	rels := new(relationshipsPart)
	for _, file := range files {
		name := file.Name()
		if !isRelationshipURI(name) {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		rls, err := decodeRelationships(reader)
		if err != nil {
			return nil, err
		}
		ext := filepath.Ext(name)
		pname2 := filepath.Base(name)
		path := filepath.Dir(filepath.Dir(name))
		path = strings.Replace(path, `\`, "/", -1)
		pname := "/" + path + "/" + strings.TrimSuffix(pname2, ext)
		rels.addRelationship(pname, rls)
	}
	return rels, nil
}
