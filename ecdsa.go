// Based on segher's wii.git "ec.c"
// Copyright 2007,2008  Segher Boessenkool  <segher@kernel.crashing.org>

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/big"
	"unicode"
)

var (
	curveN      = new(big.Int).SetBytes([]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x13, 0xe9, 0x74, 0xe7, 0x2f, 0x8a, 0x69, 0x22, 0x03, 0x1d, 0x26, 0x03, 0xcf, 0xe0, 0xd7})
	curveGBytes = []byte{0x00, 0xfa, 0xc9, 0xdf, 0xcb, 0xac, 0x83, 0x13, 0xbb, 0x21, 0x39, 0xf1, 0xbb, 0x75, 0x5f, 0xef, 0x65, 0xbc, 0x39, 0x1f, 0x8b, 0x36, 0xf8, 0xf8, 0xeb, 0x73, 0x71, 0xfd, 0x55, 0x8b, 0x01, 0x00, 0x6a, 0x08, 0xa4, 0x19, 0x03, 0x35, 0x06, 0x78, 0xe5, 0x85, 0x28, 0xbe, 0xbf, 0x8a, 0x0b, 0xef, 0xf8, 0x67, 0xa7, 0xca, 0x36, 0x71, 0x6f, 0x7e, 0x01, 0xf8, 0x10, 0x52}
	square      = []byte{0x00, 0x01, 0x04, 0x05, 0x10, 0x11, 0x14, 0x15, 0x40, 0x41, 0x44, 0x45, 0x50, 0x51, 0x54, 0x55}
)

func eltIsZero(d []byte) bool {
	for i := 0; i < 30; i++ {
		if d[i] != 0 {
			return false
		}
	}
	return true
}

func eltAdd(a []byte, b []byte) []byte {
	d := make([]byte, 30)
	for i := 0; i < 30; i++ {
		d[i] = a[i] ^ b[i]
	}
	return d
}

func eltMulX(d []byte, a []byte) {
	carry := a[0] & 1

	x := byte(0)
	for i := 0; i < 29; i++ {
		y := a[i+1]
		d[i] = x ^ (y >> 7)
		x = y << 1
	}
	d[29] = x ^ carry
	d[20] ^= carry << 2
}

func eltMul(a []byte, b []byte) []byte {
	d := make([]byte, 30)

	i := 0
	mask := byte(1)
	for n := 0; n < 233; n++ {
		eltMulX(d, d)

		if (a[i] & mask) != 0 {
			d = eltAdd(d, b)
		}

		mask >>= 1
		if mask == 0 {
			mask = 0x80
			i++
		}
	}

	return d
}

func eltSquareToWide(a []byte) []byte {
	d := make([]byte, 60)
	for i := 0; i < 30; i++ {
		d[2*i] = square[a[i]>>4]
		d[2*i+1] = square[a[i]&15]
	}
	return d
}

func wideReduce(d []byte) {
	for i := 0; i < 30; i++ {
		x := d[i]

		d[i+19] ^= x >> 7
		d[i+20] ^= x << 1

		d[i+29] ^= x >> 1
		d[i+30] ^= x << 7
	}

	x := d[30] & 0xfe

	d[49] ^= x >> 7
	d[50] ^= x << 1

	d[59] ^= x >> 1

	d[30] &= 1
}

func eltSquare(a []byte) []byte {
	wide := eltSquareToWide(a)
	wideReduce(wide)
	return wide[30:]
}

func itohTsujii(a []byte, b []byte, j uint32) []byte {
	t := bytes.Clone(a)
	for ; j != 0; j-- {
		t = eltSquare(t)
	}

	return eltMul(t, b)
}

func eltInv(a []byte) []byte {

	t := itohTsujii(a, a, 1)
	s := itohTsujii(t, a, 1)
	t = itohTsujii(s, s, 3)
	s = itohTsujii(t, a, 1)
	t = itohTsujii(s, s, 7)
	s = itohTsujii(t, t, 14)
	t = itohTsujii(s, a, 1)
	s = itohTsujii(t, t, 29)
	t = itohTsujii(s, s, 58)
	s = itohTsujii(t, t, 116)
	return eltSquare(s)
}

