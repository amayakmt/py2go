package runtime

import (
	"crypto/pbkdf2"
	"crypto/sha3"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"math/bits"
	"strings"
)

var blake2Sigma = [10][16]byte{
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	{14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3},
	{11, 8, 12, 0, 5, 2, 15, 13, 10, 14, 3, 6, 7, 1, 9, 4},
	{7, 9, 3, 1, 13, 12, 11, 14, 2, 6, 5, 10, 4, 0, 15, 8},
	{9, 0, 5, 7, 2, 4, 10, 15, 14, 1, 11, 12, 6, 8, 3, 13},
	{2, 12, 6, 10, 0, 11, 8, 3, 4, 13, 7, 5, 15, 14, 1, 9},
	{12, 5, 1, 15, 14, 13, 4, 10, 0, 7, 6, 3, 9, 2, 8, 11},
	{13, 11, 7, 14, 12, 1, 3, 9, 5, 0, 15, 4, 8, 6, 2, 10},
	{6, 15, 14, 9, 11, 3, 0, 8, 12, 2, 13, 7, 1, 4, 10, 5},
	{10, 2, 8, 4, 7, 6, 1, 5, 15, 11, 9, 14, 3, 12, 13, 0},
}

var blake2bIV = [8]uint64{0x6a09e667f3bcc908, 0xbb67ae8584caa73b, 0x3c6ef372fe94f82b, 0xa54ff53a5f1d36f1, 0x510e527fade682d1, 0x9b05688c2b3e6c1f, 0x1f83d9abfb41bd6b, 0x5be0cd19137e2179}
var blake2sIV = [8]uint32{0x6A09E667, 0xBB67AE85, 0x3C6EF372, 0xA54FF53A, 0x510E527F, 0x9B05688C, 0x1F83D9AB, 0x5BE0CD19}

// blake2Params are the keyword arguments of blake2b and blake2s.
type blake2Params struct {
	size              int
	key, salt, person []byte
	fanout, depth     int
	leaf              uint32
	nodeOffset        uint64
	nodeDepth, inner  int
	last              bool
}

// blake2 is BLAKE2b (wide) or BLAKE2s (not wide).
type blake2 struct {
	wide  bool
	p     blake2Params
	h64   [8]uint64
	h32   [8]uint32
	t     uint64
	buf   []byte
	block int
}

func newBlake2(wide bool, p blake2Params) *blake2 {
	b := &blake2{wide: wide, p: p, block: 64}
	if wide {
		b.block = 128
	}
	b.reset()
	return b
}

func (b *blake2) reset() {
	p := b.p
	b.t, b.buf = 0, nil
	if b.wide {
		var w [8]uint64
		w[0] = uint64(p.size) | uint64(len(p.key))<<8 | uint64(p.fanout)<<16 | uint64(p.depth)<<24 | uint64(p.leaf)<<32
		w[1] = p.nodeOffset
		w[2] = uint64(p.nodeDepth) | uint64(p.inner)<<8
		var salt, per [16]byte
		copy(salt[:], p.salt)
		copy(per[:], p.person)
		w[4], w[5] = binary.LittleEndian.Uint64(salt[:8]), binary.LittleEndian.Uint64(salt[8:])
		w[6], w[7] = binary.LittleEndian.Uint64(per[:8]), binary.LittleEndian.Uint64(per[8:])
		for i := range b.h64 {
			b.h64[i] = blake2bIV[i] ^ w[i]
		}
	} else {
		var w [8]uint32
		w[0] = uint32(p.size) | uint32(len(p.key))<<8 | uint32(p.fanout)<<16 | uint32(p.depth)<<24
		w[1] = p.leaf
		w[2] = uint32(p.nodeOffset)
		w[3] = uint32(p.nodeOffset>>32)&0xffff | uint32(p.nodeDepth)<<16 | uint32(p.inner)<<24
		var salt, per [8]byte
		copy(salt[:], p.salt)
		copy(per[:], p.person)
		w[4], w[5] = binary.LittleEndian.Uint32(salt[:4]), binary.LittleEndian.Uint32(salt[4:])
		w[6], w[7] = binary.LittleEndian.Uint32(per[:4]), binary.LittleEndian.Uint32(per[4:])
		for i := range b.h32 {
			b.h32[i] = blake2sIV[i] ^ w[i]
		}
	}
	if len(p.key) > 0 {
		blk := make([]byte, b.block)
		copy(blk, p.key)
		b.buf = blk
	}
}

