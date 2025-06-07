package utils

import (
	"testing"
)

func TestDecodeCredentialRawID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantBytes []byte
		wantErr   bool
	}{
		{
			name:      "RawURLEncoding 無 padding",
			input:     "aGVsbG8td29ybGQ", // "hello-world" => raw URL encoded
			wantBytes: []byte("hello-world"),
			wantErr:   false,
		},
		{
			name:      "URLEncoding 有 padding",
			input:     "aGVsbG8td29ybGQ=", // same string with =
			wantBytes: []byte("hello-world"),
			wantErr:   false,
		},
		{
			name:    "無效字元",
			input:   "!!!???",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecodeCredentialRawID(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("預期錯誤但實際沒有")
				}
			} else {
				if err != nil {
					t.Errorf("不應該有錯: %v", err)
				}
				if string(got) != string(tc.wantBytes) {
					t.Errorf("結果錯誤，got=%v, want=%v", got, tc.wantBytes)
				}
			}
		})
	}
}