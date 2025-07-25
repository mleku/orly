// Copyright (c) 2013-2021 The btcsuite developers
// Copyright (c) 2015-2021 The Decred developers

package btcec

import (
	"orly.dev/pkg/crypto/ec/secp256k1"
)

// Error identifies an error related to public key cryptography using a
// sec256k1 curve. It has full support for errors.Is and errors.As, so the
// caller can ascertain the specific reason for the error by checking the
// underlying error.
type Error = secp256k1.Error

// ErrorKind identifies a kind of error. It has full support for errors.Is and
// errors.As, so the caller can directly check against an error kind when
// determining the reason for an error.
type ErrorKind = secp256k1.ErrorKind

// makeError creates an secp256k1.Error given a set of arguments.
func makeError(kind ErrorKind, desc string) Error {
	return Error{Err: kind, Description: desc}
}
