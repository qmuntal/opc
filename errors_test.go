package gopc

import "testing"

func TestError_Code(t *testing.T) {
	tests := []struct {
		name string
		e    *Error
		want int
	}{
		{"empty", new(Error), 0},
		{"base", &Error{1, "base"}, 1},
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
		name string
		e    *Error
		want string
	}{
		{"base", &Error{101, "/doc.xml"}, "OPC: Part='/doc.xml' | Reason='a part name shall not be empty'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("Error.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
