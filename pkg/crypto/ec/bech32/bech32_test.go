// Copyright (c) 2017-2020 The btcsuite developers
// Copyright (c) 2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bech32

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestBech32 tests whether decoding and re-encoding the valid BIP-173 test
// vectors works and if decoding invalid test vectors fails for the correct
// reason.
func TestBech32(t *testing.T) {
	tests := []struct {
		str           string
		expectedError error
	}{
		{"A12UEL5L", nil},
		{"a12uel5l", nil},
		{
			"an83characterlonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio1tt5tgs",
			nil,
		},
		{"abcdef1qpzry9x8gf2tvdw0s3jn54khce6mua7lmqqqxw", nil},
		{
			"11qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqc8247j",
			nil,
		},
		{"split1checkupstagehandshakeupstreamerranterredcaperred2y9e3w", nil},
		{
			"split1checkupstagehandshakeupstreamerranterredcaperred2y9e2w",
			ErrInvalidChecksum{
				"2y9e3w", "2y9e3wlc445v",
				"2y9e2w",
			},
		}, // invalid checksum
		{
			"s lit1checkupstagehandshakeupstreamerranterredcaperredp8hs2p",
			ErrInvalidCharacter(' '),
		}, // invalid character (space) in hrp
		{
			"spl\x7Ft1checkupstagehandshakeupstreamerranterredcaperred2y9e3w",
			ErrInvalidCharacter(127),
		}, // invalid character (DEL) in hrp
		{
			"split1cheo2y9e2w",
			ErrNonCharsetChar('o'),
		},                                            // invalid character (o) in data part
		{"split1a2y9w", ErrInvalidSeparatorIndex(5)}, // too short data part
		{
			"1checkupstagehandshakeupstreamerranterredcaperred2y9e3w",
			ErrInvalidSeparatorIndex(0),
		}, // empty hrp
		{
			"11qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqc8247j",
			ErrInvalidLength(91),
		}, // too long
		// Additional test vectors used in bitcoin core
		{" 1nwldj5", ErrInvalidCharacter(' ')},
		{"\x7f" + "1axkwrx", ErrInvalidCharacter(0x7f)},
		{"\x801eym55h", ErrInvalidCharacter(0x80)},
		{
			"an84characterslonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio1569pvx",
			ErrInvalidLength(91),
		},
		{"pzry9x0s0muk", ErrInvalidSeparatorIndex(-1)},
		{"1pzry9x0s0muk", ErrInvalidSeparatorIndex(0)},
		{"x1b4n0q5v", ErrNonCharsetChar(98)},
		{"li1dgmt3", ErrInvalidSeparatorIndex(2)},
		{"de1lg7wt\xff", ErrInvalidCharacter(0xff)},
		{"A1G7SGD8", ErrInvalidChecksum{"2uel5l", "2uel5llqfn3a", "g7sgd8"}},
		{"10a06t8", ErrInvalidLength(7)},
		{"1qzzfhee", ErrInvalidSeparatorIndex(0)},
		{"a12UEL5L", ErrMixedCase{}},
		{"A12uEL5L", ErrMixedCase{}},
	}
	for i, test := range tests {
		str := []byte(test.str)
		hrp, decoded, err := Decode([]byte(str))
		if !errors.Is(err, test.expectedError) {
			t.Errorf(
				"%d: expected decoding error %v "+
					"instead got %v", i, test.expectedError, err,
			)
			continue
		}
		if err != nil {
			// End test case here if a decoding error was expected.
			continue
		}
		// Check that it encodes to the same string
		encoded, err := Encode(hrp, decoded)
		if err != nil {
			t.Errorf("encoding failed: %v", err)
		}
		if !bytes.Equal(encoded, bytes.ToLower([]byte(str))) {
			t.Errorf(
				"expected data to encode to %v, but got %v",
				str, encoded,
			)
		}
		// Flip a bit in the string an make sure it is caught.
		pos := bytes.LastIndexAny(str, "1")
		flipped := []byte(string(str[:pos+1]) + string(str[pos+1]^1) + string(str[pos+2:]))
		_, _, err = Decode(flipped)
		if err == nil {
			t.Error("expected decoding to fail")
		}
	}
}

