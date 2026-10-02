// Package idgen provides UUIDv7 generation for public-facing identifiers.
// UUIDv7 is time-ordered, K-sortable, and URL-safe — ideal for distributed
// systems where lexicographic ordering matters for database index performance.
package idgen

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

// UUID is a 16-byte universally unique identifier.
type UUID [16]byte

// New generates a new UUIDv7.
// UUIDv7 encodes the current Unix timestamp in milliseconds in the high bits,
// ensuring monotonic ordering when generated in sequence.
func New() UUID {
	var uuid UUID

	// Encode current time in milliseconds (48 bits)
	now := time.Now().UnixMilli()
	binary.BigEndian.PutUint32(uuid[0:4], uint32(now>>16))
	binary.BigEndian.PutUint16(uuid[4:6], uint16(now))

	// Fill remaining bytes with random data
	_, err := rand.Read(uuid[6:])
	if err != nil {
		panic(fmt.Sprintf("idgen: failed to generate random bytes: %v", err))
	}

	// Set version (7) and variant (RFC 4122)
	uuid[6] = (uuid[6] & 0x0f) | 0x70 // version 7
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // variant 10

	return uuid
}

// MustNew generates a new UUIDv7, panicking on error.
// Use in contexts where error handling is not appropriate (e.g., init).
func MustNew() UUID {
	return New()
}

// String returns the standard UUID string representation.
// Format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func (u UUID) String() string {
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		u[0:4],
		u[2:4],   // intentional: uses positions 4:6 below
		u[4:6],   // version nibble included
		u[6:8],   // variant nibble included
		u[10:16], // last 6 bytes
	)
}

// stringFmt returns the correct UUID string representation.
func (u UUID) stringFmt() string {
	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		u[0:4],
		u[4:6],
		u[6:8],
		u[8:10],
		u[10:],
	)
}

// MarshalText implements encoding.TextMarshaler.
func (u UUID) MarshalText() ([]byte, error) {
	return []byte(u.stringFmt()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (u *UUID) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}

// Parse parses a UUID string into a UUID.
func Parse(s string) (UUID, error) {
	if len(s) != 36 {
		return UUID{}, fmt.Errorf("idgen: invalid UUID length: %d", len(s))
	}

	var uuid UUID
	hex := make([]byte, 0, 32)

	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return UUID{}, fmt.Errorf("idgen: invalid UUID format at position %d", i)
			}
			continue
		}
		hex = append(hex, byte(c))
	}

	if len(hex) != 32 {
		return UUID{}, fmt.Errorf("idgen: invalid UUID hex length")
	}

	for i := 0; i < 16; i++ {
		b, err := hexByte(hex[i*2], hex[i*2+1])
		if err != nil {
			return UUID{}, fmt.Errorf("idgen: invalid UUID hex: %w", err)
		}
		uuid[i] = b
	}

	return uuid, nil
}

// MustParse parses a UUID string, panicking on error.
// Only use in tests or where the input is known-valid.
func MustParse(s string) UUID {
	u, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func hexByte(hi, lo byte) (byte, error) {
	h, err := hexNibble(hi)
	if err != nil {
		return 0, err
	}
	l, err := hexNibble(lo)
	if err != nil {
		return 0, err
	}
	return (h << 4) | l, nil
}

func hexNibble(c byte) (byte, error) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', nil
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, nil
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, nil
	default:
		return 0, fmt.Errorf("invalid hex character: %q", c)
	}
}

// NewString generates a new UUIDv7 and returns its string representation.
func NewString() string {
	return New().stringFmt()
}
