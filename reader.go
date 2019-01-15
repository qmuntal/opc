package gopc

import (
	"archive/zip"
	"errors"
	"io"
	"path/filepath"
)

// Reader implements a OPC file reader.
type Reader struct {
	Parts []*Part
	p     *Package
	r     *zip.Reader
}

// NewReader returns a new Reader reading an OPC file to r.
func NewReader(r io.ReaderAt, size int64) (*Reader, error) {
	read, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}
	return &Reader{r: read}, nil
}

func (r *Reader) loadParts() {
	files := r.r.File
	for i := 0; i < len(files); i++ {
		r.Parts[i] = &Part{Name: files[i].Name}
	}
}

func (r *Reader) getContentType() error {
	// Process descrived in ISO/IEC 29500-2 §10.1.2.4
	files := r.r.File
	var found bool
	for i := 0; i < len(files); i++ {
		if files[i].Name != "[Content_Types].xml" {
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
			if cDefault, ok := c.(defaultContentTypeXML); ok {
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
			} else if cOverride, ok := c.(overrideContentTypeXML); ok {
				for j := 0; j < len(files); j++ {
					if r.Parts[j].Name == cOverride.PartName {
						r.Parts[j].ContentType = cOverride.ContentType
					}
				}
			} else {
				return errors.New("OPC: content types has a element with an unexpected type")
			}
		}
	}
	if !found {
		return errors.New("OPC: the file content type must exist in the package")
	}
	return nil
}
