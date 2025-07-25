// Package eventenvelope is a codec for the event Submission request EVENT envelope
// (client) and event Result (to a REQ) from a relay.
package eventenvelope

import (
	"io"
	"orly.dev/pkg/encoders/envelopes"
	"orly.dev/pkg/encoders/event"
	"orly.dev/pkg/encoders/subscription"
	"orly.dev/pkg/interfaces/codec"
	"orly.dev/pkg/utils/chk"
	"orly.dev/pkg/utils/errorf"
)

// L is the label associated with this type of codec.Envelope.
const L = "EVENT"

// Submission is a request from a client for a realy to store an event.
type Submission struct {
	*event.E
}

var _ codec.Envelope = (*Submission)(nil)

// NewSubmission creates an empty new eventenvelope.Submission.
func NewSubmission() *Submission { return &Submission{E: &event.E{}} }

// NewSubmissionWith creates a new eventenvelope.Submission with a provided event.E.
func NewSubmissionWith(ev *event.E) *Submission { return &Submission{E: ev} }

// Label returns the label of a event eventenvelope.Submission envelope.
func (en *Submission) Label() string { return L }

func (en *Submission) Id() []byte { return en.E.ID }

// Write the Submission to a provided io.Writer.
func (en *Submission) Write(w io.Writer) (err error) {
	_, err = w.Write(en.Marshal(nil))
	return
}

// Marshal an event Submission envelope in minified JSON, appending to a
// provided destination slice.
func (en *Submission) Marshal(dst []byte) (b []byte) {
	var err error
	b = dst
	b = envelopes.Marshal(
		b, L,
		func(bst []byte) (o []byte) {
			o = bst
			o = en.E.Marshal(o)
			return
		},
	)
	_ = err
	return
}

// Unmarshal an event eventenvelope.Submission from minified JSON, returning the
// remainder after the end of the envelope.
func (en *Submission) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	en.E = event.New()
	if r, err = en.E.Unmarshal(r); chk.T(err) {
		return
	}
	r = en.E.Marshal(nil)
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

// ParseSubmission reads an event envelope Submission from minified JSON into a newly
// allocated eventenvelope.Submission.
func ParseSubmission(b []byte) (t *Submission, rem []byte, err error) {
	t = NewSubmission()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}

// Result is an event matching a filter associated with a subscription.
type Result struct {
	Subscription *subscription.Id
	Event        *event.E
}

var _ codec.Envelope = (*Result)(nil)

// NewResult creates a new empty eventenvelope.Result.
func NewResult() *Result { return &Result{} }

// NewResultWith creates a new eventenvelope.Result with a provided
// subscription.Id string and event.E.
func NewResultWith[V string | []byte](s V, ev *event.E) (
	res *Result, err error,
) {
	if len(s) < 0 || len(s) > 64 {
		err = errorf.E("subscription id must be length > 0 and <= 64")
		return
	}
	return &Result{subscription.MustNew(s), ev}, nil
}

func (en *Result) Id() []byte { return en.Event.ID }

// Label returns the label of a event eventenvelope.Result envelope.
func (en *Result) Label() string { return L }

// Write the eventenvelope.Result to a provided io.Writer.
func (en *Result) Write(w io.Writer) (err error) {
	_, err = w.Write(en.Marshal(nil))
	return
}

// Marshal an eventenvelope.Result envelope in minified JSON, appending to a
// provided destination slice.
func (en *Result) Marshal(dst []byte) (b []byte) {
	var err error
	b = dst
	b = envelopes.Marshal(
		b, L,
		func(bst []byte) (o []byte) {
			o = bst
			o = en.Subscription.Marshal(o)
			o = append(o, ',')
			o = en.Event.Marshal(o)
			return
		},
	)
	_ = err
	return
}

// Unmarshal an event Result envelope from minified JSON, returning the
// remainder after the end of the envelope.
func (en *Result) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	if en.Subscription, err = subscription.NewId([]byte{0}); chk.E(err) {
		return
	}
	if r, err = en.Subscription.Unmarshal(r); chk.E(err) {
		return
	}
	en.Event = event.New()
	if r, err = en.Event.Unmarshal(r); err != nil {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

// ParseResult allocates a new eventenvelope.Result and unmarshalls an EVENT
// envelope into it.
func ParseResult(b []byte) (t *Result, rem []byte, err error) {
	t = NewResult()
	if rem, err = t.Unmarshal(b); err != nil {
		return
	}
	return
}