// TestBech32M tests that the following set of strings, based on the test
// vectors in BIP-350 are either valid or invalid using the new bech32m
// checksum algo. Some of these strings are similar to the set of above test
// vectors, but end up with different checksums.
func TestBech32M(t *testing.T) {
	tests := []struct {
		str           string
		expectedError error
	}{
		{"A1LQFN3A", nil},
		{"a1lqfn3a", nil},
		{
			"an83characterlonghumanreadablepartthatcontainsthetheexcludedcharactersbioandnumber11sg7hg6",
			nil,
		},
		{"abcdef1l7aum6echk45nj3s0wdvt2fg8x9yrzpqzd3ryx", nil},
		{
			"11llllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllludsr8",
			nil,
		},
		{"split1checkupstagehandshakeupstreamerranterredcaperredlc445v", nil},
		{"?1v759aa", nil},
		// Additional test vectors used in bitcoin core
		{"\x201xj0phk", ErrInvalidCharacter('\x20')},
		{"\x7f1g6xzxy", ErrInvalidCharacter('\x7f')},
		{"\x801vctc34", ErrInvalidCharacter('\x80')},
		{
			"an84characterslonghumanreadablepartthatcontainsthetheexcludedcharactersbioandnumber11d6pts4",
			ErrInvalidLength(91),
		},
		{"qyrz8wqd2c9m", ErrInvalidSeparatorIndex(-1)},
		{"1qyrz8wqd2c9m", ErrInvalidSeparatorIndex(0)},
		{"y1b0jsk6g", ErrNonCharsetChar(98)},
		{"lt1igcx5c0", ErrNonCharsetChar(105)},
		{"in1muywd", ErrInvalidSeparatorIndex(2)},
		{"mm1crxm3i", ErrNonCharsetChar(105)},
		{"au1s5cgom", ErrNonCharsetChar(111)},
		{"M1VUXWEZ", ErrInvalidChecksum{"mzl49c", "mzl49cw70eq6", "vuxwez"}},
		{"16plkw9", ErrInvalidLength(7)},
		{"1p2gdwpf", ErrInvalidSeparatorIndex(0)},

		{" 1nwldj5", ErrInvalidCharacter(' ')},
		{"\x7f" + "1axkwrx", ErrInvalidCharacter(0x7f)},
		{"\x801eym55h", ErrInvalidCharacter(0x80)},
	}
	for i, test := range tests {
		str := []byte(test.str)
		hrp, decoded, err := Decode(str)
		if test.expectedError != err {
			t.Errorf(
				"%d: (%v) expected decoding error %v "+
					"instead got %v", i, str, test.expectedError,
				err,
			)
			continue
		}
		if err != nil {
			// End test case here if a decoding error was expected.
			continue
		}
		// Check that it encodes to the same string, using bech32 m.
		encoded, err := EncodeM(hrp, decoded)
		if err != nil {
			t.Errorf("encoding failed: %v", err)
		}

		if !bytes.Equal(encoded, bytes.ToLower(str)) {
			t.Errorf(
				"expected data to encode to %v, but got %v",
				str, encoded,
			)
		}
		// Flip a bit in the string an make sure it is caught.
		pos := bytes.LastIndexAny(str, "1")
		flipped := []byte(string(str[:pos+1]) + string(str[pos+1]^1) + string(str[pos+2:]))
		_, _, err = Decode(flipped)
		if err == nil {
			t.Error("expected decoding to fail")
		}
	}
}

