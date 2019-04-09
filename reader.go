package opc

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type archiveFile interface {
	Open() (io.ReadCloser, error)
	Name() string
	Size() int
}

type archive interface {
	Files() []archiveFile
	RegisterDecompressor(method uint16, dcomp func(r io.Reader) io.ReadCloser)
}

// ReadCloser wrapps a Reader than can be closed.
type ReadCloser struct {
	f *os.File
	*Reader
}

// OpenReader will open the OPC file specified by name and return a ReadCloser.
func OpenReader(name string) (*ReadCloser, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	r, err := NewReader(f, fi.Size())
	return &ReadCloser{f: f, Reader: r}, err
}

// Close closes the OPC file, rendering it unusable for I/O.
func (r *ReadCloser) Close() error {
	return r.f.Close()
}

// File is used to read a part from the OPC package.
type File struct {
	*Part
	Size int
	a    archiveFile
}

// Open returns a ReadCloser that provides access to the File's contents.
// Multiple files may be read concurrently.
func (f *File) Open() (io.ReadCloser, error) {
	return f.a.Open()
}

// Reader implements a OPC file reader.
type Reader struct {
	Files         []*File
	Relationships []*Relationship
	Properties    CoreProperties
	p             *pkg
	r             archive
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

// SetDecompressor sets or overrides a custom decompressor for the DEFLATE.
func (r *Reader) SetDecompressor(dcomp func(r io.Reader) io.ReadCloser) {
	r.r.RegisterDecompressor(zip.Deflate, dcomp)
}

func (r *Reader) loadPackage() error {
	ct, rels, err := r.loadPartPropierties()
	if err != nil {
		return err
	}
	files := r.r.Files()
	r.Files = make([]*File, 0, len(files)-1) // -1 is for [Content_Types].xml

	for _, file := range files {
		fileName := "/" + file.Name()
		// skip content types part, relationship parts and directories
		if strings.EqualFold(fileName, contentTypesName) || isRelationshipURI(fileName) || strings.HasSuffix(fileName, "/") {
			continue
		}
		if strings.EqualFold(fileName, r.Properties.PartName) {
			cp, err := r.loadCoreProperties(file)
			if err != nil {
				return err
			}
			r.Properties = *cp
		} else {
			cType, err := ct.findType(fileName)
			if err != nil {
				return err
			}
			part := &Part{Name: fileName, ContentType: cType, Relationships: rels.findRelationship(fileName)}
			r.Files = append(r.Files, &File{part, file.Size(), file})
			if err = r.p.add(part); err != nil {
				return err
			}
		}
	}
	r.p.contentTypes = *ct
	return nil
}

func (r *Reader) loadPartPropierties() (*contentTypes, *relationshipsPart, error) {
	var ct *contentTypes
	rels := new(relationshipsPart)
	for _, file := range r.r.Files() {
		var err error
		name := "/" + file.Name()
		if strings.EqualFold(name, contentTypesName) {
			ct, err = r.loadContentType(file)
		} else if isRelationshipURI(name) {
			if strings.EqualFold(name, packageRelName) {
				err = r.loadPackageRelationships(file)
			} else {
				err = loadRelationships(file, rels)
			}
		}
		if err != nil {
			return nil, nil, err
		}
	}
	if ct == nil {
		return nil, nil, newError(310, "/")
	}
	return ct, rels, nil
}

func (r *Reader) loadContentType(file archiveFile) (*contentTypes, error) {
	// Process descrived in ISO/IEC 29500-2 §10.1.2.4
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	return decodeContentTypes(reader)
}

func (r *Reader) loadCoreProperties(file archiveFile) (*CoreProperties, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	return decodeCoreProperties(reader)
}

func loadRelationships(file archiveFile, rels *relationshipsPart) error {
	reader, err := file.Open()
	if err != nil {
		return err
	}
	rls, err := decodeRelationships(reader)
	if err != nil {
		return err
	}

	// get part name from rels part
	path := strings.Replace(filepath.Dir(filepath.Dir(file.Name())), `\`, "/", -1)
	pname := "/" + path + "/" + strings.TrimSuffix(filepath.Base(file.Name()), filepath.Ext(file.Name()))
	rels.addRelationship(pname, rls)
	return nil
}

func (r *Reader) loadPackageRelationships(file archiveFile) error {
	reader, err := file.Open()
	if err != nil {
		return err
	}
	rls, err := decodeRelationships(reader)
	if err != nil {
		return err
	}
	r.Relationships = rls
	for _, rel := range rls {
		if strings.EqualFold(rel.Type, corePropsRel) {
			r.Properties.PartName = rel.TargetURI
			break
		}
	}
	return nil
}
