package ports

import "errors"

// ErrHashMismatch is returned by Hasher.Compare when the plain value does not match the hash.
var ErrHashMismatch = errors.New("hash mismatch")

// Hasher hashes and checks passwords.
type Hasher interface {
	// Compare returns ErrHashMismatch when literal doesn't match hashed.
	Compare(hashed string, literal string) error
	Hash(literal string) (string, error)
}
