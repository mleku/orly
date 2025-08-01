package ints

import (
	"math"
	"orly.dev/pkg/utils/chk"
	"strconv"
	"testing"

	"lukechampine.com/frand"
)

func TestMarshalUnmarshal(t *testing.T) {
	b := make([]byte, 0, 8)
	var rem []byte
	var n *T
	var err error
	for _ = range 10000000 {
		n = New(uint64(frand.Intn(math.MaxInt64)))
		b = n.Marshal(b)
		m := New(0)
		if rem, err = m.Unmarshal(b); chk.E(err) {
			t.Fatal(err)
		}
		if n.N != m.N {
			t.Fatalf("failed to convert to int64 at %d %s %d", n.N, b, m.N)
		}
		if len(rem) > 0 {
			t.Fatalf("leftover bytes after converting back: '%s'", rem)
		}
		b = b[:0]
	}
}

func BenchmarkByteStringToInt64(bb *testing.B) {
	b := make([]byte, 0, 19)
	var i int
	const nTests = 10000000
	testInts := make([]*T, nTests)
	for i = range nTests {
		testInts[i] = New(frand.Intn(math.MaxInt64))
	}
	bb.Run(
		"Marshal", func(bb *testing.B) {
			bb.ReportAllocs()
			for i = 0; i < bb.N; i++ {
				n := testInts[i%10000]
				b = n.Marshal(b)
				b = b[:0]
			}
		},
	)
	bb.Run(
		"Itoa", func(bb *testing.B) {
			bb.ReportAllocs()
			var s string
			for i = 0; i < bb.N; i++ {
				n := testInts[i%10000]
				s = strconv.Itoa(int(n.N))
				_ = s
			}
		},
	)
	bb.Run(
		"MarshalUnmarshal", func(bb *testing.B) {
			bb.ReportAllocs()
			m := New(0)
			for i = 0; i < bb.N; i++ {
				n := testInts[i%10000]
				b = m.Marshal(b)
				_, _ = n.Unmarshal(b)
				b = b[:0]
			}
		},
	)
	bb.Run(
		"ItoaAtoi", func(bb *testing.B) {
			bb.ReportAllocs()
			var s string
			for i = 0; i < bb.N; i++ {
				n := testInts[i%10000]
				s = strconv.Itoa(int(n.N))
				_, _ = strconv.Atoi(s)
			}
		},
	)
}
