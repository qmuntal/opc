package gopc

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"reflect"
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

var validContentTypes = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Override ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml"/>
<Default Extension="png" ContentType="image/png"/>
</Types>`

var incorrectOverrideXML = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Override ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml">
</Types>`

var incorrectDefaultXML = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml">
</Types>`

var defaultDuplicated = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Override ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml"/>
<Default Extension="png" ContentType="image/png"/>
<Default Extension="png" ContentType="image/png2"/>
</Types>`

var incorrectType = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Override ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml"/>
<Default Extension="png" ContentType="image/png"/>
<Default Extension="png" ContentType="image/png2"/>
<Fake Extension="" ContentType=""/>
</Types>`

func newMockFile(name string, r io.ReadCloser, e error) *mockFile {
	f := new(mockFile)
	f.On("Name").Return(name)
	if r != nil {
		f.On("Open").Return(r, e)
	}
	return f
}

func Test_newReader(t *testing.T) {
	p1 := newPackage()
	p1.parts["/DOCPROPS/APP.XML"] = &Part{Name: "/docProps/app.xml", ContentType: "application/vnd.openxmlformats-officedocument.extended-properties+xml"}
	p1.parts["/HOLA/PHOTO.PNG"] = &Part{Name: "/hola/photo.png", ContentType: "image/png"}
	p1.contentTypes.addOverride("/docProps/app.xml", "application/vnd.openxmlformats-officedocument.extended-properties+xml")
	p1.contentTypes.addDefault("png", "image/png")

	tests := []struct {
		name    string
		files   []archiveFile
		p       *Package
		wantErr bool
	}{
		{"base", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(validContentTypes)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("hola/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, p1, false},
		{"incorrectContent", []archiveFile{
			newMockFile("a.xml", nil, nil),
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString("content")), nil),
		}, nil, true},
		{"incorrectDefault", []archiveFile{newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("hola/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(defaultDuplicated)), nil),
		}, nil, true},
		{"openError", []archiveFile{
			newMockFile("a.xml", nil, nil),
			newMockFile("[Content_Types].xml", ioutil.NopCloser(nil), errors.New("")),
		}, nil, true},
		{"noError", []archiveFile{
			newMockFile("a.xml", nil, nil),
			newMockFile("a.xml", nil, nil),
		}, nil, true},
		{"fakeType", []archiveFile{
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("hola/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(incorrectType)), nil),
		}, nil, true},
		{"noCType", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(validContentTypes)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("hola/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("fake/fake.fake", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, nil, true},
		{"incorrectDefaultXML", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(incorrectDefaultXML)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("hola/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, nil, true},
		{"incorrectOverrideXML", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(incorrectOverrideXML)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("hola/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := new(mockArchive)
			a.On("Files").Return(tt.files)
			got, err := newReader(a)
			if (err != nil) != tt.wantErr {
				t.Errorf("newReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.p, tt.p) {
				t.Errorf("newReader() = %v, want %v", got, tt.p)
			}
		})
	}
}
