package errors

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		e string
	}
	tests := []struct {
		name    string
		args    args
		wantIe  *IdefixError
		wantErr bool
	}{
		{"[200] OK", args{e: "[200] OK"}, &IdefixError{Code: 200, Message: "OK"}, false},
		{"[200] OK (Extra)", args{e: "[200] OK (Extra)"}, &IdefixError{Code: 200, Message: "OK", Extra: "Extra"}, false},
		{"[200] OK (abc (2))", args{e: "[200] OK (abc (2))"}, &IdefixError{Code: 200, Message: "OK", Extra: "abc (2)"}, false},
		{"[-300] BAD ERROR", args{e: "[-300] BAD ERROR"}, &IdefixError{Code: -300, Message: "BAD ERROR"}, false},
		{"[-300] BAD ERROR", args{e: "[-300] BAD ERROR (OH NO)"}, &IdefixError{Code: -300, Message: "BAD ERROR", Extra: "OH NO"}, false},
		{"[1] This is a really long error", args{e: "[1] This is a really long error"}, &IdefixError{Code: 1, Message: "This is a really long error"}, false},
		{"[202] Invalid data type", args{e: "[202] Invalid data type (1 error(s) decoding:\n\n* 'admins': source data must be an array or slice, got string)"}, &IdefixError{Code: 202, Message: "Invalid data type", Extra: "1 error(s) decoding:\n\n* 'admins': source data must be an array or slice, got string"}, false},
		{"-300] FAIL", args{e: "-300] FAIL"}, nil, true},
		{"[-300 FAIL", args{e: "[-300 FAIL"}, nil, true},
		{"-300 FAIL", args{e: "-300 FAIL"}, nil, true},
		{"FAIL", args{e: "FAIL"}, nil, true},
		{"[300]FAIL", args{e: "[300]FAIL"}, nil, true},
		{"-300] FAIL", args{e: "[-300] FAIL asd)"}, nil, true},
		{"-300] FAIL", args{e: "[-300] FAIL (asd"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIe, err := Parse(tt.args.e)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(gotIe, tt.wantIe) {
				t.Errorf("Parse() = %v, want %v", gotIe, tt.wantIe)
			}
		})
	}
}

func TestIdefixError_Error(t *testing.T) {
	type fields struct {
		Code    int
		Message string
		Extra   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"1", fields{100, "hola", "extra"}, "[100] hola (extra)"},
		{"2", fields{100, "hola", ""}, "[100] hola"},
		{"3", fields{100, "", ""}, "[100]"},
		{"4", fields{}, "[0]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ie := IdefixError{
				Code:    tt.fields.Code,
				Message: tt.fields.Message,
				Extra:   tt.fields.Extra,
			}
			if got := ie.Error(); got != tt.want {
				t.Errorf("IdefixError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
