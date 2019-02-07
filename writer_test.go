package opc

import (
	"archive/zip"
	"bytes"
	"math/rand"
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

func fakeRand() *rand.Rand {
	return rand.New(rand.NewSource(42))
}

func TestWriter_Close(t *testing.T) {
	p := newPackage()
	p.contentTypes.add("/a.xml", "a/b")
	p.contentTypes.add("/b.xml", "c/d")
	pC := newPackage()
	pC.parts["/[CONTENT_TYPES].XML"] = new(Part)
	pCore := newPackage()
	pCore.parts["/PROPS/CORE.XML"] = new(Part)
	pRel := newPackage()
	pRel.parts["/_RELS/.RELS"] = new(Part)
	tests := []struct {
		name    string
		w       *Writer
		wantErr bool
	}{
		{"base", NewWriter(&bytes.Buffer{}), false},
		{"invalidContentType", &Writer{p: pC, w: zip.NewWriter(&bytes.Buffer{}), rnd: fakeRand()}, true},
		{"withCt", &Writer{p: p, w: zip.NewWriter(&bytes.Buffer{}), rnd: fakeRand()}, false},
		{"invalidPartRel", &Writer{p: newPackage(), w: zip.NewWriter(&bytes.Buffer{}), last: &Part{Name: "/b.xml", Relationships: []*Relationship{{}}}, rnd: fakeRand()}, true},
		{"invalidOwnRel", &Writer{p: newPackage(), w: zip.NewWriter(&bytes.Buffer{}), Relationships: []*Relationship{{}}, rnd: fakeRand()}, true},
		{"withDuplicatedCoreProps", &Writer{p: pCore, w: zip.NewWriter(&bytes.Buffer{}), Properties: CoreProperties{Title: "Song"}, rnd: fakeRand()}, true},
		{"withDuplicatedRels", &Writer{p: pRel, w: zip.NewWriter(&bytes.Buffer{}), Properties: CoreProperties{Title: "Song"}, rnd: fakeRand()}, true},
		{"withDuplicatedRels", &Writer{p: pRel, w: zip.NewWriter(&bytes.Buffer{}), Properties: CoreProperties{Title: "Song"}, rnd: fakeRand()}, true},
		{"withCoreProps", &Writer{p: newPackage(), w: zip.NewWriter(&bytes.Buffer{}), Properties: CoreProperties{Title: "Song"}, rnd: fakeRand()}, false},
		{"withCorePropsWithName", &Writer{p: newPackage(), w: zip.NewWriter(&bytes.Buffer{}), Properties: CoreProperties{Title: "Song", PartName: "/props.xml"}, rnd: fakeRand()}, false},
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
		{"nameErr", NewWriter(&bytes.Buffer{}), args{"a.xml", "a/b"}, true},
		{"base", NewWriter(&bytes.Buffer{}), args{"/a.xml", "a/b"}, false},
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
	pRel.parts["/_RELS/A.XML.RELS"] = new(Part)
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
		{"fhErr", NewWriter(&bytes.Buffer{}), args{&Part{"/a.xml", "a/b", nil}, -3}, true},
		{"nameErr", NewWriter(&bytes.Buffer{}), args{&Part{"a.xml", "a/b", nil}, CompressionNone}, true},
		{"failRel", &Writer{w: zip.NewWriter(nil), last: &Part{Name: "/b.xml", Relationships: []*Relationship{{}}}, rnd: fakeRand()}, args{&Part{"/a.xml", "a/b", nil}, CompressionNone}, true},
		{"failRel2", &Writer{p: pRel, w: zip.NewWriter(nil), last: &Part{Name: "/a.xml", Relationships: []*Relationship{rel}}, rnd: fakeRand()}, args{&Part{"/b.xml", "a/b", nil}, CompressionNone}, true},
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
		{"base", &Writer{p: newPackage(), w: zip.NewWriter(nil), last: &Part{Name: "/a.xml", Relationships: []*Relationship{rel}}, rnd: fakeRand()}, false},
		{"base2", &Writer{p: newPackage(), w: zip.NewWriter(nil), last: &Part{Name: "/b/a.xml", Relationships: []*Relationship{rel}}, rnd: fakeRand()}, false},
		{"hasSome", w, false},
		{"duplicated", &Writer{w: zip.NewWriter(nil), last: &Part{Name: "/a.xml", Relationships: []*Relationship{rel, rel}}, rnd: fakeRand()}, true},
		{"invalidRelation", &Writer{w: zip.NewWriter(nil), last: &Part{Name: "/a.xml", Relationships: []*Relationship{{}}}, rnd: fakeRand()}, true},
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