func pointIsZero(p []byte) bool {
	for i := 0; i < 60; i++ {
		if p[i] != 0 {
			return false
		}
	}
	return true
}

func pointDouble(p []byte) []byte {
	px := p[:30]
	py := p[30:]

	if eltIsZero(px) {
		return make([]byte, 60)
	}

	t := eltInv(px)
	s := eltMul(py, t)
	s = eltAdd(s, px)

	t = eltSquare(px)

	rx := eltSquare(s)
	rx = eltAdd(rx, s)
	rx[29] ^= 1

	ry := eltMul(s, rx)
	ry = eltAdd(ry, rx)
	ry = eltAdd(ry, t)

	return append(rx, ry...)
}

func pointAdd(p []byte, q []byte) []byte {
	if pointIsZero(p) {
		return bytes.Clone(q)
	}

	if pointIsZero(q) {
		return bytes.Clone(p)
	}

	px := p[:30]
	py := p[30:]
	qx := q[:30]
	qy := q[30:]

	u := eltAdd(px, qx)

	if eltIsZero(u) {
		u = eltAdd(py, qy)
		if eltIsZero(u) {
			return pointDouble(p)
		}
		return make([]byte, 60)
	}

	t := eltInv(u)
	u = eltAdd(py, qy)
	s := eltMul(t, u)

	t = eltSquare(s)
	t = eltAdd(t, s)
	t = eltAdd(t, qx)
	t[29] ^= 1

	u = eltMul(s, t)
	s = eltAdd(u, py)
	rx := eltAdd(t, px)
	ry := eltAdd(s, rx)

	return append(rx, ry...)
}

func pointMul(a []byte, b []byte) []byte {
	d := make([]byte, 60)

	for i := 0; i < 30; i++ {
		for mask := byte(0x80); mask != 0; mask >>= 1 {
			d = pointDouble(d)
			if (a[i] & mask) != 0 {
				d = pointAdd(d, b)
			}
		}
	}
	return d
}

func bigIntToBytes(a *big.Int) []byte {
	r := a.Bytes()
	if len(r) < 30 {
		return append(make([]byte, 30-len(r)), r...)
	}
	return r
}

func verifyECDSA(publicKey []byte, signature []byte, hash []byte) bool {
	r := big.NewInt(0).SetBytes(signature[0x00:0x1E])
	s := big.NewInt(0).SetBytes(signature[0x1E:0x3C])

	inv := big.NewInt(0).ModInverse(s, curveN)
	e := big.NewInt(0).SetBytes(hash)

	// printHex(bigIntToBytes(inv))

	w1 := big.NewInt(0).Mul(e, inv)
	w1.Mod(w1, curveN)
	w2 := big.NewInt(0).Mul(r, inv)
	w2.Mod(w2, curveN)

	// printHex(bigIntToBytes(w1))
	// printHex(bigIntToBytes(w2))

	r1 := pointMul(bigIntToBytes(w1), curveGBytes)
	r2 := pointMul(bigIntToBytes(w2), publicKey)
	r3 := pointAdd(r1, r2)
	rx := big.NewInt(0).SetBytes(r3[:30])

	// printHex(r1)
	// printHex(r2)
	// printHex(bigIntToBytes(rx))

	if rx.Cmp(curveN) >= 0 {
		// TODO: This is correct right?
		rx.Sub(rx, curveN)
		rx.Mod(rx, curveN)
	}

	return rx.Cmp(r) == 0
}

var msPublicKey = []byte{
	0x00, 0xFD, 0x56, 0x04, 0x18, 0x2C, 0xF1, 0x75, 0x09, 0x21, 0x00, 0xC3, 0x08, 0xAE, 0x48, 0x39,
	0x91, 0x1B, 0x6F, 0x9F, 0xA1, 0xD5, 0x3A, 0x95, 0xAF, 0x08, 0x33, 0x49, 0x47, 0x2B, 0x00, 0x01,
	0x71, 0x31, 0x69, 0xB5, 0x91, 0xFF, 0xD3, 0x0C, 0xBF, 0x73, 0xDA, 0x76, 0x64, 0xBA, 0x8D, 0x0D,
	0xF9, 0x5B, 0x4D, 0x11, 0x04, 0x44, 0x64, 0x35, 0xC0, 0xED, 0xA4, 0x2F,
}

