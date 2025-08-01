// Copyright (c) 2017 The btcsuite developers
// Copyright (c) 2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bech32

import (
	"bytes"
	"strings"
)

// Charset is the set of characters used in the data section of bech32 strings.
// Note that this is ordered, such that for a given charset[i], i is the binary
// value of the character.
//
// This wasn't exported in the original lol.
const Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

// gen encodes the generator polynomial for the bech32 BCH checksum.
var gen = []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}

// toBytes converts each character in the string 'chars' to the value of the
// index of the corresponding character in 'charset'.
func toBytes(chars []byte) ([]byte, error) {
	decoded := make([]byte, 0, len(chars))
	for i := 0; i < len(chars); i++ {
		index := strings.IndexByte(Charset, chars[i])
		if index < 0 {
			return nil, ErrNonCharsetChar(chars[i])
		}
		decoded = append(decoded, byte(index))
	}
	return decoded, nil
}

// bech32Polymod calculates the BCH checksum for a given hrp, values and
// checksum data. Checksum is optional, and if nil a 0 checksum is assumed.
//
// Values and checksum (if provided) MUST be encoded as 5 bits per element (base
// 32), otherwise the results are undefined.
//
// For more details on the polymod calculation, please refer to BIP 173.
func bech32Polymod(hrp []byte, values, checksum []byte) int {
	check := 1
	// Account for the high bits of the HRP in the checksum.
	for i := 0; i < len(hrp); i++ {
		b := check >> 25
		hiBits := int(hrp[i]) >> 5
		check = (check&0x1ffffff)<<5 ^ hiBits
		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				check ^= gen[i]
			}
		}
	}
	// Account for the separator (0) between high and low bits of the HRP.
	// x^0 == x, so we eliminate the redundant xor used in the other rounds.
	b := check >> 25
	check = (check & 0x1ffffff) << 5
	for i := 0; i < 5; i++ {
		if (b>>uint(i))&1 == 1 {
			check ^= gen[i]
		}
	}
	// Account for the low bits of the HRP.
	for i := 0; i < len(hrp); i++ {
		b := check >> 25
		loBits := int(hrp[i]) & 31
		check = (check&0x1ffffff)<<5 ^ loBits
		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				check ^= gen[i]
			}
		}
	}
	// Account for the values.
	for _, v := range values {
		b := check >> 25
		check = (check&0x1ffffff)<<5 ^ int(v)
		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				check ^= gen[i]
			}
		}
	}
	if checksum == nil {
		// A nil checksum is used during encoding, so assume all bytes are zero.
		// x^0 == x, so we eliminate the redundant xor used in the other rounds.
		for v := 0; v < 6; v++ {
			b := check >> 25
			check = (check & 0x1ffffff) << 5
			for i := 0; i < 5; i++ {
				if (b>>uint(i))&1 == 1 {
					check ^= gen[i]
				}
			}
		}
	} else {
		// Checksum is provided during decoding, so use it.
		for _, v := range checksum {
			b := check >> 25
			check = (check&0x1ffffff)<<5 ^ int(v)
			for i := 0; i < 5; i++ {
				if (b>>uint(i))&1 == 1 {
					check ^= gen[i]
				}
			}
		}
	}
	return check
}

// writeBech32Checksum calculates the checksum data expected for a string that
// will have the given hrp and payload data and writes it to the provided string
// builder.
//
// The payload data MUST be encoded as a base 32 (5 bits per element) byte slice
// and the hrp MUST only use the allowed character set (ascii chars between 33
// and 126), otherwise the results are undefined.
//
// For more details on the checksum calculation, please refer to BIP 173.
func writeBech32Checksum(
	hrp []byte, data []byte, bldr *bytes.Buffer,
	version Version,
) {

	bech32Const := int(VersionToConsts[version])
	polymod := bech32Polymod(hrp, data, nil) ^ bech32Const
	for i := 0; i < 6; i++ {
		b := byte((polymod >> uint(5*(5-i))) & 31)
		// This can't fail, given we explicitly cap the previous b byte by the
		// first 31 bits.
		c := Charset[b]
		bldr.WriteByte(c)
	}
}

// bech32VerifyChecksum verifies whether the bech32 string specified by the
// provided hrp and payload data (encoded as 5 bits per element byte slice) has
// the correct checksum suffix. The version of bech32 used (bech32 OG, or
// bech32m) is also returned to allow the caller to perform proper address
// validation (segwitv0 should use bech32, v1+ should use bech32m).
//
// Data MUST have more than 6 elements, otherwise this function panics.
//
// For more details on the checksum verification, please refer to BIP 173.
func bech32VerifyChecksum(hrp []byte, data []byte) (Version, bool) {
	checksum := data[len(data)-6:]
	values := data[:len(data)-6]
	polymod := bech32Polymod(hrp, values, checksum)
	// Before BIP-350, we'd always check this against a static constant of
	// 1 to know if the checksum was computed properly. As we want to
	// generically support decoding for bech32m as well as bech32, we'll
	// look up the returned value and compare it to the set of defined
	// constants.
	bech32Version, ok := ConstsToVersion[ChecksumConst(polymod)]
	if ok {
		return bech32Version, true
	}
	return VersionUnknown, false
}