func (b *blake2) compress(block []byte, last bool, lastNode bool) {
	if b.wide {
		var m [16]uint64
		for i := range m {
			m[i] = binary.LittleEndian.Uint64(block[8*i:])
		}
		var v [16]uint64
		copy(v[:8], b.h64[:])
		copy(v[8:], blake2bIV[:])
		v[12] ^= b.t
		if last {
			v[14] = ^v[14]
			if lastNode {
				v[15] = ^v[15]
			}
		}
		g := func(a, b2, c, d int, x, y uint64) {
			v[a] = v[a] + v[b2] + x
			v[d] = bits.RotateLeft64(v[d]^v[a], -32)
			v[c] += v[d]
			v[b2] = bits.RotateLeft64(v[b2]^v[c], -24)
			v[a] = v[a] + v[b2] + y
			v[d] = bits.RotateLeft64(v[d]^v[a], -16)
			v[c] += v[d]
			v[b2] = bits.RotateLeft64(v[b2]^v[c], -63)
		}
		for r := 0; r < 12; r++ {
			s := &blake2Sigma[r%10]
			g(0, 4, 8, 12, m[s[0]], m[s[1]])
			g(1, 5, 9, 13, m[s[2]], m[s[3]])
			g(2, 6, 10, 14, m[s[4]], m[s[5]])
			g(3, 7, 11, 15, m[s[6]], m[s[7]])
			g(0, 5, 10, 15, m[s[8]], m[s[9]])
			g(1, 6, 11, 12, m[s[10]], m[s[11]])
			g(2, 7, 8, 13, m[s[12]], m[s[13]])
			g(3, 4, 9, 14, m[s[14]], m[s[15]])
		}
		for i := 0; i < 8; i++ {
			b.h64[i] ^= v[i] ^ v[i+8]
		}
		return
	}
	var m [16]uint32
	for i := range m {
		m[i] = binary.LittleEndian.Uint32(block[4*i:])
	}
	var v [16]uint32
	copy(v[:8], b.h32[:])
	copy(v[8:], blake2sIV[:])
	v[12] ^= uint32(b.t)
	v[13] ^= uint32(b.t >> 32)
	if last {
		v[14] = ^v[14]
		if lastNode {
			v[15] = ^v[15]
		}
	}
	g := func(a, b2, c, d int, x, y uint32) {
		v[a] = v[a] + v[b2] + x
		v[d] = bits.RotateLeft32(v[d]^v[a], -16)
		v[c] += v[d]
		v[b2] = bits.RotateLeft32(v[b2]^v[c], -12)
		v[a] = v[a] + v[b2] + y
		v[d] = bits.RotateLeft32(v[d]^v[a], -8)
		v[c] += v[d]
		v[b2] = bits.RotateLeft32(v[b2]^v[c], -7)
	}
	for r := 0; r < 10; r++ {
		s := &blake2Sigma[r]
		g(0, 4, 8, 12, m[s[0]], m[s[1]])
		g(1, 5, 9, 13, m[s[2]], m[s[3]])
		g(2, 6, 10, 14, m[s[4]], m[s[5]])
		g(3, 7, 11, 15, m[s[6]], m[s[7]])
		g(0, 5, 10, 15, m[s[8]], m[s[9]])
		g(1, 6, 11, 12, m[s[10]], m[s[11]])
		g(2, 7, 8, 13, m[s[12]], m[s[13]])
		g(3, 4, 9, 14, m[s[14]], m[s[15]])
	}
	for i := 0; i < 8; i++ {
		b.h32[i] ^= v[i] ^ v[i+8]
	}
}