func SanitizeBase64(encoded string) string {
	// Remove all whitespace
	var buf bytes.Buffer
	for _, r := range encoded {
		if !unicode.IsSpace(r) {
			buf.WriteRune(r)
		}
	}
	clean := buf.String()
	return clean
}

func verifySignature(authToken string, signature string) (uint32, error) {
	// Clean the string first
	// signature = SanitizeBase64(signature)
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return 0, err
	}

	if len(sigBytes) != 328 {
		return 0, fmt.Errorf("invalid size")
	}

	ngId := sigBytes[0x000:0x004]

	// TODO: Block Dolphin
	ngTimestamp := sigBytes[0x004:0x008]
	caId := sigBytes[0x008:0x00C]
	msId := sigBytes[0x00C:0x010]
	apId := sigBytes[0x010:0x018]
	msSignature := sigBytes[0x018:0x054]
	ngPublicKey := sigBytes[0x054:0x090]
	ngSignature := sigBytes[0x090:0x0CC]
	apPublicKey := sigBytes[0x0CC:0x108]
	apSignature := sigBytes[0x108:0x144]

	ngIssuer := fmt.Sprintf("Root-CA%02x%02x%02x%02x-MS%02x%02x%02x%02x", caId[0], caId[1], caId[2], caId[3], msId[0], msId[1], msId[2], msId[3])
	ngName := fmt.Sprintf("NG%02x%02x%02x%02x", ngId[0], ngId[1], ngId[2], ngId[3])

	ngCertBlob := []byte(ngIssuer)
	ngCertBlob = append(ngCertBlob, make([]byte, 0x40-len(ngIssuer))...)
	ngCertBlob = append(ngCertBlob, 0x00, 0x00, 0x00, 0x02)
	ngCertBlob = append(ngCertBlob, []byte(ngName)...)
	ngCertBlob = append(ngCertBlob, make([]byte, 0x40-len(ngName))...)
	ngCertBlob = append(ngCertBlob, ngTimestamp...)
	ngCertBlob = append(ngCertBlob, ngPublicKey...)
	ngCertBlob = append(ngCertBlob, make([]byte, 0x3C)...)
	ngCertBlobHash := sha1.Sum(ngCertBlob)

	if !verifyECDSA(msPublicKey, msSignature, ngCertBlobHash[:]) {
		return 0, fmt.Errorf("NG cert verify failed")
	}

	apIssuer := ngIssuer + "-" + ngName
	apName := fmt.Sprintf("AP%02x%02x%02x%02x%02x%02x%02x%02x", apId[0], apId[1], apId[2], apId[3], apId[4], apId[5], apId[6], apId[7])

	apCertBlob := []byte(apIssuer)
	apCertBlob = append(apCertBlob, make([]byte, 0x40-len(apIssuer))...)
	apCertBlob = append(apCertBlob, 0x00, 0x00, 0x00, 0x02)
	apCertBlob = append(apCertBlob, []byte(apName)...)
	apCertBlob = append(apCertBlob, make([]byte, 0x40-len(apName))...)
	apCertBlob = append(apCertBlob, 0x00, 0x00, 0x00, 0x00)
	apCertBlob = append(apCertBlob, apPublicKey...)
	apCertBlob = append(apCertBlob, make([]byte, 0x3C)...)
	apCertBlobHash := sha1.Sum(apCertBlob)

	if !verifyECDSA(ngPublicKey, ngSignature, apCertBlobHash[:]) {
		return 0, fmt.Errorf("AP cert verify failed")
	}

	authTokenHash := sha1.Sum([]byte(authToken))
	if !verifyECDSA(apPublicKey, apSignature, authTokenHash[:]) {
		return 0, fmt.Errorf("auth token signature failed")
	}

	return binary.BigEndian.Uint32(ngId), nil
}