// DecodeNoLimit is a bech32 checksum version aware arbitrary string length
// decoder. This function will return the version of the decoded checksum
// constant so higher level validation can be performed to ensure the correct
// version of bech32 was used when encoding.
func decodeNoLimit(bech []byte) ([]byte, []byte, Version, error) {
	// The minimum allowed size of a bech32 string is 8 characters, since it
	// needs a non-empty HRP, a separator, and a 6 character checksum.
	if len(bech) < 8 {
		return nil, nil, VersionUnknown, ErrInvalidLength(len(bech))
	}
	// Only	ASCII characters between 33 and 126 are allowed.
	var hasLower, hasUpper bool
	for i := 0; i < len(bech); i++ {
		if bech[i] < 33 || bech[i] > 126 {
			return nil, nil, VersionUnknown, ErrInvalidCharacter(bech[i])
		}
		// The characters must be either all lowercase or all uppercase. Testing
		// directly with ascii codes is safe here, given the previous test.
		hasLower = hasLower || (bech[i] >= 97 && bech[i] <= 122)
		hasUpper = hasUpper || (bech[i] >= 65 && bech[i] <= 90)
		if hasLower && hasUpper {
			return nil, nil, VersionUnknown, ErrMixedCase{}
		}
	}
	// Bech32 standard uses only the lowercase for of strings for checksum
	// calculation.
	if hasUpper {
		bech = bytes.ToLower(bech)
	}
	// The string is invalid if the last '1' is non-existent, it is the
	// first character of the string (no human-readable part) or one of the
	// last 6 characters of the string (since checksum cannot contain '1').
	one := bytes.LastIndexByte(bech, '1')
	if one < 1 || one+7 > len(bech) {
		return nil, nil, VersionUnknown, ErrInvalidSeparatorIndex(one)
	}
	// The human-readable part is everything before the last '1'.
	hrp := bech[:one]
	data := bech[one+1:]
	// Each character corresponds to the byte with value of the index in
	// 'charset'.
	decoded, err := toBytes(data)
	if err != nil {
		return nil, nil, VersionUnknown, err
	}
	// Verify if the checksum (stored inside decoded[:]) is valid, given the
	// previously decoded hrp.
	bech32Version, ok := bech32VerifyChecksum(hrp, decoded)
	if !ok {
		// Invalid checksum. Calculate what it should have been, so that the
		// error contains this information.
		//
		// Extract the payload bytes and actual checksum in the string.
		actual := bech[len(bech)-6:]
		payload := decoded[:len(decoded)-6]
		// Calculate the expected checksum, given the hrp and payload
		// data. We'll actually compute _both_ possibly valid checksum
		// to further aide in debugging.
		var expectedBldr bytes.Buffer
		expectedBldr.Grow(6)
		writeBech32Checksum(hrp, payload, &expectedBldr, Version0)
		expectedVersion0 := expectedBldr.String()
		var b strings.Builder
		b.Grow(6)
		writeBech32Checksum(hrp, payload, &expectedBldr, VersionM)
		expectedVersionM := expectedBldr.String()
		err = ErrInvalidChecksum{
			Expected:  expectedVersion0,
			ExpectedM: expectedVersionM,
			Actual:    string(actual),
		}
		return nil, nil, VersionUnknown, err
	}
	// We exclude the last 6 bytes, which is the checksum.
	return hrp, decoded[:len(decoded)-6], bech32Version, nil
}

// DecodeNoLimit decodes a bech32 encoded string, returning the human-readable
// part and the data part excluding the checksum.  This function does NOT
// validate against the BIP-173 maximum length allowed for bech32 strings and
// is meant for use in custom applications (such as lightning network payment
// requests), NOT on-chain addresses.
//
// Note that the returned data is 5-bit (base32) encoded and the human-readable
// part will be lowercase.
func DecodeNoLimit(bech []byte) ([]byte, []byte, error) {
	hrp, data, _, err := decodeNoLimit(bech)
	return hrp, data, err
}

// Decode decodes a bech32 encoded string, returning the human-readable part and
// the data part excluding the checksum.
//
// Note that the returned data is 5-bit (base32) encoded and the human-readable
// part will be lowercase.
func Decode(bech []byte) ([]byte, []byte, error) {
	// The maximum allowed length for a bech32 string is 90.
	if len(bech) > 90 {
		return nil, nil, ErrInvalidLength(len(bech))
	}
	hrp, data, _, err := decodeNoLimit(bech)
	return hrp, data, err
}

// DecodeGeneric is identical to the existing Decode method, but will also
// return bech32 version that matches the decoded checksum. This method should
// be used when decoding segwit addresses, as it enables additional
// verification to ensure the proper checksum is used.
func DecodeGeneric(bech []byte) ([]byte, []byte, Version, error) {
	// The maximum allowed length for a bech32 string is 90.
	if len(bech) > 90 {
		return nil, nil, VersionUnknown, ErrInvalidLength(len(bech))
	}
	return decodeNoLimit(bech)
}

