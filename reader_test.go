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

func newMockFile(name string, r io.ReadCloser, e error) *mockFile {
	f := new(mockFile)
	f.On("Name").Return(name)
	if r != nil {
		f.On("Open").Return(r, e)
	}
	return f
}

var validContentTypes = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Override ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml"/>
<Override ContentType="application/vnd.openxmlformats-package.core-properties+xml" PartName="/docProps/core.xml"/>
<Default Extension="png" ContentType="image/png"/>
<Default ContentType="application/xml" Extension="xml"/>
</Types>`

var validRelationships = `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rel-1" Type="exampleRelationType" Target="/"/>
<Relationship Id="rel-2" Type="exampleRelationType" Target="/" TargetMode="External"/>
</Relationships>`

func Test_newReader(t *testing.T) {

	p1 := newPackage()
	p1.parts["/DOCPROPS/APP.XML"] = &Part{Name: "/docProps/app.xml", ContentType: "application/vnd.openxmlformats-officedocument.extended-properties+xml"}
	p1.parts["/PICTURES/PHOTO.PNG"] = &Part{Name: "/pictures/photo.png", ContentType: "image/png"}
	p1.parts["/DOCPROPS/CORE.XML"] = &Part{Name: "/docProps/core.xml", ContentType: "application/vnd.openxmlformats-package.core-properties+xml"}
	p1.contentTypes.addOverride("/docProps/app.xml", "application/vnd.openxmlformats-officedocument.extended-properties+xml")
	p1.contentTypes.addOverride("/docProps/core.xml", "application/vnd.openxmlformats-package.core-properties+xml")
	p1.contentTypes.addDefault("png", "image/png")
	p1.contentTypes.addDefault("xml", "application/xml")

	p2 := newPackage()
	p2.parts["/DOCPROPS/APP.XML"] = &Part{Name: "/docProps/app.xml", ContentType: "application/vnd.openxmlformats-officedocument.extended-properties+xml",
		Relationships: []*Relationship{
			&Relationship{ID: "rel-1", RelType: "exampleRelationType", TargetURI: "/", TargetMode: ModeInternal, sourceURI: ""},
			&Relationship{ID: "rel-2", RelType: "exampleRelationType", TargetURI: "/", TargetMode: ModeExternal, sourceURI: ""},
		},
	}
	p2.parts["/PICTURES/PHOTO.PNG"] = &Part{Name: "/pictures/photo.png", ContentType: "image/png"}
	p2.parts["/DOCPROPS/CORE.XML"] = &Part{Name: "/docProps/core.xml", ContentType: "application/vnd.openxmlformats-package.core-properties+xml"}
	p2.contentTypes.addOverride("/docProps/app.xml", "application/vnd.openxmlformats-officedocument.extended-properties+xml")
	p2.contentTypes.addOverride("/docProps/core.xml", "application/vnd.openxmlformats-package.core-properties+xml")
	p2.contentTypes.addDefault("xml", "application/xml")
	p2.contentTypes.addDefault("png", "image/png")

	tests := []struct {
		name    string
		files   []archiveFile
		p       *Package
		wantErr bool
	}{
		{"baseWithRels", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(validContentTypes)), nil),
			newMockFile("docProps/_rels/app.xml.rels", ioutil.NopCloser(bytes.NewBufferString(validRelationships)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("docProps/core.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, p2, false},
		{"baseWithoutRels", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(validContentTypes)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("docProps/core.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, p1, false},
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

var duplicatedExtensionDefault = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="png" ContentType="image/png"/>
<Default Extension="png" ContentType="image/png2"/>
</Types>`

var duplicatedPartNameOverride = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Override ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml"/>
<Override ContentType="application2/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml"/>
</Types>`

var emptyExtension = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Override ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml"/>
<Default Extension="png" ContentType="image/png"/>
<Default Extension="" ContentType="image/png2"/>
</Types>`

var invalidType = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Fake Extension="" ContentType=""/>
</Types>`

var incorrectOverrideXML = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Override ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml">
</Types>`

