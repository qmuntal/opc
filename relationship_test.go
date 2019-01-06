package gopc

import (
	"bytes"
	"encoding/xml"
	"reflect"
	"testing"
)

func TestRelationship_writeToXML(t *testing.T) {
	tests := []struct {
		name    string
		r       *Relationship
		want    string
		wantErr bool
	}{
		{"xmlWriter", &Relationship{"fakeId", "fakeType", "", "fakeTarget", ModeInternal}, `<Relationship Id="fakeId" Type="fakeType" Target="/fakeTarget"></Relationship>`, false},
		{"emptyURI", &Relationship{"fakeId", "fakeType", "", "", ModeInternal}, `<Relationship Id="fakeId" Type="fakeType" Target="/"></Relationship>`, false},
		{"externalMode", &Relationship{"fakeId", "fakeType", "", "fakeTarget", ModeExternal}, `<Relationship Id="fakeId" Type="fakeType" Target="fakeTarget" TargetMode="External"></Relationship>`, false},
		{"base", &Relationship{"fakeId", "fakeType", "", "/fakeTarget", ModeInternal}, `<Relationship Id="fakeId" Type="fakeType" Target="/fakeTarget"></Relationship>`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buff := bytes.NewBufferString("")
			encoder := xml.NewEncoder(buff)
			if err := tt.r.writeToXML(encoder); (err != nil) != tt.wantErr {
				t.Errorf("Relationship.writeToXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := buff.String()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Relationship.writeToXML() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
		{"new", &Relationship{"fakeId", "fakeType", "", "fakeTarget", ModeExternal}, false},
		{"abs", &Relationship{"fakeId", "fakeType", "", "http://a.com/b", ModeExternal}, false},
		{"internalRelRel", &Relationship{"fakeId", "fakeType", "/", "/_rels/.rels", ModeInternal}, true},
		{"internalRelNoSource", &Relationship{"fakeId", "fakeType", "", "/fakeTarget", ModeInternal}, true},
		{"invalidTarget2", &Relationship{"fakeId", "fakeType", "", "  ", ModeInternal}, true},
		{"invalid", &Relationship{"fakeId", "fakeType", "", "://a.com/b", ModeExternal}, true},
		{"invalidID", &Relationship{"  ", "fakeType", "", "http://a.com/b", ModeInternal}, true},
		{"invalidAbsTarget", &Relationship{"fakeId", "fakeType", "", "http://a.com/b", ModeInternal}, true},
		{"invalidTarget", &Relationship{"fakeId", "fakeType", "", "", ModeInternal}, true},
		{"invalidRel1", &Relationship{"fakeId", "", "", "fakeTarget", ModeInternal}, true},
		{"invalidRel2", &Relationship{"fakeId", " ", "", "fakeTarget", ModeInternal}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Relationship.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
