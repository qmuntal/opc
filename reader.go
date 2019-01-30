package gopc

import (
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
	Parts         []*Part
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

func (r *Reader) loadPackage() error {
	files := r.r.Files()
	ct, rels, err := r.loadPartPropierties()
	if err != nil {
		return err
	}
	r.Parts = make([]*Part, 0, len(files)-1) // -1 is for [Content_Types].xml
	for _, file := range files {
		fileName := "/" + file.Name()
		// skip content types part, relationship parts and directories
		if fileName == contentTypesName || isRelationshipURI(fileName) || strings.HasSuffix(fileName, "/") {
			continue
		}
		if fileName == r.Properties.PartName {
			cp, err := r.loadCoreProperties(file)
			if err != nil {
				return err
			}
			r.Properties = *cp
			continue
		}
		cType, err := ct.findType(fileName)
		if err != nil {
			return err
		}
		part := &Part{Name: fileName, ContentType: cType, Relationships: rels.findRelationship(fileName)}
		r.Parts = append(r.Parts, part)
		r.p.add(part)
	}
	r.p.contentTypes = *ct
	return nil
}

func (r *Reader) loadPartPropierties() (*contentTypes, *relationshipsPart, error) {
	ct := new(contentTypes)
	rels := new(relationshipsPart)
	var err error
	foundCT := false
	files := r.r.Files()
	for _, file := range files {
		name := file.Name()
		if name == "[Content_Types].xml" {
			ct, err = r.loadContentType(file)
			if err != nil {
				return nil, nil, err
			}
			foundCT = true
			continue
		}
		name = "/" + name
		if !isRelationshipURI(name) {
			continue
		}
		if name == packageRelName {
			err = r.loadPackageRelationships(file)
			if err != nil {
				return nil, nil, err
			}
			continue
		}

		rls, pname, err := loadRelationships(file)
		if err != nil {
			return nil, nil, err
		}

		rels.addRelationship(pname, rls)
	}
	if !foundCT {
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

func loadRelationships(file archiveFile) ([]*Relationship, string, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, "", err
	}
	rls, err := decodeRelationships(reader)
	if err != nil {
		return nil, "", err
	}
	ext := filepath.Ext(file.Name())
	pname2 := filepath.Base(file.Name())
	path := strings.Replace(filepath.Dir(filepath.Dir(file.Name())), `\`, "/", -1)
	pname := "/" + path + "/" + strings.TrimSuffix(pname2, ext)
	return rls, pname, nil
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
		if rel.Type == corePropsRel {
			r.Properties.PartName = rel.TargetURI
		}
	}
	return nil
}
