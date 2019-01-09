package gopc

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
	tests := []struct {
		name    string
		r       *Relationship
		wantErr bool
	}{
		{"new", &Relationship{ID: "fakeId", RelType: "fakeType", sourceURI: "", TargetURI: "fakeTarget", TargetMode: ModeExternal}, false},
		{"abs", &Relationship{ID: "fakeId", RelType: "fakeType", sourceURI: "", TargetURI: "http://a.com/b", TargetMode: ModeExternal}, false},
		{"internalRelRel", &Relationship{ID: "fakeId", RelType: "fakeType", sourceURI: "/", TargetURI: "/_rels/.rels", TargetMode: ModeInternal}, true},
		{"internalRelNoSource", &Relationship{ID: "fakeId", RelType: "fakeType", sourceURI: "", TargetURI: "/fakeTarget", TargetMode: ModeInternal}, true},
		{"invalidTarget2", &Relationship{ID: "fakeId", RelType: "fakeType", sourceURI: "", TargetURI: "  ", TargetMode: ModeInternal}, true},
		{"invalid", &Relationship{ID: "fakeId", RelType: "fakeType", sourceURI: "", TargetURI: "://a.com/b", TargetMode: ModeExternal}, true},
		{"invalidID", &Relationship{ID: "  ", RelType: "fakeType", sourceURI: "", TargetURI: "http://a.com/b", TargetMode: ModeInternal}, true},
		{"invalidAbsTarget", &Relationship{ID: "fakeId", RelType: "fakeType", sourceURI: "", TargetURI: "http://a.com/b", TargetMode: ModeInternal}, true},
		{"invalidTarget", &Relationship{ID: "fakeId", RelType: "fakeType", sourceURI: "", TargetURI: "", TargetMode: ModeInternal}, true},
		{"invalidRel1", &Relationship{ID: "fakeId", RelType: "", sourceURI: "", TargetURI: "fakeTarget", TargetMode: ModeInternal}, true},
		{"invalidRel2", &Relationship{ID: "fakeId", RelType: " ", sourceURI: "", TargetURI: "fakeTarget", TargetMode: ModeInternal}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.validate(); (err != nil) != tt.wantErr {
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
		{"base", args{[]*Relationship{&Relationship{ID: "fakeId", RelType: "asd", sourceURI: "", TargetURI: "fakeTarget", TargetMode: ModeInternal}}}, expectedsolution(), false},
		{"base2", args{[]*Relationship{&Relationship{ID: "fakeId", RelType: "asd", sourceURI: "", TargetURI: "fakeTarget", TargetMode: ModeExternal}}}, expectedsolution2(), false},
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
	return `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="fakeId" Type="asd" Target="/fakeTarget"></Relationship></Relationships>`
}

func expectedsolution2() string {
	return `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="fakeId" Type="asd" Target="fakeTarget" TargetMode="External"></Relationship></Relationships>`
}