// TestBech32DecodeGeneric tests that given a bech32 string, or a bech32m
// string, the proper checksum version is returned so that callers can perform
// segwit addr validation.
func TestBech32DecodeGeneric(t *testing.T) {
	tests := []struct {
		str     string
		version Version
	}{
		{"A1LQFN3A", VersionM},
		{"a1lqfn3a", VersionM},
		{
			"an83characterlonghumanreadablepartthatcontainsthetheexcludedcharactersbioandnumber11sg7hg6",
			VersionM,
		},
		{"abcdef1l7aum6echk45nj3s0wdvt2fg8x9yrzpqzd3ryx", VersionM},
		{
			"11llllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllludsr8",
			VersionM,
		},
		{
			"split1checkupstagehandshakeupstreamerranterredcaperredlc445v",
			VersionM,
		},
		{"?1v759aa", VersionM},
		{"A12UEL5L", Version0},
		{"a12uel5l", Version0},
		{
			"an83characterlonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio1tt5tgs",
			Version0,
		},
		{"abcdef1qpzry9x8gf2tvdw0s3jn54khce6mua7lmqqqxw", Version0},
		{
			"11qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqc8247j",
			Version0,
		},
		{
			"split1checkupstagehandshakeupstreamerranterredcaperred2y9e3w",
			Version0,
		},
		{"BC1QW508D6QEJXTDG4Y5R3ZARVARY0C5XW7KV8F3T4", Version0},
		{
			"tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3q0sl5k7",
			Version0,
		},
		{
			"bc1pw508d6qejxtdg4y5r3zarvary0c5xw7kw508d6qejxtdg4y5r3zarvary0c5xw7kt5nd6y",
			VersionM,
		},
		{"BC1SW50QGDZ25J", VersionM},
		{"bc1zw508d6qejxtdg4y5r3zarvaryvaxxpcs", VersionM},
		{
			"tb1qqqqqp399et2xygdj5xreqhjjvcmzhxw4aywxecjdzew6hylgvsesrxh6hy",
			Version0,
		},
		{
			"tb1pqqqqp399et2xygdj5xreqhjjvcmzhxw4aywxecjdzew6hylgvsesf3hn0c",
			VersionM,
		},
		{
			"bc1p0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7vqzk5jj0",
			VersionM,
		},
	}
	for i, test := range tests {
		_, _, version, err := DecodeGeneric([]byte(test.str))
		if err != nil {
			t.Errorf(
				"%d: (%v) unexpected error during "+
					"decoding: %v", i, test.str, err,
			)
			continue
		}
		if version != test.version {
			t.Errorf(
				"(%v): invalid version: expected %v, got %v",
				test.str, test.version, version,
			)
		}
	}
}

// TestMixedCaseEncode ensures mixed case HRPs are converted to lowercase as
// expected when encoding and that decoding the produced encoding when converted
// to all uppercase produces the lowercase HRP and original data.
func TestMixedCaseEncode(t *testing.T) {
	tests := []struct {
		name    string
		hrp     string
		data    string
		encoded string
	}{
		{
			name:    "all uppercase HRP with no data",
			hrp:     "A",
			data:    "",
			encoded: "a12uel5l",
		}, {
			name:    "all uppercase HRP with data",
			hrp:     "UPPERCASE",
			data:    "787878",
			encoded: "uppercase10pu8sss7kmp",
		}, {
			name:    "mixed case HRP even offsets uppercase",
			hrp:     "AbCdEf",
			data:    "00443214c74254b635cf84653a56d7c675be77df",
			encoded: "abcdef1qpzry9x8gf2tvdw0s3jn54khce6mua7lmqqqxw",
		}, {
			name:    "mixed case HRP odd offsets uppercase ",
			hrp:     "aBcDeF",
			data:    "00443214c74254b635cf84653a56d7c675be77df",
			encoded: "abcdef1qpzry9x8gf2tvdw0s3jn54khce6mua7lmqqqxw",
		}, {
			name:    "all lowercase HRP",
			hrp:     "abcdef",
			data:    "00443214c74254b635cf84653a56d7c675be77df",
			encoded: "abcdef1qpzry9x8gf2tvdw0s3jn54khce6mua7lmqqqxw",
		},
	}
	for _, test := range tests {
		// Convert the text hex to bytes, convert those bytes from base256 to
		// base32, then ensure the encoded result with the HRP provided in the
		// test data is as expected.
		data, err := hex.DecodeString(test.data)
		if err != nil {
			t.Errorf("%q: invalid hex %q: %v", test.name, test.data, err)
			continue
		}
		convertedData, err := ConvertBits(data, 8, 5, true)
		if err != nil {
			t.Errorf(
				"%q: unexpected convert bits error: %v", test.name,
				err,
			)
			continue
		}
		gotEncoded, err := Encode([]byte(test.hrp), convertedData)
		if err != nil {
			t.Errorf("%q: unexpected encode error: %v", test.name, err)
			continue
		}
		if !bytes.Equal(gotEncoded, []byte(test.encoded)) {
			t.Errorf(
				"%q: mismatched encoding -- got %q, want %q", test.name,
				gotEncoded, test.encoded,
			)
			continue
		}
		// Ensure the decoding the expected lowercase encoding converted to all
		// uppercase produces the lowercase HRP and original data.
		gotHRP, gotData, err := Decode(bytes.ToUpper([]byte(test.encoded)))
		if err != nil {
			t.Errorf("%q: unexpected decode error: %v", test.name, err)
			continue
		}
		wantHRP := strings.ToLower(test.hrp)
		if !bytes.Equal(gotHRP, []byte(wantHRP)) {
			t.Errorf(
				"%q: mismatched decoded HRP -- got %q, want %q", test.name,
				gotHRP, wantHRP,
			)
			continue
		}
		convertedGotData, err := ConvertBits(gotData, 5, 8, false)
		if err != nil {
			t.Errorf(
				"%q: unexpected convert bits error: %v", test.name,
				err,
			)
			continue
		}
		if !bytes.Equal(convertedGotData, data) {
			t.Errorf(
				"%q: mismatched data -- got %x, want %x", test.name,
				convertedGotData, data,
			)
			continue
		}
	}
}

