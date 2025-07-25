package event

import (
	"bytes"
	"orly.dev/pkg/crypto/p256k"
	"orly.dev/pkg/interfaces/signer"
	"orly.dev/pkg/utils/chk"
	"orly.dev/pkg/utils/errorf"
	"orly.dev/pkg/utils/log"
)

// Sign the event using the signer.I. Uses github.com/bitcoin-core/secp256k1 if
// available for much faster signatures.
//
// Note that this only populates the Pubkey, ID and Sig. The caller must
// set the CreatedAt timestamp as intended.
func (ev *E) Sign(keys signer.I) (err error) {
	ev.Pubkey = keys.Pub()
	ev.ID = ev.GetIDBytes()
	if ev.Sig, err = keys.Sign(ev.ID); chk.E(err) {
		return
	}
	return
}

// Verify an event is signed by the pubkey it contains. Uses
// github.com/bitcoin-core/secp256k1 if available for faster verification.
func (ev *E) Verify() (valid bool, err error) {
	keys := p256k.Signer{}
	if err = keys.InitPub(ev.Pubkey); chk.E(err) {
		return
	}
	if valid, err = keys.Verify(ev.ID, ev.Sig); chk.T(err) {
		// check that this isn't because of a bogus ID
		id := ev.GetIDBytes()
		if !bytes.Equal(id, ev.ID) {
			log.E.Ln("event ID incorrect")
			ev.ID = id
			err = nil
			if valid, err = keys.Verify(ev.ID, ev.Sig); chk.E(err) {
				return
			}
			err = errorf.W("event ID incorrect but signature is valid on correct ID")
		}
		return
	}
	return
}
