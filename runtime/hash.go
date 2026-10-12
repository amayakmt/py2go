package runtime

import (
	"crypto/rand"
	"encoding/binary"
	"math/bits"
	"os"
	"strconv"
)

// hashSeed holds the SipHash key CPython derives from PYTHONHASHSEED: random
// by default, all zero for "0", and a fixed sequence for any other number.
// Set iteration order depends on it, so programs that print sets of strings
// match CPython only when the seed is fixed.
var hashK0, hashK1 = hashKey(os.Getenv("PYTHONHASHSEED"))

// hashKey derives the SipHash key from a PYTHONHASHSEED value.
func hashKey(seed string) (k0, k1 uint64) {
	var secret [24]byte
	switch seed {
	case "", "random":
		rand.Read(secret[:])
	default:
		n, err := strconv.ParseUint(seed, 10, 32)
		if err != nil {
			rand.Read(secret[:])
			break
		}
		if n != 0 {
			// CPython's lcg_urandom.
			x := uint32(n)
			for i := range secret {
				x = x*214013 + 2531011
				secret[i] = byte(x >> 16)
			}
		}
	}
	return binary.LittleEndian.Uint64(secret[0:8]), binary.LittleEndian.Uint64(secret[8:16])
}

// siphash13 is SipHash-1-3, the hash CPython 3.13 uses for str and bytes.
func siphash13(k0, k1 uint64, data []byte) uint64 {
	v0 := k0 ^ 0x736f6d6570736575
	v1 := k1 ^ 0x646f72616e646f6d
	v2 := k0 ^ 0x6c7967656e657261
	v3 := k1 ^ 0x7465646279746573
	round := func() {
		v0 += v1
		v1 = bits.RotateLeft64(v1, 13)
		v1 ^= v0
		v0 = bits.RotateLeft64(v0, 32)
		v2 += v3
		v3 = bits.RotateLeft64(v3, 16)
		v3 ^= v2
		v0 += v3
		v3 = bits.RotateLeft64(v3, 21)
		v3 ^= v0
		v2 += v1
		v1 = bits.RotateLeft64(v1, 17)
		v1 ^= v2
		v2 = bits.RotateLeft64(v2, 32)
	}
	n := len(data)
	for ; len(data) >= 8; data = data[8:] {
		m := binary.LittleEndian.Uint64(data)
		v3 ^= m
		round()
		v0 ^= m
	}
	last := uint64(n) << 56
	for i, b := range data {
		last |= uint64(b) << (8 * uint(i))
	}
	v3 ^= last
	round()
	v0 ^= last
	v2 ^= 0xff
	round()
	round()
	round()
	return v0 ^ v1 ^ v2 ^ v3
}

// hashStr is hash(s). CPython hashes the string in its internal form: one
// byte per character when every character is below 256, otherwise two or
// four bytes in native (little-endian) order.
func hashStr(s string) int64 { return hashStrWith(hashK0, hashK1, s) }

func hashStrWith(k0, k1 uint64, s string) int64 {
	if s == "" {
		return 0
	}
	maxRune := rune(0)
	n := 0
	for _, r := range s {
		if r > maxRune {
			maxRune = r
		}
		n++
	}
	var buf []byte
	switch {
	case maxRune < 0x100:
		buf = make([]byte, 0, n)
		for _, r := range s {
			buf = append(buf, byte(r))
		}
	case maxRune < 0x10000:
		buf = make([]byte, 0, 2*n)
		for _, r := range s {
			buf = binary.LittleEndian.AppendUint16(buf, uint16(r))
		}
	default:
		buf = make([]byte, 0, 4*n)
		for _, r := range s {
			buf = binary.LittleEndian.AppendUint32(buf, uint32(r))
		}
	}
	return fixHash(int64(siphash13(k0, k1, buf)))
}

// fixHash maps -1 to -2: -1 is the error value of CPython's hash function.
func fixHash(h int64) int64 {
	if h == -1 {
		return -2
	}
	return h
}

// hashInt is hash(n): n modulo 2**61 - 1, keeping the sign.
func hashInt(n int64) int64 {
	const modulus = 1<<61 - 1
	u := uint64(n)
	if n < 0 {
		u = -u
	}
	h := int64(u % modulus)
	if n < 0 {
		h = -h
	}
	return fixHash(h)
}

// hashTuple is CPython's xxHash based tuple hash.
func hashTuple(s Seq) int64 {
	const (
		prime1 = 11400714785074694791
		prime2 = 14029467366897019727
		prime5 = 2870177450012600261
	)
	acc := uint64(prime5)
	n := s.Len()
	for i := int64(0); i < n; i++ {
		lane := uint64(Hash(s.At(i)))
		acc += lane * prime2
		acc = bits.RotateLeft64(acc, 31)
		acc *= prime1
	}
	acc += uint64(n) ^ (prime5 ^ 3527539)
	if acc == ^uint64(0) {
		return 1546275796
	}
	return int64(acc)
}

// Hash is hash(v) for the hashable types: int, bool, str and tuples of them.
func Hash(v any) int64 {
	switch v := v.(type) {
	case int64:
		return hashInt(v)
	case int:
		return hashInt(int64(v))
	case bool:
		return b2i(v)
	case string:
		return hashStr(v)
	case Seq:
		if v.Kind() == "tuple" {
			return hashTuple(v)
		}
	}
	panic(newError("TypeError", "unhashable type"))
}
