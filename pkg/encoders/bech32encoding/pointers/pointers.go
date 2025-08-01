// Package pointers is a set of basic nip-19 data types for generating bech32
// encoded nostr entities.
package pointers

import (
	"orly.dev/pkg/encoders/eventid"
	"orly.dev/pkg/encoders/kind"
)

// Profile pointer is a combination of pubkey and relay list.
type Profile struct {
	PublicKey []byte   `json:"pubkey"`
	Relays    [][]byte `json:"relays,omitempty"`
}

// Event pointer is the combination of an event ID, relay hints, author, pubkey,
// and kind.
type Event struct {
	ID     *eventid.T `json:"id"`
	Relays [][]byte   `json:"relays,omitempty"`
	Author []byte     `json:"author,omitempty"`
	Kind   *kind.T    `json:"kind,omitempty"`
}

// Entity is the combination of a pubkey, kind, arbitrary identifier, and relay
// hints.
type Entity struct {
	PublicKey  []byte   `json:"pubkey"`
	Kind       *kind.T  `json:"kind,omitempty"`
	Identifier []byte   `json:"identifier,omitempty"`
	Relays     [][]byte `json:"relays,omitempty"`
}