// TestCanDecodeUnlimtedBech32 tests whether decoding a large bech32 string works
// when using the DecodeNoLimit version
func TestCanDecodeUnlimtedBech32(t *testing.T) {
	input := "11qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq5kx0yd"
	// Sanity check that an input of this length errors on regular Decode()
	_, _, err := Decode([]byte(input))
	if err == nil {
		t.Fatalf("Test vector not appropriate")
	}
	// Try and decode it.
	hrp, data, err := DecodeNoLimit([]byte(input))
	if err != nil {
		t.Fatalf(
			"Expected decoding of large string to work. Got error: %v",
			err,
		)
	}
	// Verify data for correctness.
	if !bytes.Equal(hrp, []byte("1")) {
		t.Fatalf("Unexpected hrp: %v", hrp)
	}
	decodedHex := fmt.Sprintf("%x", data)
	expected := "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000"
	if decodedHex != expected {
		t.Fatalf("Unexpected decoded data: %s", decodedHex)
	}
}

// TestBech32Base256 ensures decoding and encoding various bech32, HRPs, and
// data produces the expected results when using EncodeFromBase256 and
// DecodeToBase256.  It includes tests for proper handling of case
// manipulations.
func TestBech32Base256(t *testing.T) {
	tests := []struct {
		name    string // test name
		encoded string // bech32 string to decode
		hrp     string // expected human-readable part
		data    string // expected hex-encoded data
		err     error  // expected error
	}{
		{
			name:    "all uppercase, no data",
			encoded: "A12UEL5L",
			hrp:     "a",
			data:    "",
		}, {
			name:    "long hrp with separator and excluded chars, no data",
			encoded: "an83characterlonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio1tt5tgs",
			hrp:     "an83characterlonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio",
			data:    "",
		}, {
			name:    "6 char hrp with data with leading zero",
			encoded: "abcdef1qpzry9x8gf2tvdw0s3jn54khce6mua7lmqqqxw",
			hrp:     "abcdef",
			data:    "00443214c74254b635cf84653a56d7c675be77df",
		}, {
			name:    "hrp same as separator and max length encoded string",
			encoded: "11qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqc8247j",
			hrp:     "1",
			data:    "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		}, {
			name:    "5 char hrp with data chosen to produce human-readable data part",
			encoded: "split1checkupstagehandshakeupstreamerranterredcaperred2y9e3w",
			hrp:     "split",
			data:    "c5f38b70305f519bf66d85fb6cf03058f3dde463ecd7918f2dc743918f2d",
		}, {
			name:    "same as previous but with checksum invalidated",
			encoded: "split1checkupstagehandshakeupstreamerranterredcaperred2y9e2w",
			err:     ErrInvalidChecksum{"2y9e3w", "2y9e3wlc445v", "2y9e2w"},
		}, {
			name:    "hrp with invalid character (space)",
			encoded: "s lit1checkupstagehandshakeupstreamerranterredcaperredp8hs2p",
			err:     ErrInvalidCharacter(' '),
		}, {
			name:    "hrp with invalid character (DEL)",
			encoded: "spl\x7ft1checkupstagehandshakeupstreamerranterredcaperred2y9e3w",
			err:     ErrInvalidCharacter(127),
		}, {
			name:    "data part with invalid character (o)",
			encoded: "split1cheo2y9e2w",
			err:     ErrNonCharsetChar('o'),
		}, {
			name:    "data part too short",
			encoded: "split1a2y9w",
			err:     ErrInvalidSeparatorIndex(5),
		}, {
			name:    "empty hrp",
			encoded: "1checkupstagehandshakeupstreamerranterredcaperred2y9e3w",
			err:     ErrInvalidSeparatorIndex(0),
		}, {
			name:    "no separator",
			encoded: "pzry9x0s0muk",
			err:     ErrInvalidSeparatorIndex(-1),
		}, {
			name:    "too long by one char",
			encoded: "11qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqc8247j",
			err:     ErrInvalidLength(91),
		}, {
			name:    "invalid due to mixed case in hrp",
			encoded: "aBcdef1qpzry9x8gf2tvdw0s3jn54khce6mua7lmqqqxw",
			err:     ErrMixedCase{},
		}, {
			name:    "invalid due to mixed case in data part",
			encoded: "abcdef1Qpzry9x8gf2tvdw0s3jn54khce6mua7lmqqqxw",
			err:     ErrMixedCase{},
		},
	}
	for _, test := range tests {
		// Ensure the decode either produces an error or not as expected.
		str := test.encoded
		gotHRP, gotData, err := DecodeToBase256([]byte(str))
		if test.err != err {
			t.Errorf(
				"%q: unexpected decode error -- got %v, want %v",
				test.name, err, test.err,
			)
			continue
		}
		if err != nil {
			// End test case here if a decoding error was expected.
			continue
		}
		// Ensure the expected HRP and original data are as expected.
		if !bytes.Equal(gotHRP, []byte(test.hrp)) {
			t.Errorf(
				"%q: mismatched decoded HRP -- got %q, want %q", test.name,
				gotHRP, test.hrp,
			)
			continue
		}
		data, err := hex.DecodeString(test.data)
		if err != nil {
			t.Errorf("%q: invalid hex %q: %v", test.name, test.data, err)
			continue
		}
		if !bytes.Equal(gotData, data) {
			t.Errorf(
				"%q: mismatched data -- got %x, want %x", test.name,
				gotData, data,
			)
			continue
		}
		// Encode the same data with the HRP converted to all uppercase and
		// ensure the result is the lowercase version of the original encoded
		// bech32 string.
		gotEncoded, err := EncodeFromBase256(
			bytes.ToUpper([]byte(test.hrp)), data,
		)
		if err != nil {
			t.Errorf(
				"%q: unexpected uppercase HRP encode error: %v", test.name,
				err,
			)
		}
		wantEncoded := bytes.ToLower([]byte(str))
		if !bytes.Equal(gotEncoded, wantEncoded) {
			t.Errorf(
				"%q: mismatched encoding -- got %q, want %q", test.name,
				gotEncoded, wantEncoded,
			)
		}
		// Encode the same data with the HRP converted to all lowercase and
		// ensure the result is the lowercase version of the original encoded
		// bech32 string.
		gotEncoded, err = EncodeFromBase256(
			bytes.ToLower([]byte(test.hrp)), data,
		)
		if err != nil {
			t.Errorf(
				"%q: unexpected lowercase HRP encode error: %v", test.name,
				err,
			)
		}
		if !bytes.Equal(gotEncoded, wantEncoded) {
			t.Errorf(
				"%q: mismatched encoding -- got %q, want %q", test.name,
				gotEncoded, wantEncoded,
			)
		}
		// Encode the same data with the HRP converted to mixed upper and
		// lowercase and ensure the result is the lowercase version of the
		// original encoded bech32 string.
		var mixedHRPBuilder bytes.Buffer
		for i, r := range test.hrp {
			if i%2 == 0 {
				mixedHRPBuilder.WriteString(strings.ToUpper(string(r)))
				continue
			}
			mixedHRPBuilder.WriteRune(r)
		}
		gotEncoded, err = EncodeFromBase256(mixedHRPBuilder.Bytes(), data)
		if err != nil {
			t.Errorf(
				"%q: unexpected lowercase HRP encode error: %v", test.name,
				err,
			)
		}
		if !bytes.Equal(gotEncoded, wantEncoded) {
			t.Errorf(
				"%q: mismatched encoding -- got %q, want %q", test.name,
				gotEncoded, wantEncoded,
			)
		}
		// Ensure a bit flip in the string is caught.
		pos := strings.LastIndexAny(test.encoded, "1")
		flipped := str[:pos+1] + string(str[pos+1]^1) + str[pos+2:]
		_, _, err = DecodeToBase256([]byte(flipped))
		if err == nil {
			t.Error("expected decoding to fail")
		}
	}
}

