package opc

import (
	"bytes"
	"testing"
)

func Test_isRelationshipURI(t *testing.T) {
	type args struct {
		uri string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{""}, false},
		{"withoutExtension", args{"/b/a.xml"}, false},
		{"withoutFolder", args{"/b/.rels"}, false},
		{"base", args{"/_rels/.rels"}, true},
		{"case1", args{"/_rels/.RELS"}, true},
		{"case2", args{"/_RELS/.rels"}, true},
		{"case3", args{"/_RELS/.RELS"}, true},
		{"nested", args{"XXX/_rels/YYY.rels"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRelationshipURI(tt.args.uri); got != tt.want {
				t.Errorf("IsRelationshipURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelationship_validate(t *testing.T) {
	type args struct {
		sourceURI string
	}
	tests := []struct {
		name    string
		r       *Relationship
		args    args
		wantErr bool
	}{
		{"relative", &Relationship{ID: "fakeId", Type: "fakeType", TargetURI: "./two/two.txt", TargetMode: ModeExternal}, args{"/one.txt"}, false},
		{"new", &Relationship{ID: "fakeId", Type: "fakeType", TargetURI: "fakeTarget", TargetMode: ModeExternal}, args{""}, false},
		{"abs", &Relationship{ID: "fakeId", Type: "fakeType", TargetURI: "http://a.com/b", TargetMode: ModeExternal}, args{""}, false},
		{"internalRelRel", &Relationship{ID: "fakeId", Type: "fakeType", TargetURI: "/_rels/.rels", TargetMode: ModeInternal}, args{"/"}, true},
		{"internalRelNoSource", &Relationship{ID: "fakeId", Type: "fakeType", TargetURI: "/fakeTarget", TargetMode: ModeInternal}, args{""}, true},
		{"invalidTarget2", &Relationship{ID: "fakeId", Type: "fakeType", TargetURI: "  ", TargetMode: ModeInternal}, args{""}, true},
		{"invalid", &Relationship{ID: "fakeId", Type: "fakeType", TargetURI: "://a.com/b", TargetMode: ModeExternal}, args{""}, true},
		{"invalidID", &Relationship{ID: "  ", Type: "fakeType", TargetURI: "http://a.com/b", TargetMode: ModeInternal}, args{""}, true},
		{"invalidAbsTarget", &Relationship{ID: "fakeId", Type: "fakeType", TargetURI: "http://a.com/b", TargetMode: ModeInternal}, args{""}, true},
		{"invalidTarget", &Relationship{ID: "fakeId", Type: "fakeType", TargetURI: "", TargetMode: ModeInternal}, args{""}, true},
		{"invalidRel1", &Relationship{ID: "fakeId", Type: "", TargetURI: "fakeTarget", TargetMode: ModeInternal}, args{""}, true},
		{"invalidRel2", &Relationship{ID: "fakeId", Type: " ", TargetURI: "fakeTarget", TargetMode: ModeInternal}, args{""}, true},
		{"invalidRel2", &Relationship{ID: "fakeId", Type: "fakeType", TargetURI: "./fakeTarget", TargetMode: ModeInternal}, args{""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.validate(tt.args.sourceURI); (err != nil) != tt.wantErr {
				t.Errorf("Relationship.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_encodeRelationships(t *testing.T) {
	type args struct {
		rs []*Relationship
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{"base", args{[]*Relationship{{ID: "fakeId", Type: "asd", TargetURI: "fakeTarget", TargetMode: ModeInternal}}}, expectedsolution(), false},
		{"base2", args{[]*Relationship{{ID: "fakeId", Type: "asd", TargetURI: "fakeTarget", TargetMode: ModeExternal}}}, expectedsolution2(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := encodeRelationships(w, tt.args.rs); (err != nil) != tt.wantErr {
				t.Errorf("encodeRelationships() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("encodeRelationships() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func expectedsolution() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
    <Relationship Id="fakeId" Type="asd" Target="fakeTarget"></Relationship>
</Relationships>`
}

func expectedsolution2() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
    <Relationship Id="fakeId" Type="asd" Target="fakeTarget" TargetMode="External"></Relationship>
</Relationships>`
}
