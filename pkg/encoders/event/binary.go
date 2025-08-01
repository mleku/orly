package event

import (
	"io"
	"orly.dev/pkg/crypto/ec/schnorr"
	"orly.dev/pkg/encoders/kind"
	"orly.dev/pkg/encoders/tag"
	"orly.dev/pkg/encoders/tags"
	"orly.dev/pkg/encoders/timestamp"
	"orly.dev/pkg/encoders/varint"
	"orly.dev/pkg/utils/chk"
)

// MarshalBinary writes a binary encoding of an event.
//
// [ 32 bytes ID ]
// [ 32 bytes Pubkey ]
// [ varint CreatedAt ]
// [ 2 bytes Kind ]
// [ varint Tags length ]
//
//	[ varint tag length ]
//	  [ varint tag element length ]
//	  [ tag element data ]
//	...
//
// [ varint Content length ]
// [ 64 bytes Sig ]
func (ev *E) MarshalBinary(w io.Writer) {
	_, _ = w.Write(ev.ID)
	_, _ = w.Write(ev.Pubkey)
	varint.Encode(w, uint64(ev.CreatedAt.V))
	varint.Encode(w, uint64(ev.Kind.K))
	varint.Encode(w, uint64(ev.Tags.Len()))
	for _, x := range ev.Tags.ToSliceOfTags() {
		varint.Encode(w, uint64(x.Len()))
		for _, y := range x.ToSliceOfBytes() {
			varint.Encode(w, uint64(len(y)))
			_, _ = w.Write(y)
		}
	}
	varint.Encode(w, uint64(len(ev.Content)))
	_, _ = w.Write(ev.Content)
	_, _ = w.Write(ev.Sig)
	return
}

func (ev *E) UnmarshalBinary(r io.Reader) (err error) {
	ev.ID = make([]byte, 32)
	if _, err = r.Read(ev.ID); chk.E(err) {
		return
	}
	ev.Pubkey = make([]byte, 32)
	if _, err = r.Read(ev.Pubkey); chk.E(err) {
		return
	}
	var ca uint64
	if ca, err = varint.Decode(r); chk.E(err) {
		return
	}
	ev.CreatedAt = timestamp.New(int64(ca))
	var k uint64
	if k, err = varint.Decode(r); chk.E(err) {
		return
	}
	ev.Kind = kind.New(k)
	var nTags uint64
	if nTags, err = varint.Decode(r); chk.E(err) {
		return
	}
	ev.Tags = tags.NewWithCap(int(nTags))
	for range nTags {
		var nField uint64
		if nField, err = varint.Decode(r); chk.E(err) {
			return
		}
		t := tag.NewWithCap(int(nField))
		for range nField {
			var lenField uint64
			if lenField, err = varint.Decode(r); chk.E(err) {
				return
			}
			field := make([]byte, lenField)
			if _, err = r.Read(field); chk.E(err) {
				return
			}
			t = t.Append(field)
		}
		ev.Tags.AppendTags(t)
	}
	var cLen uint64
	if cLen, err = varint.Decode(r); chk.E(err) {
		return
	}
	ev.Content = make([]byte, cLen)
	if _, err = r.Read(ev.Content); chk.E(err) {
		return
	}
	ev.Sig = make([]byte, schnorr.SignatureSize)
	if _, err = r.Read(ev.Sig); chk.E(err) {
		return
	}
	return
}