// BenchmarkEncodeDecodeCycle performs a benchmark for a full encode/decode
// cycle of a bech32 string. It also  reports the allocation count, which we
// expect to be 2 for a fully optimized cycle.
func BenchmarkEncodeDecodeCycle(b *testing.B) {
	// Use a fixed, 49-byte raw data for testing.
	inputData, err := hex.DecodeString("cbe6365ddbcda9a9915422c3f091c13f8c7b2f263b8d34067bd12c274408473fa764871c9dd51b1bb34873b3473b633ed1")
	if err != nil {
		b.Fatalf("failed to initialize input data: %v", err)
	}
	// Convert this into a 79-byte, base 32 byte slice.
	base32Input, err := ConvertBits(inputData, 8, 5, true)
	if err != nil {
		b.Fatalf("failed to convert input to 32 bits-per-element: %v", err)
	}
	// Use a fixed hrp for the tests. This should generate an encoded bech32
	// string of size 90 (the maximum allowed by BIP-173).
	hrp := "bc"
	// Begin the benchmark. Given that we test one roundtrip per iteration
	// (that is, one Encode() and one Decode() operation), we expect at most
	// 2 allocations per reported test op.
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		str, err := Encode([]byte(hrp), base32Input)
		if err != nil {
			b.Fatalf("failed to encode input: %v", err)
		}
		_, _, err = Decode(str)
		if err != nil {
			b.Fatalf("failed to decode string: %v", err)
		}
	}
}

