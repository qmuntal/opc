package opc

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestWriter_Flush(t *testing.T) {
	tests := []struct {
		name    string
		w       *Writer
		wantErr bool
	}{
		{"base", NewWriter(&bytes.Buffer{}), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.w.Flush(); (err != nil) != tt.wantErr {
				t.Errorf("Writer.Flush() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_Close(t *testing.T) {
	p := newPackage()
	p.contentTypes.add("/a.xml", "a/b")
	p.contentTypes.add("/b.xml", "c/d")
	pCore := newPackage()
	pCore.parts["/PROPS/CORE.XML"] = struct{}{}
	pRel := newPackage()
	pRel.parts["/_RELS/.RELS"] = struct{}{}
	tests := []struct {
		name    string
		w       *Writer
		wantErr bool
	}{
		{"base", NewWriter(&bytes.Buffer{}), false},
		{"withCt", &Writer{p: p, w: zip.NewWriter(&bytes.Buffer{})}, false},
		{"invalidPartRel", &Writer{p: newPackage(), w: zip.NewWriter(&bytes.Buffer{}), last: &Part{Name: "/b.xml", Relationships: []*Relationship{{}}}}, true},
		{"invalidOwnRel", &Writer{p: newPackage(), w: zip.NewWriter(&bytes.Buffer{}), Relationships: []*Relationship{{}}}, true},
		{"withDuplicatedCoreProps", &Writer{p: pCore, w: zip.NewWriter(&bytes.Buffer{}), Properties: CoreProperties{Title: "Song"}}, true},
		{"withDuplicatedRels", &Writer{p: pRel, w: zip.NewWriter(&bytes.Buffer{}), Properties: CoreProperties{Title: "Song"}}, true},
		{"withCoreProps", &Writer{p: newPackage(), w: zip.NewWriter(&bytes.Buffer{}), Properties: CoreProperties{Title: "Song"}}, false},
		{"withCorePropsWithName", &Writer{p: newPackage(), w: zip.NewWriter(&bytes.Buffer{}), Relationships: []*Relationship{
			{TargetURI: "props.xml", Type: corePropsRel},
		}, Properties: CoreProperties{Title: "Song", PartName: "props.xml"}}, false},
		{"withCorePropsWithNameAndId", &Writer{p: newPackage(), w: zip.NewWriter(&bytes.Buffer{}), Properties: CoreProperties{Title: "Song", PartName: "/docProps/props.xml", RelationshipID: "rId1"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.w.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Writer.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_setCompressor(t *testing.T) {
	type args struct {
		fh          *zip.FileHeader
		compression CompressionOption
	}
	tests := []struct {
		name     string
		w        *Writer
		args     args
		wantFlag uint16
	}{
		{"none", NewWriter(nil), args{&zip.FileHeader{}, CompressionNone}, 0x0},
		{"normal", NewWriter(nil), args{&zip.FileHeader{}, CompressionNormal}, 0x0},
		{"max", NewWriter(nil), args{&zip.FileHeader{}, CompressionMaximum}, 0x2},
		{"fast", NewWriter(nil), args{&zip.FileHeader{}, CompressionFast}, 0x4},
		{"sfast", NewWriter(nil), args{&zip.FileHeader{}, CompressionSuperFast}, 0x6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.w.setCompressor(tt.args.fh, tt.args.compression)
			if tt.args.fh.Method != zip.Deflate {
				t.Error("Writer.setCompressor() should have set the method flag the deflate")
			}
		})
	}
}

func Test_compressionFunc(t *testing.T) {
	type args struct {
		comp int
	}
	tests := []struct {
		name string
		args args
	}{
		{"base", args{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressionFunc(tt.args.comp)(&bytes.Buffer{})
		})
	}
}

func TestWriter_Create(t *testing.T) {
	type args struct {
		uri         string
		contentType string
	}
	tests := []struct {
		name    string
		w       *Writer
		args    args
		wantErr bool
	}{
		{"base", NewWriter(&bytes.Buffer{}), args{"/a.xml", "application/xml"}, false},
		{"nameErr", NewWriter(&bytes.Buffer{}), args{"a.xml", "a/b"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.w.Create(tt.args.uri, tt.args.contentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("Writer.Create() want writer")
			}
		})
	}
}

func TestWriter_CreatePart(t *testing.T) {
	rel := &Relationship{ID: "fakeId", Type: "asd", TargetURI: "/fakeTarget", TargetMode: ModeInternal}
	w := NewWriter(&bytes.Buffer{})
	pRel := newPackage()
	pRel.parts["/_RELS/A.XML.RELS"] = struct{}{}
	type args struct {
		part        *Part
		compression CompressionOption
	}
	tests := []struct {
		name    string
		w       *Writer
		args    args
		wantErr bool
	}{
		{"unicode", NewWriter(&bytes.Buffer{}), args{&Part{"/a/ц.xml", "a/b", nil}, CompressionNone}, false},
		{"fhErr", NewWriter(&bytes.Buffer{}), args{&Part{"/a.xml", "a/b", nil}, -3}, true},
		{"nameErr", NewWriter(&bytes.Buffer{}), args{&Part{"a.xml", "a/b", nil}, CompressionNone}, true},
		{"failRel", &Writer{w: zip.NewWriter(nil), last: &Part{Name: "/b.xml", Relationships: []*Relationship{{}}}}, args{&Part{"/a.xml", "a/b", nil}, CompressionNone}, true},
		{"failRel2", &Writer{p: pRel, w: zip.NewWriter(nil), last: &Part{Name: "/a.xml", Relationships: []*Relationship{rel}}}, args{&Part{"/b.xml", "a/b", nil}, CompressionNone}, true},
		{"base", w, args{&Part{"/a.xml", "a/b", nil}, CompressionNone}, false},
		{"multipleDiffName", w, args{&Part{"/b.xml", "a/b", nil}, CompressionNone}, false},
		{"multipleDiffContentType", w, args{&Part{"/c.xml", "c/d", nil}, CompressionNone}, false},
		{"duplicated", w, args{&Part{"/c.xml", "c/d", nil}, CompressionNone}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.w.CreatePart(tt.args.part, tt.args.compression)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.CreatePart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("Writer.CreatePart want a valid writer")
			}
		})
	}
}

func TestWriter_createLastPartRelationships(t *testing.T) {
	rel := &Relationship{ID: "fakeId", Type: "asd", TargetURI: "/fakeTarget", TargetMode: ModeInternal}
	w := NewWriter(&bytes.Buffer{})
	w.last = &Part{Name: "/a.xml", Relationships: []*Relationship{rel}}
	tests := []struct {
		name    string
		w       *Writer
		wantErr bool
	}{
		{"base", &Writer{p: newPackage(), w: zip.NewWriter(nil), last: &Part{Name: "/a.xml", Relationships: []*Relationship{rel}}}, false},
		{"base2", &Writer{p: newPackage(), w: zip.NewWriter(nil), last: &Part{Name: "/b/a.xml", Relationships: []*Relationship{rel}}}, false},
		{"hasSome", w, false},
		{"duplicated", &Writer{w: zip.NewWriter(nil), last: &Part{Name: "/a.xml", Relationships: []*Relationship{rel, rel}}}, true},
		{"invalidRelation", &Writer{w: zip.NewWriter(nil), last: &Part{Name: "/a.xml", Relationships: []*Relationship{{}}}}, true},
		{"empty", NewWriter(&bytes.Buffer{}), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.w.createLastPartRelationships(); (err != nil) != tt.wantErr {
				t.Errorf("Writer.createLastPartRelationships() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewWriterFromReader(t *testing.T) {
	r, err := OpenReader("testdata/office.docx")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	var buf bytes.Buffer
	w, err := NewWriterFromReader(&buf, r.Reader)
	if err != nil {
		t.Fatalf("NewWriterFromReader() error: %v", err)
	}
	if w.Properties != r.Properties {
		t.Error("NewWriterFromReader() haven't copied core properties")
	}
	if len(w.Relationships) != len(r.Relationships) {
		t.Error("NewWriterFromReader() haven't copied package relationships")
	}
	if err = w.Close(); err != nil {
		t.Errorf("NewWriterFromReader() created package that cannot be closed: %v", err)
	}
}
