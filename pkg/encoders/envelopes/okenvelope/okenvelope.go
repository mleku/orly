// Package okenvelope is a codec for the OK message, which is an acknowledgement
// for an EVENT eventenvelope.Submission, containing true/false and if false a
// message with a machine readable error type as found in the messages package.
package okenvelope

import (
	"io"
	"orly.dev/pkg/crypto/sha256"
	"orly.dev/pkg/encoders/envelopes"
	"orly.dev/pkg/encoders/eventid"
	text2 "orly.dev/pkg/encoders/text"
	"orly.dev/pkg/interfaces/codec"
	"orly.dev/pkg/utils/chk"
	"orly.dev/pkg/utils/errorf"
	"orly.dev/pkg/utils/log"
)

// L is the label associated with this type of codec.Envelope.
const L = "OK"

// T is an OK envelope, used to signal acceptance or rejection, with a reason,
// to an eventenvelope.Submission.
type T struct {
	EventID *eventid.T
	OK      bool
	Reason  []byte
}

var _ codec.Envelope = (*T)(nil)

// New creates a new empty OK T.
func New() *T { return &T{} }

// NewFrom creates a new okenvelope.T with a string for the subscription.Id and
// the Reason.
func NewFrom[V string | []byte](eid V, ok bool, msg ...V) *T {
	var m []byte
	if len(msg) > 0 {
		m = []byte(msg[0])
	}
	if len(eid) != sha256.Size {
		log.W.F(
			"event ID unexpected length, expect %d got %d",
			len(eid), sha256.Size,
		)
	}
	return &T{EventID: eventid.NewWith(eid), OK: ok, Reason: m}
}

// Label returns the label of an okenvelope.T.
func (en *T) Label() string { return L }

// ReasonString returns the Reason in the form of a string.
func (en *T) ReasonString() string { return string(en.Reason) }

// Write the okenvelope.T to a provided io.Writer.
func (en *T) Write(w io.Writer) (err error) {
	_, err = w.Write(en.Marshal(nil))
	return
}

// Marshal a okenvelope.T from minified JSON, appending to a provided
// destination slice. Note that this ensures correct string escaping on the
// subscription.Id and Reason fields.
func (en *T) Marshal(dst []byte) (b []byte) {
	var err error
	_ = err
	b = dst
	b = envelopes.Marshal(
		b, L,
		func(bst []byte) (o []byte) {
			o = bst
			o = append(o, '"')
			o = en.EventID.ByteString(o)
			o = append(o, '"')
			o = append(o, ',')
			o = text2.MarshalBool(o, en.OK)
			o = append(o, ',')
			o = append(o, '"')
			o = text2.NostrEscape(o, en.Reason)
			o = append(o, '"')
			return
		},
	)
	return
}

// Unmarshal a okenvelope.T from minified JSON, returning the remainder after
// the end of the envelope. Note that this ensures the Reason and
// subscription.Id strings are correctly unescaped by NIP-01 escaping rules.
func (en *T) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	var idHex []byte
	if idHex, r, err = text2.UnmarshalHex(r); chk.E(err) {
		return
	}
	if len(idHex) != sha256.Size {
		err = errorf.E(
			"invalid size for ID, require %d got %d",
			len(idHex), sha256.Size,
		)
	}
	en.EventID = eventid.NewWith(idHex)
	if r, err = text2.Comma(r); chk.E(err) {
		return
	}
	if r, en.OK, err = text2.UnmarshalBool(r); chk.E(err) {
		return
	}
	if r, err = text2.Comma(r); chk.E(err) {
		return
	}
	if en.Reason, r, err = text2.UnmarshalQuoted(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

// Parse reads a OK envelope in minified JSON into a newly allocated
// okenvelope.T.
func Parse(b []byte) (t *T, rem []byte, err error) {
	t = New()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