// TestConvertBits tests whether base conversion works using TestConvertBits().
func TestConvertBits(t *testing.T) {
	tests := []struct {
		input    string
		output   string
		fromBits uint8
		toBits   uint8
		pad      bool
	}{
		// Trivial empty conversions.
		{"", "", 8, 5, false},
		{"", "", 8, 5, true},
		{"", "", 5, 8, false},
		{"", "", 5, 8, true},
		// Conversions of 0 value with/without padding.
		{"00", "00", 8, 5, false},
		{"00", "0000", 8, 5, true},
		{"0000", "00", 5, 8, false},
		{"0000", "0000", 5, 8, true},
		// Testing when conversion ends exactly at the byte edge. This makes
		// both padded and unpadded versions the same.
		{"0000000000", "0000000000000000", 8, 5, false},
		{"0000000000", "0000000000000000", 8, 5, true},
		{"0000000000000000", "0000000000", 5, 8, false},
		{"0000000000000000", "0000000000", 5, 8, true},
		// Conversions of full byte sequences.
		{"ffffff", "1f1f1f1f1e", 8, 5, true},
		{"1f1f1f1f1e", "ffffff", 5, 8, false},
		{"1f1f1f1f1e", "ffffff00", 5, 8, true},
		// Sample random conversions.
		{"c9ca", "190705", 8, 5, false},
		{"c9ca", "19070500", 8, 5, true},
		{"19070500", "c9ca", 5, 8, false},
		{"19070500", "c9ca00", 5, 8, true},
		// Test cases tested on TestConvertBitsFailures with their corresponding
		// fixes.
		{"ff", "1f1c", 8, 5, true},
		{"1f1c10", "ff20", 5, 8, true},
		// Large conversions.
		{
			"cbe6365ddbcda9a9915422c3f091c13f8c7b2f263b8d34067bd12c274408473fa764871c9dd51b1bb34873b3473b633ed1",
			"190f13030c170e1b1916141a13040a14040b011f01040e01071e0607160b1906070e06130801131b1a0416020e110008081c1f1a0e19040703120e1d0a06181b160d0407070c1a07070d11131d1408",
			8, 5, true,
		},
		{
			"190f13030c170e1b1916141a13040a14040b011f01040e01071e0607160b1906070e06130801131b1a0416020e110008081c1f1a0e19040703120e1d0a06181b160d0407070c1a07070d11131d1408",
			"cbe6365ddbcda9a9915422c3f091c13f8c7b2f263b8d34067bd12c274408473fa764871c9dd51b1bb34873b3473b633ed100",
			5, 8, true,
		},
	}
	for i, tc := range tests {
		input, err := hex.DecodeString(tc.input)
		if err != nil {
			t.Fatalf("invalid test input data: %v", err)
		}
		expected, err := hex.DecodeString(tc.output)
		if err != nil {
			t.Fatalf("invalid test output data: %v", err)
		}
		actual, err := ConvertBits(input, tc.fromBits, tc.toBits, tc.pad)
		if err != nil {
			t.Fatalf("test case %d failed: %v", i, err)
		}
		if !bytes.Equal(actual, expected) {
			t.Fatalf(
				"test case %d has wrong output; expected=%x actual=%x",
				i, expected, actual,
			)
		}
	}
}

