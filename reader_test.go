package gopc

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/mock"
)

type mockFile struct {
	mock.Mock
}

func (m *mockFile) Open() (io.ReadCloser, error) {
	args := m.Called()
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *mockFile) Name() string {
	args := m.Called()
	return args.String(0)
}

type mockArchive struct {
	mock.Mock
}

func (m *mockArchive) Files() []archiveFile {
	args := m.Called()
	return args.Get(0).([]archiveFile)
}

func TestReader_loadParts(t *testing.T) {
	tests := []struct {
		name string
		r    *Reader
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.loadParts()
		})
	}
}

var buff = bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Override ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml"/>
<Default Extension="png" ContentType="image/png"/>
</Types>`)

var buff2 = bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Override ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml"/>
<Default Extension="png" ContentType="image/png"/>
<Default Extension="png" ContentType="image/png2"/>
</Types>`)

func TestReader_getContentType(t *testing.T) {
	f1 := newMockFile("a.xml", nil, nil)
	f2 := newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString("content")), nil)
	f3 := newMockFile("[Content_Types].xml", ioutil.NopCloser(nil), errors.New(""))
	f4 := newMockFile("[Content_Types].xml", ioutil.NopCloser(buff), nil)
	f5 := newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("holi")), nil)
	f6 := newMockFile(".png", ioutil.NopCloser(bytes.NewBufferString("aixo es una imatge")), nil)
	f7 := newMockFile("[Content_Types].xml", ioutil.NopCloser(buff2), nil)
	tests := []struct {
		name    string
		files   []archiveFile
		wantErr bool
	}{
		{"base", []archiveFile{f4, f5, f6}, false},
		{"incorrectContent", []archiveFile{f1, f2}, true},
		{"incorrectDefault", []archiveFile{f5, f6, f7}, true},
		{"openError", []archiveFile{f1, f3}, true},
		{"noError", []archiveFile{f1, f1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := new(mockArchive)
			r := newReader(a)
			a.On("Files").Return(tt.files)
			if err := r.getContentType(); (err != nil) != tt.wantErr {
				t.Errorf("Reader.getContentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			a.AssertExpectations(t)
		})
	}
	f1.AssertExpectations(t)
	f2.AssertExpectations(t)
	f3.AssertExpectations(t)
}

func newMockFile(name string, r io.ReadCloser, e error) *mockFile {
	f := new(mockFile)
	f.On("Name").Return(name)
	if r != nil {
		f.On("Open").Return(r, e)
	}
	return f
}
