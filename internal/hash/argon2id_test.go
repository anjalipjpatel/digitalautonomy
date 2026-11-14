package hash

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{name: "non-empty password", password: "correct horse battery staple"},
		{name: "empty password", password: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			parts := strings.Split(encoded, "$")
			if len(parts) != 6 || parts[1] != "argon2id" {
				t.Fatalf("Hash() returned malformed hash %q", encoded)
			}
			if parts[2] != fmt.Sprintf("v=%d", argon2.Version) {
				t.Errorf("version = %q, want v=%d", parts[2], argon2.Version)
			}
			if parts[3] != fmt.Sprintf("m=%d,t=%d,p=%d", memory, time, threads) {
				t.Errorf("parameters = %q, want m=%d,t=%d,p=%d", parts[3], memory, time, threads)
			}

			salt, err := b64.DecodeString(parts[4])
			if err != nil {
				t.Fatalf("decoding salt: %v", err)
			}
			if len(salt) != saltLen {
				t.Errorf("salt length = %d, want %d", len(salt), saltLen)
			}

			key, err := b64.DecodeString(parts[5])
			if err != nil {
				t.Fatalf("decoding key: %v", err)
			}
			if len(key) != keyLen {
				t.Errorf("key length = %d, want %d", len(key), keyLen)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	encoded := testEncodedHash("correct password")
	tests := []struct {
		name        string
		password    string
		encoded     string
		wantValid   bool
		wantErr     bool
	}{
		{
			name:      "correct password",
			password: "correct password",
			encoded:  encoded,
			wantValid: true,
		},
		{
			name:      "incorrect password",
			password: "wrong password",
			encoded:  encoded,
			wantValid: false,
		},
		{
			name:    "empty encoded hash",
			encoded: "",
			wantErr: true,
		},
		{
			name:    "unsupported algorithm",
			encoded: strings.Replace(encoded, "argon2id", "argon2i", 1),
			wantErr: true,
		},
		{
			name:    "unsupported version",
			encoded: strings.Replace(encoded, fmt.Sprintf("v=%d", argon2.Version), "v=1", 1),
			wantErr: true,
		},
		{
			name:    "invalid salt encoding",
			encoded: replaceEncodedPart(encoded, 4, "not-base64"),
			wantErr: true,
		},
		{
			name:    "invalid hash encoding",
			encoded: replaceEncodedPart(encoded, 5, "not-base64"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Verify(tt.password, tt.encoded)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Verify() error = %v, wantErr %t", err, tt.wantErr)
			}
			if err == nil && got != tt.wantValid {
				t.Errorf("Verify() = %t, want %t", got, tt.wantValid)
			}
		})
	}
}

func testEncodedHash(password string) string {
	salt := []byte("0123456789abcdef")
	const testMemory = 32
	const testTime = 1
	const testThreads = 1
	key := argon2.IDKey([]byte(password), salt, testTime, testMemory, testThreads, keyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		testMemory,
		testTime,
		testThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
}

func replaceEncodedPart(encoded string, index int, replacement string) string {
	parts := strings.Split(encoded, "$")
	parts[index] = replacement
	return strings.Join(parts, "$")
}