func (b *blake2) Write(data []byte) (int, error) {
	n := len(data)
	for len(data) > 0 {
		if len(b.buf) == b.block {
			b.t += uint64(b.block)
			b.compress(b.buf, false, false)
			b.buf = b.buf[:0]
		}
		take := b.block - len(b.buf)
		if take > len(data) {
			take = len(data)
		}
		b.buf = append(b.buf, data[:take]...)
		data = data[take:]
	}
	return n, nil
}

func (b *blake2) Sum(in []byte) []byte {
	c := *b
	c.buf = append([]byte(nil), b.buf...)
	blk := make([]byte, c.block)
	copy(blk, c.buf)
	c.t += uint64(len(c.buf))
	c.compress(blk, true, c.p.last)
	var out []byte
	if c.wide {
		out = make([]byte, 64)
		for i, h := range c.h64 {
			binary.LittleEndian.PutUint64(out[8*i:], h)
		}
	} else {
		out = make([]byte, 32)
		for i, h := range c.h32 {
			binary.LittleEndian.PutUint32(out[4*i:], h)
		}
	}
	return append(in, out[:c.p.size]...)
}

func (b *blake2) Reset()         { b.reset() }
func (b *blake2) Size() int      { return b.p.size }
func (b *blake2) BlockSize() int { return b.block }
func (b *blake2) Clone() hash.Hash {
	c := *b
	c.buf = append([]byte(nil), b.buf...)
	return &c
}

// shake is hashlib's shake_128 and shake_256.
type shakeObj struct {
	name  string
	h     *sha3.SHAKE
	mk    func() *sha3.SHAKE
	data  []byte
	block int
}

func (s *shakeObj) PyClassName() string { return "HASHXOF" }
func (s *shakeObj) PyRepr() string {
	return "<" + s.name + " _hashlib.HASHXOF object @ " + addr(s) + ">"
}

var shakeObjNames = names("update digest hexdigest copy")

func (s *shakeObj) PyDynTable() map[string]bool { return shakeObjNames }

func (s *shakeObj) PyGetAttr(name string) (any, bool) {
	switch name {
	case "name":
		return s.name, true
	case "digest_size":
		return int64(0), true
	case "block_size":
		return int64(s.block), true
	}
	if shakeObjNames[name] {
		return func(args ...any) any {
			r, _ := s.DynMethod(name, args, nil)
			return r
		}, true
	}
	return nil, false
}

func (s *shakeObj) sum(n int) []byte {
	h := s.mk()
	h.Write(s.data)
	out := make([]byte, n)
	h.Read(out)
	return out
}

func (s *shakeObj) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "update":
		data, ok := bytesOf(argAt(args, 0))
		if !ok {
			if _, isStr := Unbox(argAt(args, 0)).(string); isStr {
				panic(newError("TypeError", "Strings must be encoded before hashing"))
			}
			panic(newError("TypeError", "object supporting the buffer API required"))
		}
		s.data = append(s.data, data...)
		return nil, true
	case "digest", "hexdigest":
		var length any = argAt(args, 0)
		for _, k := range kw {
			if k.Name == "length" {
				length = k.Val
			}
		}
		if length == nil {
			panic(newError("TypeError", name+"() missing required argument 'length' (pos 1)"))
		}
		n := DynInt(length)
		if n < 0 {
			panic(valueError("negative digest length"))
		}
		out := s.sum(int(n))
		if name == "digest" {
			return Bytes(out), true
		}
		return hex.EncodeToString(out), true
	case "copy":
		return &shakeObj{name: s.name, mk: s.mk, block: s.block, data: append([]byte(nil), s.data...)}, true
	}
	return nil, false
}

func blake2Bytes(v any, what string) []byte {
	b, ok := bytesOf(v)
	if !ok {
		panic(newError("TypeError", "a bytes-like object is required, not '"+TypeName(Unbox(v))+"'"))
	}
	return []byte(b)
}

