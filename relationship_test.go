package gopc

import (
	"bytes"
	"encoding/xml"
	"reflect"
	"testing"
)

func Test_relationship_ID(t *testing.T) {
	tests := []struct {
		name string
		r    *relationship
		want string
	}{
		{"id", &relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}, "fakeId"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.ID(); got != tt.want {
				t.Errorf("relationship.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_relationship_Type(t *testing.T) {
	tests := []struct {
		name string
		r    *relationship
		want string
	}{
		{"id", &relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}, "fakeType"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Type(); got != tt.want {
				t.Errorf("relationship.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_relationship_TargetURI(t *testing.T) {
	tests := []struct {
		name string
		r    *relationship
		want string
	}{
		{"id", &relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}, "fakeTarget"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.TargetURI(); got != tt.want {
				t.Errorf("relationship.TargetURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_relationship_writeToXML(t *testing.T) {
	tests := []struct {
		name    string
		r       *relationship
		want    string
		wantErr bool
	}{
		{"xmlWriter", &relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}, `<relationship Id="fakeId" Type="fakeType" Target="/fakeTarget"></relationship>`, false},
		{"emptyURI", &relationship{"fakeId", "fakeType", "", ModeInternal}, `<relationship Id="fakeId" Type="fakeType" Target="/"></relationship>`, false},
		{"externalMode", &relationship{"fakeId", "fakeType", "fakeTarget", ModeExternal}, `<relationship Id="fakeId" Type="fakeType" Target="fakeTarget" TargetMode="External"></relationship>`, false},
		{"base", &relationship{"fakeId", "fakeType", "/fakeTarget", ModeInternal}, `<relationship Id="fakeId" Type="fakeType" Target="/fakeTarget"></relationship>`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buff := bytes.NewBufferString("")
			encoder := xml.NewEncoder(buff)
			if err := tt.r.writeToXML(encoder); (err != nil) != tt.wantErr {
				t.Errorf("relationship.writeToXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := buff.String()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("relationship.writeToXML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newRelationship(t *testing.T) {
	type args struct {
		id         string
		relType    string
		targetURI  string
		targetMode TargetMode
	}
	tests := []struct {
		name    string
		args    args
		want    *relationship
		wantErr bool
	}{
		{"new", args{"fakeId", "fakeType", "fakeTarget", ModeExternal}, &relationship{"fakeId", "fakeType", "fakeTarget", ModeExternal}, false},
		{"absExternañ", args{"fakeId", "fakeType", "http://a.com/b", ModeExternal}, &relationship{"fakeId", "fakeType", "http://a.com/b", ModeExternal}, false},
		{"absExternañ", args{"fakeId", "fakeType", "://a.com/b", ModeExternal}, nil, true},
		{"invalidAbsTarget", args{"fakeId", "fakeType", "http://a.com/b", ModeInternal}, nil, true},
		{"invalidTarget", args{"fakeId", "fakeType", "", ModeInternal}, nil, true},
		{"invalidTarget2", args{"fakeId", "fakeType", "  ", ModeInternal}, nil, true},
		{"invalidRel1", args{"fakeId", "", "fakeTarget", ModeInternal}, nil, true},
		{"invalidRel2", args{"fakeId", " ", "fakeTarget", ModeInternal}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newRelationship(tt.args.id, tt.args.relType, tt.args.targetURI, tt.args.targetMode)
			if (err != nil) != tt.wantErr {
				t.Errorf("newRelationship() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newRelationship() = %v, want %v", got, tt.want)
			}
		})
	}
}