// TestConvertBitsFailures tests for the expected conversion failures of
// ConvertBits().
func TestConvertBitsFailures(t *testing.T) {
	tests := []struct {
		input    string
		fromBits uint8
		toBits   uint8
		pad      bool
		err      error
	}{
		// Not enough output bytes when not using padding.
		{"ff", 8, 5, false, ErrInvalidIncompleteGroup{}},
		{"1f1c10", 5, 8, false, ErrInvalidIncompleteGroup{}},
		// Unsupported bit conversions.
		{"", 0, 5, false, ErrInvalidBitGroups{}},
		{"", 10, 5, false, ErrInvalidBitGroups{}},
		{"", 5, 0, false, ErrInvalidBitGroups{}},
		{"", 5, 10, false, ErrInvalidBitGroups{}},
	}
	for i, tc := range tests {
		input, err := hex.DecodeString(tc.input)
		if err != nil {
			t.Fatalf("invalid test input data: %v", err)
		}
		_, err = ConvertBits(input, tc.fromBits, tc.toBits, tc.pad)
		if err != tc.err {
			t.Fatalf(
				"test case %d failure: expected '%v' got '%v'", i,
				tc.err, err,
			)
		}
	}
}

// BenchmarkConvertBitsDown benchmarks the speed and memory allocation behavior
// of ConvertBits when converting from a higher base into a lower base (e.g. 8
// => 5).
//
// Only a single allocation is expected, which is used for the output array.
func BenchmarkConvertBitsDown(b *testing.B) {
	// Use a fixed, 49-byte raw data for testing.
	inputData, err := hex.DecodeString("cbe6365ddbcda9a9915422c3f091c13f8c7b2f263b8d34067bd12c274408473fa764871c9dd51b1bb34873b3473b633ed1")
	if err != nil {
		b.Fatalf("failed to initialize input data: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ConvertBits(inputData, 8, 5, true)
		if err != nil {
			b.Fatalf("error converting bits: %v", err)
		}
	}
}

// BenchmarkConvertBitsDown benchmarks the speed and memory allocation behavior
// of ConvertBits when converting from a lower base into a higher base (e.g. 5
// => 8).
//
// Only a single allocation is expected, which is used for the output array.
func BenchmarkConvertBitsUp(b *testing.B) {
	// Use a fixed, 79-byte raw data for testing.
	inputData, err := hex.DecodeString("190f13030c170e1b1916141a13040a14040b011f01040e01071e0607160b1906070e06130801131b1a0416020e110008081c1f1a0e19040703120e1d0a06181b160d0407070c1a07070d11131d1408")
	if err != nil {
		b.Fatalf("failed to initialize input data: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ConvertBits(inputData, 8, 5, true)
		if err != nil {
			b.Fatalf("error converting bits: %v", err)
		}
	}
}