func init() {
	RegisterNative("hashlib.blake2", func(args []any, kw []KW) any {
		// (wide, data, digest_size, key, salt, person, fanout, depth, leaf_size, node_offset, node_depth, inner_size, last_node)
		wide := DynBool(args[0])
		name := "blake2s"
		maxSize, maxKey, maxSalt := 32, 32, 8
		if wide {
			name, maxSize, maxKey, maxSalt = "blake2b", 64, 64, 16
		}
		size := int(DynInt(args[2]))
		if size < 1 || size > maxSize {
			panic(valueError("digest_size must be between 1 and " + itoa(maxSize) + " bytes"))
		}
		p := blake2Params{size: size, key: blake2Bytes(args[3], "key"), salt: blake2Bytes(args[4], "salt"), person: blake2Bytes(args[5], "person"),
			fanout: int(DynInt(args[6])), depth: int(DynInt(args[7])), leaf: uint32(DynInt(args[8])), nodeOffset: uint64(DynInt(args[9])), nodeDepth: int(DynInt(args[10])), inner: int(DynInt(args[11])), last: DynBool(args[12])}
		if len(p.key) > maxKey {
			panic(valueError("maximum key length is " + itoa(maxKey) + " bytes"))
		}
		if len(p.salt) > maxSalt {
			panic(valueError("maximum salt length is " + itoa(maxSalt) + " bytes"))
		}
		if len(p.person) > maxSalt {
			panic(valueError("maximum person length is " + itoa(maxSalt) + " bytes"))
		}
		if p.fanout < 0 || p.fanout > 255 {
			panic(valueError("fanout must be between 0 and 255"))
		}
		if p.depth <= 0 || p.depth > 255 {
			panic(valueError("depth must be between 1 and 255"))
		}
		h := &Hash2{name: name, h: newBlake2(wide, p), block: 64}
		if wide {
			h.block = 128
		}
		h.mk = func() hash.Hash { return newBlake2(wide, p) }
		if !IsNone(args[1]) {
			h.DynMethod("update", []any{args[1]}, nil)
		}
		return h
	})
	RegisterNative("hashlib.shake", func(args []any, kw []KW) any {
		bits := int(DynInt(args[0]))
		s := &shakeObj{name: "shake_" + itoa(bits), block: 168}
		s.mk = func() *sha3.SHAKE { return sha3.NewSHAKE128() }
		if bits == 256 {
			s.block = 136
			s.mk = func() *sha3.SHAKE { return sha3.NewSHAKE256() }
		}
		if len(args) > 1 && !IsNone(args[1]) {
			s.DynMethod("update", []any{args[1]}, nil)
		}
		return s
	})
	RegisterNative("hashlib.pbkdf2_hmac", func(args []any, kw []KW) any {
		nativeArgs("pbkdf2_hmac", args, 4, 5)
		name := strings.ReplaceAll(strings.ToLower(Str(args[0])), "-", "_")
		mk, ok := hashMakers[name]
		if !ok {
			panic(valueError("unsupported hash type " + Str(args[0])))
		}
		pw, ok1 := bytesOf(args[1])
		salt, ok2 := bytesOf(args[2])
		if !ok1 || !ok2 {
			panic(newError("TypeError", "a bytes-like object is required"))
		}
		iter := int(DynInt(args[3]))
		if iter < 1 {
			panic(valueError("iteration value must be greater than 0."))
		}
		dklen := mk.mk().Size()
		if len(args) > 4 && !IsNone(args[4]) {
			dklen = int(DynInt(args[4]))
			if dklen < 1 {
				panic(valueError("key length must be greater than 0."))
			}
		}
		key, err := pbkdf2.Key(mk.mk, string(pw), []byte(salt), iter, dklen)
		if err != nil {
			panic(valueError(err.Error()))
		}
		return Bytes(key)
	})
}
