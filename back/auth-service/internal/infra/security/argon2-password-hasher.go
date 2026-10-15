package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
	"golang.org/x/crypto/argon2"
)

var _ ports.IPasswordHasher = (*Argon2PasswordHasher)(nil)

var ErrInvalidHash = errors.New("invalid argon2id hash")

const (
	saltLength = 16
	keyLength  = 32
)

// Argon2PasswordHasher stores passwords as argon2id hashes in the PHC string format, so the
// parameters can be raised without invalidating existing hashes.
type Argon2PasswordHasher struct {
	memory  uint32
	time    uint32
	threads uint8
}

// NewArgon2PasswordHasher uses the minimum parameters recommended by OWASP.
func NewArgon2PasswordHasher() *Argon2PasswordHasher {
	return &Argon2PasswordHasher{memory: 19 * 1024, time: 2, threads: 1}
}

func (a *Argon2PasswordHasher) Hash(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, a.time, a.memory, a.threads, keyLength)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, a.memory, a.time, a.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func (a *Argon2PasswordHasher) Verify(password, hash string) (bool, error) {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false, ErrInvalidHash
	}

	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrInvalidHash
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(key) == 0 {
		return false, ErrInvalidHash
	}

	candidate := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(key)))
	return subtle.ConstantTimeCompare(key, candidate) == 1, nil
}