var incorrectDefaultXML = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml">
</Types>`

func Test_newReader_ContentType(t *testing.T) {
	tests := []struct {
		name    string
		files   []archiveFile
		p       *Package
		wantErr bool
	}{
		{"openError", []archiveFile{
			newMockFile("a.xml", nil, nil),
			newMockFile("[Content_Types].xml", ioutil.NopCloser(nil), errors.New("")),
		}, nil, true},

		{"duplicatedExtensionDefault", []archiveFile{
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(duplicatedExtensionDefault)), nil),
		}, nil, true},

		{"duplicatedPartNameOverride", []archiveFile{
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(duplicatedPartNameOverride)), nil),
		}, nil, true},

		{"emptyExtension", []archiveFile{
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(emptyExtension)), nil),
		}, nil, true},

		{"invalidType", []archiveFile{
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(invalidType)), nil),
		}, nil, true},

		{"incorrectDefaultXML", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(incorrectDefaultXML)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, nil, true},

		{"incorrectOverrideXML", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(incorrectOverrideXML)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, nil, true},

		{"partWithoutContentType", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(validContentTypes)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo2.jpg", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, nil, true},

		{"noContentType", []archiveFile{
			newMockFile("docProps/app.xml", nil, nil),
			newMockFile("pictures/photo2.jpg", nil, nil),
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

var generalRelationships = `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rel-1" Type="exampleRelationType" Target="/docProps/app.xml"/>
</Relationships>`

var relationship2 = `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rel-3" Type="exampleRelationType" Target="/"/>
<Relationship Id="rel-4" Type="exampleRelationType" Target="/"/>
<Relationship Id="rel-5" Type="exampleRelationType" Target="/"/>
</Relationships>`

func Test_newReader_Relationships(t *testing.T) {

	p3 := newPackage()
	p3.parts["/DOCPROPS/APP.XML"] = &Part{Name: "/docProps/app.xml", ContentType: "application/vnd.openxmlformats-officedocument.extended-properties+xml",
		Relationships: []*Relationship{
			&Relationship{ID: "rel-1", RelType: "exampleRelationType", TargetURI: "/", TargetMode: ModeInternal, sourceURI: ""},
			&Relationship{ID: "rel-2", RelType: "exampleRelationType", TargetURI: "/", TargetMode: ModeExternal, sourceURI: ""},
		},
	}
	p3.parts["/PICTURES/PHOTO.PNG"] = &Part{Name: "/pictures/photo.png", ContentType: "image/png"}
	p3.parts["/DOCPROPS/CORE.XML"] = &Part{Name: "/docProps/core.xml", ContentType: "application/vnd.openxmlformats-package.core-properties+xml"}
	p3.contentTypes.addOverride("/docProps/app.xml", "application/vnd.openxmlformats-officedocument.extended-properties+xml")
	p3.contentTypes.addOverride("/docProps/core.xml", "application/vnd.openxmlformats-package.core-properties+xml")
	p3.contentTypes.addDefault("xml", "application/xml")
	p3.contentTypes.addDefault("png", "image/png")

	p4 := newPackage()
	p4.parts["/DOCPROPS/APP.XML"] = &Part{Name: "/docProps/app.xml", ContentType: "application/vnd.openxmlformats-officedocument.extended-properties+xml",
		Relationships: []*Relationship{
			&Relationship{ID: "rel-1", RelType: "exampleRelationType", TargetURI: "/", TargetMode: ModeInternal, sourceURI: ""},
			&Relationship{ID: "rel-2", RelType: "exampleRelationType", TargetURI: "/", TargetMode: ModeExternal, sourceURI: ""},
		},
	}
	p4.parts["/PICTURES/SUMMER/PHOTO.PNG"] = &Part{Name: "/pictures/summer/photo.png", ContentType: "image/png",
		Relationships: []*Relationship{
			&Relationship{ID: "rel-3", RelType: "exampleRelationType", TargetURI: "/", TargetMode: ModeInternal, sourceURI: ""},
			&Relationship{ID: "rel-4", RelType: "exampleRelationType", TargetURI: "/", TargetMode: ModeInternal, sourceURI: ""},
			&Relationship{ID: "rel-5", RelType: "exampleRelationType", TargetURI: "/", TargetMode: ModeInternal, sourceURI: ""},
		},
	}
	p4.parts["/PICTURES/SUMMER/PHOTO2.PNG"] = &Part{Name: "/pictures/summer/photo2.png", ContentType: "image/png"}
	p4.parts["/DOCPROPS/CORE.XML"] = &Part{Name: "/docProps/core.xml", ContentType: "application/vnd.openxmlformats-package.core-properties+xml"}
	p4.contentTypes.addOverride("/docProps/app.xml", "application/vnd.openxmlformats-officedocument.extended-properties+xml")
	p4.contentTypes.addOverride("/docProps/core.xml", "application/vnd.openxmlformats-package.core-properties+xml")
	p4.contentTypes.addDefault("xml", "application/xml")
	p4.contentTypes.addDefault("png", "image/png")

	tests := []struct {
		name    string
		files   []archiveFile
		p       *Package
		wantErr bool
	}{
		{"complexRelationships", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(validContentTypes)), nil),
			newMockFile("_rels/.rels", ioutil.NopCloser(bytes.NewBufferString(generalRelationships)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("docProps/_rels/app.xml.rels", ioutil.NopCloser(bytes.NewBufferString(validRelationships)), nil),
			newMockFile("docProps/core.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, p3, false},

		{"ComplexRoute", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(validContentTypes)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("docProps/_rels/app.xml.rels", ioutil.NopCloser(bytes.NewBufferString(validRelationships)), nil),
			newMockFile("docProps/core.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/summer/photo2.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/summer/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/summer/_rels/photo.png.rels", ioutil.NopCloser(bytes.NewBufferString(relationship2)), nil),
		}, p4, false},

		{"openError", []archiveFile{
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(validContentTypes)), nil),
			newMockFile("docProps/_rels/app.xml.rels", ioutil.NopCloser(nil), errors.New("")),
		}, nil, true},

		{"decodeError", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(validContentTypes)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("docProps/_rels/app.xml.rels", ioutil.NopCloser(bytes.NewBufferString("relations")), nil),
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
