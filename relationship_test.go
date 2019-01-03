package gopc

import (
	"bytes"
	"encoding/xml"
	"reflect"
	"testing"
)

func TestRelationship_ID(t *testing.T) {
	tests := []struct {
		name string
		r    *Relationship
		want string
	}{
		{"id", &Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}, "fakeId"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.ID(); got != tt.want {
				t.Errorf("Relationship.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelationship_Type(t *testing.T) {
	tests := []struct {
		name string
		r    *Relationship
		want string
	}{
		{"id", &Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}, "fakeType"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Type(); got != tt.want {
				t.Errorf("Relationship.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelationship_TargetURI(t *testing.T) {
	tests := []struct {
		name string
		r    *Relationship
		want string
	}{
		{"id", &Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}, "fakeTarget"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.TargetURI(); got != tt.want {
				t.Errorf("Relationship.TargetURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelationship_writeToXML(t *testing.T) {
	tests := []struct {
		name    string
		r       *Relationship
		want    string
		wantErr bool
	}{
		{"xmlWriter", &Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}, `<Relationship Id="fakeId" Type="fakeType" Target="/fakeTarget"></Relationship>`, false},
		{"emptyURI", &Relationship{"fakeId", "fakeType", "", ModeInternal}, `<Relationship Id="fakeId" Type="fakeType" Target="/"></Relationship>`, false},
		{"externalMode", &Relationship{"fakeId", "fakeType", "fakeTarget", ModeExternal}, `<Relationship Id="fakeId" Type="fakeType" Target="fakeTarget" TargetMode="External"></Relationship>`, false},
		{"base", &Relationship{"fakeId", "fakeType", "/fakeTarget", ModeInternal}, `<Relationship Id="fakeId" Type="fakeType" Target="/fakeTarget"></Relationship>`, false},
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
		want    *Relationship
		wantErr bool
	}{
		{"invalidTarget2", args{"fakeId", "fakeType", "  ", ModeInternal}, nil, true},
		{"new", args{"fakeId", "fakeType", "fakeTarget", ModeExternal}, &Relationship{"fakeId", "fakeType", "fakeTarget", ModeExternal}, false},
		{"absExternañ", args{"fakeId", "fakeType", "http://a.com/b", ModeExternal}, &Relationship{"fakeId", "fakeType", "http://a.com/b", ModeExternal}, false},
		{"invalid", args{"fakeId", "fakeType", "://a.com/b", ModeExternal}, nil, true},
		{"invalidID", args{"  ", "fakeType", "http://a.com/b", ModeInternal}, nil, true},
		{"invalidAbsTarget", args{"fakeId", "fakeType", "http://a.com/b", ModeInternal}, nil, true},
		{"invalidTarget", args{"fakeId", "fakeType", "", ModeInternal}, nil, true},
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

func Test_relator_CreateRelationship(t *testing.T) {
	type args struct {
		id         string
		relType    string
		targetURI  string
		targetMode TargetMode
	}
	tests := []struct {
		name    string
		r       *relator
		args    args
		want    *Relationship
		wantErr bool
	}{
		{"duplicatedID", &relator{map[string]*Relationship{"/a": new(Relationship)}}, args{"/a", "http://a.com", "/a.xml", ModeInternal}, nil, true},
		{"newRelErr", &relator{map[string]*Relationship{}}, args{"", "http://a.com", "/a.xml", ModeInternal}, nil, true},
		{"base", &relator{map[string]*Relationship{}}, args{"/a", "http://a.com", "/a.xml", ModeInternal}, &Relationship{"/a", "http://a.com", "/a.xml", ModeInternal}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.CreateRelationship(tt.args.id, tt.args.relType, tt.args.targetURI, tt.args.targetMode)
			if (err != nil) != tt.wantErr {
				t.Errorf("relator.CreateRelationship() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("relator.CreateRelationship() = %v, want %v", got, tt.want)
			}
		})
	}
}
