// Package event provides a codec for nostr events, for the wire format (with ID
// and signature), for the canonical form, that is hashed to generate the ID,
// and a fast binary form that uses io.Reader/io.Writer.
package event

import (
	"github.com/minio/sha256-simd"
	"lukechampine.com/frand"
	"orly.dev/pkg/crypto/ec/schnorr"
	"orly.dev/pkg/encoders/eventid"
	"orly.dev/pkg/encoders/hex"
	"orly.dev/pkg/encoders/kind"
	"orly.dev/pkg/encoders/tag"
	"orly.dev/pkg/encoders/tags"
	"orly.dev/pkg/encoders/text"
	"orly.dev/pkg/encoders/timestamp"
	"orly.dev/pkg/interfaces/signer"
	"orly.dev/pkg/utils/chk"
	"orly.dev/pkg/utils/errorf"
)

// E is the primary datatype of nostr. This is the form of the structure that
// defines its JSON string-based format.
type E struct {

	// ID is the SHA256 hash of the canonical encoding of the event in binary format
	ID []byte

	// Pubkey is the public key of the event creator in binary format
	Pubkey []byte

	// CreatedAt is the UNIX timestamp of the event according to the event
	// creator (never trust a timestamp!)
	CreatedAt *timestamp.T

	// Kind is the nostr protocol code for the type of event. See kind.T
	Kind *kind.T

	// Tags are a list of tags, which are a list of strings usually structured
	// as a 3-layer scheme indicating specific features of an event.
	Tags *tags.T

	// Content is an arbitrary string that can contain anything, but usually
	// conforming to a specification relating to the Kind and the Tags.
	Content []byte

	// Sig is the signature on the ID hash that validates as coming from the
	// Pubkey in binary format.
	Sig []byte
}

// S is an array of event.E that sorts in reverse chronological order.
type S []*E

// Len returns the length of the event.Es.
func (ev S) Len() int { return len(ev) }

// Less returns whether the first is newer than the second (larger unix
// timestamp).
func (ev S) Less(i, j int) bool { return ev[i].CreatedAt.I64() > ev[j].CreatedAt.I64() }

// Swap two indexes of the event.Es with each other.
func (ev S) Swap(i, j int) { ev[i], ev[j] = ev[j], ev[i] }

// C is a channel that carries event.E.
type C chan *E

// New makes a new event.E.
func New() (ev *E) { return &E{} }

// Serialize renders an event.E into minified JSON.
func (ev *E) Serialize() (b []byte) { return ev.Marshal(nil) }

// SerializeIndented renders an event.E into nicely readable whitespaced JSON.
func (ev *E) SerializeIndented() (b []byte) {
	return ev.MarshalWithWhitespace(nil, true)
}

// EventId returns the event.E ID as an eventid.T.
func (ev *E) EventId() (eid *eventid.T) {
	return eventid.NewWith(ev.ID)
}

// stringy/numbery functions for retarded other libraries

// IdString returns the event ID as a hex-encoded string.
func (ev *E) IdString() (s string) { return hex.Enc(ev.ID) }

func (ev *E) Id() []byte { return ev.ID }

// CreatedAtInt64 returns the created_at timestamp as a standard int64.
func (ev *E) CreatedAtInt64() (i int64) { return ev.CreatedAt.I64() }

// KindInt returns the kind as an int, as is often needed for JSON.
func (ev *E) KindInt() (i int) { return int(ev.Kind.K) }

// KindInt32 returns the kind as an int32, as is often needed for JSON.
func (ev *E) KindInt32() (i int32) { return int32(ev.Kind.K) }

// PubKeyString returns the pubkey as a hex-encoded string.
func (ev *E) PubKeyString() (s string) { return hex.Enc(ev.Pubkey) }

// SigString returns the signature as a hex-encoded string.
func (ev *E) SigString() (s string) { return hex.Enc(ev.Sig) }

// TagStrings returns the tags as a slice of slice of strings.
func (ev *E) TagStrings() (s [][]string) { return ev.Tags.ToStringsSlice() }

// ContentString returns the content field as a string.
func (ev *E) ContentString() (s string) { return string(ev.Content) }