// encodeGeneric is the base bech32 encoding function that is aware of the
// existence of the checksum versions. This method is private, as the Encode
// and EncodeM methods are intended to be used instead.
func encodeGeneric(hrp []byte, data []byte, version Version) ([]byte, error) {
	// The resulting bech32 string is the concatenation of the lowercase
	// hrp, the separator 1, data and the 6-byte checksum.
	hrp = bytes.ToLower(hrp)
	var bldr bytes.Buffer
	bldr.Grow(len(hrp) + 1 + len(data) + 6)
	bldr.Write(hrp)
	bldr.WriteString("1")
	// Write the data part, using the bech32 charset.
	for _, b := range data {
		if int(b) >= len(Charset) {
			return nil, ErrInvalidDataByte(b)
		}
		bldr.WriteByte(Charset[b])
	}
	// Calculate and write the checksum of the data.
	writeBech32Checksum(hrp, data, &bldr, version)
	return bldr.Bytes(), nil
}

// Encode encodes a byte slice into a bech32 string with the given
// human-readable part (HRP).  The HRP will be converted to lowercase if needed
// since mixed cased encodings are not permitted and lowercase is used for
// checksum purposes.  Note that the bytes must each encode 5 bits (base32).
func Encode(hrp, data []byte) ([]byte, error) {
	return encodeGeneric(hrp, data, Version0)
}

// EncodeM is the exactly same as the Encode method, but it uses the new
// bech32m constant instead of the original one. It should be used whenever one
// attempts to encode a segwit address of v1 and beyond.
func EncodeM(hrp, data []byte) ([]byte, error) {
	return encodeGeneric(hrp, data, VersionM)
}

// ConvertBits converts a byte slice where each byte is encoding fromBits bits,
// to a byte slice where each byte is encoding toBits bits.
func ConvertBits(data []byte, fromBits, toBits uint8, pad bool) (
	[]byte,
	error,
) {

	if fromBits < 1 || fromBits > 8 || toBits < 1 || toBits > 8 {
		return nil, ErrInvalidBitGroups{}
	}
	// Determine the maximum size the resulting array can have after base
	// conversion, so that we can size it a single time. This might be off
	// by a byte depending on whether padding is used or not and if the input
	// data is a multiple of both fromBits and toBits, but we ignore that and
	// just size it to the maximum possible.
	maxSize := len(data)*int(fromBits)/int(toBits) + 1
	// The final bytes, each byte encoding toBits bits.
	regrouped := make([]byte, 0, maxSize)
	// Keep track of the next byte we create and how many bits we have
	// added to it out of the toBits goal.
	nextByte := byte(0)
	filledBits := uint8(0)
	for _, b := range data {
		// Discard unused bits.
		b <<= 8 - fromBits
		// How many bits remaining to extract from the input data.
		remFromBits := fromBits
		for remFromBits > 0 {
			// How many bits remaining to be added to the next byte.
			remToBits := toBits - filledBits
			// The number of bytes to next extract is the minimum of
			// remFromBits and remToBits.
			toExtract := remFromBits
			if remToBits < toExtract {
				toExtract = remToBits
			}
			// Add the next bits to nextByte, shifting the already
			// added bits to the left.
			nextByte = (nextByte << toExtract) | (b >> (8 - toExtract))
			// Discard the bits we just extracted and get ready for
			// next iteration.
			b <<= toExtract
			remFromBits -= toExtract
			filledBits += toExtract
			// If the nextByte is completely filled, we add it to
			// our regrouped bytes and start on the next byte.
			if filledBits == toBits {
				regrouped = append(regrouped, nextByte)
				filledBits = 0
				nextByte = 0
			}
		}
	}
	// We pad any unfinished group if specified.
	if pad && filledBits > 0 {
		nextByte <<= toBits - filledBits
		regrouped = append(regrouped, nextByte)
		filledBits = 0
		nextByte = 0
	}
	// Any incomplete group must be <= 4 bits, and all zeroes.
	if filledBits > 0 && (filledBits > 4 || nextByte != 0) {
		return nil, ErrInvalidIncompleteGroup{}
	}
	return regrouped, nil
}

// EncodeFromBase256 converts a base256-encoded byte slice into a base32-encoded
// byte slice and then encodes it into a bech32 string with the given
// human-readable part (HRP).  The HRP will be converted to lowercase if needed
// since mixed cased encodings are not permitted and lowercase is used for
// checksum purposes.
func EncodeFromBase256(hrp, data []byte) ([]byte, error) {
	converted, err := ConvertBits(data, 8, 5, true)
	if err != nil {
		return nil, err
	}
	return Encode(hrp, converted)
}

// DecodeToBase256 decodes a bech32-encoded string into its associated
// human-readable part (HRP) and base32-encoded data, converts that data to a
// base256-encoded byte slice and returns it along with the lowercase HRP.
func DecodeToBase256(bech []byte) ([]byte, []byte, error) {
	hrp, data, err := Decode(bech)
	if err != nil {
		return nil, nil, err
	}
	converted, err := ConvertBits(data, 5, 8, false)
	if err != nil {
		return nil, nil, err
	}
	return hrp, converted, nil
}
