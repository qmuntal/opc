package opc

import (
	"testing"
)

func TestError_Code(t *testing.T) {
	tests := []struct {
		name string
		e    *Error
		want int
	}{
		{"empty", new(Error), 0},
		{"base", newError(1, "base"), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Code(); got != tt.want {
				t.Errorf("Error.Code() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_PartName(t *testing.T) {
	tests := []struct {
		name string
		e    *Error
		want string
	}{
		{"empty", new(Error), ""},
		{"base", &Error{partName: "base"}, "base"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.PartName(); got != tt.want {
				t.Errorf("Error.PartName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name      string
		e         *Error
		want      string
		wantPanic bool
	}{
		{"base", &Error{101, "/doc.xml", ""}, "opc: /doc.xml: a part name shall not be empty", false},
		{"panic", &Error{0, "/doc.xml", ""}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantPanic {
						t.Errorf("Error.Error() want panic")
					}
				}
			}()
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("Error.Error() = %v, want %v", got, tt.want)
				return
			}
			if tt.wantPanic {
				t.Error("Error.Error() want error")
			}
		})
	}
}

func TestError_RelationshipID(t *testing.T) {
	tests := []struct {
		name string
		e    *Error
		want string
	}{
		{"empty", new(Error), ""},
		{"base", newErrorRelationship(101, "/doc.xml", "21211"), "21211"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.RelationshipID(); got != tt.want {
				t.Errorf("Error.RelationshipID() = %v, want %v", got, tt.want)
			}
		})
	}
}