// J is an event.E encoded in more basic types than used in this library.
type J struct {
	Id        string     `json:"id" doc:"event id (SHA256 hash of canonical form of event, 64 characters hex)"`
	Pubkey    string     `json:"pubkey" doc:"public key of author of event, required to verify signature (BIP-340 Schnorr public key, 64 characters hex)"`
	CreatedAt int64      `json:"created_at" doc:"unix timestamp of time when event was created"`
	Kind      int        `json:"kind" doc:"kind number of event"`
	Tags      [][]string `json:"tags" doc:"tags that add metadata to the event"`
	Content   string     `json:"content" doc:"content of event"`
	Sig       string     `json:"sig" doc:"signature of event (BIP-340 schnorr signature, 128 characters hex)"`
}

// ToEventJ converts an event.E into an event.J.
func (ev *E) ToEventJ() (j *J) {
	j = &J{}
	j.Id = ev.IdString()
	j.Pubkey = ev.PubKeyString()
	j.CreatedAt = ev.CreatedAt.I64()
	j.Kind = ev.KindInt()
	j.Content = ev.ContentString()
	j.Tags = ev.Tags.ToStringsSlice()
	j.Sig = ev.SigString()
	return
}

// IdFromString decodes an event ID and loads it into an event.E ID.
func (ev *E) IdFromString(s string) (err error) {
	ev.ID, err = hex.Dec(s)
	return
}

// CreatedAtFromInt64 encodes a unix timestamp into the CreatedAt field of an
// event.E.
func (ev *E) CreatedAtFromInt64(i int64) {
	ev.CreatedAt = timestamp.FromUnix(i)
	return
}

// KindFromInt32 encodes an int32 representation of a kind.T into an event.E.
func (ev *E) KindFromInt32(i int32) {
	ev.Kind = &kind.T{}
	ev.Kind.K = uint16(i)
	return
}

// KindFromInt encodes an int representation of a kind.T into an event.E.
func (ev *E) KindFromInt(i int) {
	ev.Kind = &kind.T{}
	ev.Kind.K = uint16(i)
	return
}

// PubKeyFromString decodes a hex-encoded string into the event.E Pubkey field.
func (ev *E) PubKeyFromString(s string) (err error) {
	if len(s) != 2*schnorr.PubKeyBytesLen {
		err = errorf.E(
			"invalid length public key hex, got %d require %d",
			len(s), 2*schnorr.PubKeyBytesLen,
		)
	}
	ev.Pubkey, err = hex.Dec(s)
	return
}

// SigFromString decodes a hex-encoded string into the event.E Sig field.
func (ev *E) SigFromString(s string) (err error) {
	if len(s) != 2*schnorr.SignatureSize {
		err = errorf.E(
			"invalid length signature hex, got %d require %d",
			len(s), 2*schnorr.SignatureSize,
		)
	}
	ev.Sig, err = hex.Dec(s)
	return
}

// TagsFromStrings converts a slice of slice of strings into tags.T for the
// event.E.
func (ev *E) TagsFromStrings(s ...[]string) {
	ev.Tags = tags.NewWithCap(len(s))
	var tgs []*tag.T
	for _, t := range s {
		tg := tag.New(t...)
		tgs = append(tgs, tg)
	}
	ev.Tags.AppendTags(tgs...)
	return
}

// ContentFromString imports a content string into the event.E Content field.
func (ev *E) ContentFromString(s string) {
	ev.Content = []byte(s)
	return
}

// ToEvent converts event.J format to the realy native form.
func (e J) ToEvent() (ev *E, err error) {
	ev = &E{}
	if err = ev.IdFromString(e.Id); chk.E(err) {
		return
	}
	ev.CreatedAtFromInt64(e.CreatedAt)
	ev.KindFromInt(e.Kind)
	if err = ev.PubKeyFromString(e.Pubkey); chk.E(err) {
		return
	}
	ev.TagsFromStrings(e.Tags...)
	ev.ContentFromString(e.Content)
	if err = ev.SigFromString(e.Sig); chk.E(err) {
		return
	}
	return
}

// Hash is a little helper generate a hash and return a slice instead of an
// array.
func Hash(in []byte) (out []byte) {
	h := sha256.Sum256(in)
	return h[:]
}

// GenerateRandomTextNoteEvent creates a generic event.E with random text
// content.
func GenerateRandomTextNoteEvent(sign signer.I, maxSize int) (
	ev *E,
	err error,
) {

	l := frand.Intn(maxSize * 6 / 8) // account for base64 expansion
	ev = &E{
		Pubkey:    sign.Pub(),
		Kind:      kind.TextNote,
		CreatedAt: timestamp.Now(),
		Content:   text.NostrEscape(nil, frand.Bytes(l)),
		Tags:      tags.New(),
	}
	if err = ev.Sign(sign); chk.E(err) {
		return
	}
	return
}
