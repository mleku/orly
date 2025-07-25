package event

import (
	"orly.dev/pkg/encoders/hex"
	"orly.dev/pkg/encoders/json"
	"orly.dev/pkg/encoders/kind"
	"orly.dev/pkg/encoders/tags"
	text2 "orly.dev/pkg/encoders/text"
	"orly.dev/pkg/encoders/timestamp"
	"orly.dev/pkg/interfaces/codec"
	"orly.dev/pkg/utils/chk"
	"orly.dev/pkg/utils/errorf"
	"orly.dev/pkg/utils/log"
	"reflect"
)

// ToCanonical converts the event to the canonical encoding used to derive the
// event ID.
func (ev *E) ToCanonical(dst []byte) (b []byte) {
	b = dst
	b = append(b, "[0,\""...)
	b = hex.EncAppend(b, ev.Pubkey)
	b = append(b, "\","...)
	b = ev.CreatedAt.Marshal(b)
	b = append(b, ',')
	b = ev.Kind.Marshal(b)
	b = append(b, ',')
	b = ev.Tags.Marshal(b)
	b = append(b, ',')
	b = text2.AppendQuote(b, ev.Content, text2.NostrEscape)
	b = append(b, ']')
	return
}

// GetIDBytes returns the raw SHA256 hash of the canonical form of an event.E.
func (ev *E) GetIDBytes() []byte { return Hash(ev.ToCanonical(nil)) }

// NewCanonical builds a new canonical encoder.
func NewCanonical() (a *json.Array) {
	a = &json.Array{
		V: []codec.JSON{
			&json.Unsigned{}, // 0
			&json.Hex{},      // pubkey
			&timestamp.T{},   // created_at
			&kind.T{},        // kind
			&tags.T{},        // tags
			&json.String{},   // content
		},
	}
	return
}

// this is an absolute minimum length canonical encoded event
var minimal = len(`[0,"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",1733739427,0,[],""]`)

// FromCanonical reverses the process of creating the canonical encoding, note
// that the signature is missing in this form. Allocate an event.E before
// calling this.
func (ev *E) FromCanonical(b []byte) (rem []byte, err error) {
	if len(b) < minimal {
		err = errorf.E(
			"event is too short to be a canonical event, require at least %d got %d",
			minimal, len(b),
		)
		return
	}
	rem = b
	id := Hash(rem)
	c := NewCanonical()
	if rem, err = c.Unmarshal(rem); chk.E(err) {
		log.I.F("%s", b)
		return
	}
	// unwrap the array
	x := (*c).V
	if v, ok := x[0].(*json.Unsigned); !ok {
		err = errorf.E(
			"did not decode expected type in first field of canonical event %v %v",
			reflect.TypeOf(x[0]), x[0],
		)
		return
	} else {
		if v.V != 0 {
			err = errorf.E(
				"unexpected value %d in first field of canonical event, expect 0",
				v.V,
			)
			return
		}
	}
	// create the event, use the ID hash to populate the ID
	ev.ID = id
	// unwrap the pubkey
	if v, ok := x[1].(*json.Hex); !ok {
		err = errorf.E(
			"failed to decode pubkey from canonical form of event %s", b,
		)
		return
	} else {
		ev.Pubkey = v.V
	}
	// populate the timestamp field
	if v, ok := x[2].(*timestamp.T); !ok {
		err = errorf.E(
			"did not decode expected type in third (created_at) field of canonical event %v %v",
			reflect.TypeOf(x[0]), x[0],
		)
	} else {
		ev.CreatedAt = v
	}
	// populate the kind field
	if v, ok := x[3].(*kind.T); !ok {
		err = errorf.E(
			"did not decode expected type in fourth (kind) field of canonical event %v %v",
			reflect.TypeOf(x[0]), x[0],
		)
	} else {
		ev.Kind = v
	}
	// populate the tags field
	if v, ok := x[4].(*tags.T); !ok {
		err = errorf.E(
			"did not decode expected type in fifth (tags) field of canonical event %v %v",
			reflect.TypeOf(x[0]), x[0],
		)
	} else {
		ev.Tags = v
	}
	// populate the content field
	if v, ok := x[5].(*json.String); !ok {
		err = errorf.E(
			"did not decode expected type in sixth (content) field of canonical event %v %v",
			reflect.TypeOf(x[0]), x[0],
		)
	} else {
		ev.Content = v.V
	}
	return
}
