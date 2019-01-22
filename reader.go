package gopc

import (
	"errors"
	"io"
	"path/filepath"
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
	return newReader(zr), nil
}

// newReader returns a new Reader reading an OPC file to r.
func newReader(a archive) *Reader {
	return &Reader{p: newPackage(), r: a}
}

// LoadPackage has the responsability of set the phisical file readed into a logical package
// so in the future it can be parsed or readed.
func (r *Reader) LoadPackage() {
	r.loadParts()
	r.getContentType()

	for i := 0; i < len(r.Parts); i++ {
		r.p.add(r.Parts[i])
	}
}

func (r *Reader) loadParts() {
	files := r.r.Files()
	r.Parts = make([]*Part, len(files))
	for i := 0; i < len(files); i++ {
		r.Parts[i] = &Part{Name: files[i].Name()}
	}
}

func (r *Reader) getContentType() error {
	r.loadParts()
	// Process descrived in ISO/IEC 29500-2 §10.1.2.4
	files := r.r.Files()
	var found bool
	for i := 0; i < len(files); i++ {
		if files[i].Name() != "[Content_Types].xml" {
			continue
		}

		found = true
		reader, err := files[i].Open()
		if err != nil {
			return err
		}
		ct, err := decodeContentTypes(reader)
		if err != nil {
			return err
		}

		for _, c := range ct.Types {
			if cDefault, ok := c.Value.(defaultContentTypeXML); ok {
				ext := cDefault.Extension
				for j := 0; j < len(files); j++ {
					if filepath.Ext(r.Parts[j].Name)[1:] == ext {
						if r.Parts[j].ContentType == "" {
							r.Parts[j].ContentType = cDefault.ContentType
						} else {
							return errors.New("OPC: there must be only one Default content type for each extension")
						}
					}
				}
			} else if cOverride, ok := c.Value.(overrideContentTypeXML); ok {
				for j := 0; j < len(files); j++ {
					if r.Parts[j].Name == cOverride.PartName[1:] {
						r.Parts[j].ContentType = cOverride.ContentType
					}
				}
			}
		}
	}
	if !found {
		return errors.New("OPC: the file content type must exist in the package")
	}
	return nil
}
