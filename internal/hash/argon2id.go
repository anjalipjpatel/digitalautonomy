package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

/*
Argon2id has been chosen as the hashing mechanism for this project as it is seen to be the
'gold standard' for password hashing and key derivation due to its balance of a
combination of security traits, whilst still being efficient.

It works by taking a set of fixed parameters, and initalising a section of RAM with this password.
Then, it applies various compression functions to mix the memory and then extract the final hash.

Everything needed to verify a hash later is given by the output, for example:
	`$argon2id$v=19$m=65536,t=3,p=4$c29tZXNhbHQ$RdescudvJCsgt3ub+b+dWRWJTmaaJObG`
	where:
		• $argon2id: Identifies varient.
		• v=19: The version number (19 = current standard).
		• m=65536,t=3,p=4: Parameters used
		• c29tZXNhbHQ: The encoded salt (in Base64)
		• RdescudvJCsgt3ub+b+dWRWJTmaaJObG: The derived hash (in Base64).

The below implementation has been taken from https://mojoauth.com/security-guides/argon2-in-go.
*/

// RFC 9106 second recommended option: 64 MiB memory, 3 passes, 4 lanes.
const (
	memory  = 64 * 1024 // KiB
	time    = 3
	threads = 4
	keyLen  = 32
	saltLen = 16
)

var b64 = base64.RawStdEncoding

func Hash(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, time, threads, b64.EncodeToString(salt), b64.EncodeToString(key)), nil
}

func Verify(password, encoded string) (bool, error) {
	// The salts vary each time the passwords are hashed. To verify the hash,
	// we must read the salt and Argon2 parameters from the stored hash,
	// and hash the plaintext password with the same values to compare the resulting bytes.
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errors.New("not an argon2id hash")
	}
	var version int
	var m uint32
	var t uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false, errors.New("unsupported argon2 version")
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false, err
	}
	salt, err := b64.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	want, err := b64.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	got := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
