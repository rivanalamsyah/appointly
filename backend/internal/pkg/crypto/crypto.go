// Package crypto provides secure cryptographic operations for the application.
// All password hashing uses Argon2id per OWASP recommendations.
// Token generation uses crypto/rand for cryptographically secure randomness.
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Params defines the parameters for Argon2id password hashing.
// These values should be tuned based on the server hardware — the goal is
// to make hashing take ~100ms on the target server.
type Argon2Params struct {
	Memory      uint32 // in KiB
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Params returns OWASP-recommended Argon2id parameters.
// Tune Memory and Iterations in production based on available hardware.
var DefaultArgon2Params = Argon2Params{
	Memory:      64 * 1024, // 64 MB
	Iterations:  3,
	Parallelism: 4,
	SaltLength:  16,
	KeyLength:   32,
}

// HashPassword hashes a plaintext password using Argon2id.
// Returns an encoded string in the format:
//
//	$argon2id$v=19$m=<mem>,t=<iter>,p=<parallel>$<salt>$<hash>
func HashPassword(password string, params Argon2Params) (string, error) {
	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("crypto: failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.Memory,
		params.Iterations,
		params.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encoded, nil
}

// HashPasswordDefault hashes a password with default parameters.
func HashPasswordDefault(password string) (string, error) {
	return HashPassword(password, DefaultArgon2Params)
}

// VerifyPassword compares a plaintext password against an Argon2id hash.
// Returns true if the password matches, false otherwise.
// Uses constant-time comparison to prevent timing attacks.
func VerifyPassword(password, encodedHash string) (bool, error) {
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, fmt.Errorf("crypto: failed to decode hash: %w", err)
	}

	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

// decodeHash parses an Argon2id encoded hash string.
func decodeHash(encodedHash string) (Argon2Params, []byte, []byte, error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return Argon2Params{}, nil, nil, fmt.Errorf("invalid hash format: expected 6 parts, got %d", len(vals))
	}

	if vals[1] != "argon2id" {
		return Argon2Params{}, nil, nil, fmt.Errorf("unsupported algorithm: %s", vals[1])
	}

	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("failed to parse version: %w", err)
	}
	if version != argon2.Version {
		return Argon2Params{}, nil, nil, fmt.Errorf("incompatible argon2 version: %d", version)
	}

	var params Argon2Params
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism)
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("failed to parse params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}
	params.SaltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}
	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil
}

// GenerateSecureToken generates a cryptographically secure random token
// of the specified byte length, returned as a URL-safe base64 string.
func GenerateSecureToken(byteLength int) (string, error) {
	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto: failed to generate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateSecureTokenHex generates a cryptographically secure random token
// and returns it as a hex string.
func GenerateSecureTokenHex(byteLength int) (string, error) {
	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto: failed to generate token: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}

// MustGenerateSecureToken generates a token, panicking on error.
// Only use in non-critical paths or tests.
func MustGenerateSecureToken(byteLength int) string {
	t, err := GenerateSecureToken(byteLength)
	if err != nil {
		panic(err)
	}
	return t
}

// ConstantTimeEqual performs constant-time string comparison.
func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// HashSHA256 computes a SHA-256 hex string of the input text.
func HashSHA256(data string) string {
	h := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", h)
}

