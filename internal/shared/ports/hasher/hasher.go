package hasher

import "errors"

var ErrHashMismatch = errors.New("hash mismatch")

type Hasher interface {
	Compare(hashed string, literal string) error
	Hash(literal string) (string, error)
}
