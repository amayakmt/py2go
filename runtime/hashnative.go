package runtime

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"strings"
)

// Hash is a hashlib hash object.
type Hash2 struct {
	name  string
	h     hash.Hash
	mk    func() hash.Hash
	block int
}

func (h *Hash2) PyClassName() string { return "HASH" }

var hashMakers = map[string]struct {
	mk    func() hash.Hash
	block int
}{
	"md5":        {func() hash.Hash { return md5.New() }, 64},
	"sha1":       {func() hash.Hash { return sha1.New() }, 64},
	"sha224":     {func() hash.Hash { return sha256.New224() }, 64},
	"sha256":     {func() hash.Hash { return sha256.New() }, 64},
	"sha384":     {func() hash.Hash { return sha512.New384() }, 128},
	"sha512":     {func() hash.Hash { return sha512.New() }, 128},
	"sha512_224": {func() hash.Hash { return sha512.New512_224() }, 128},
	"sha512_256": {func() hash.Hash { return sha512.New512_256() }, 128},
	"sha3_224":   {func() hash.Hash { return sha3.New224() }, 144},
	"sha3_256":   {func() hash.Hash { return sha3.New256() }, 136},
	"sha3_384":   {func() hash.Hash { return sha3.New384() }, 104},
	"sha3_512":   {func() hash.Hash { return sha3.New512() }, 72},
}

var hashObjNames = names("update digest hexdigest copy")

func (h *Hash2) PyDynTable() map[string]bool { return hashObjNames }

func (h *Hash2) PyRepr() string {
	return "<" + h.name + " _hashlib.HASH object @ " + addr(h) + ">"
}

func (h *Hash2) PyGetAttr(name string) (any, bool) {
	switch name {
	case "name":
		return h.name, true
	case "digest_size":
		return int64(h.h.Size()), true
	case "block_size":
		return int64(h.block), true
	}
	if hashObjNames[name] {
		return func(args ...any) any {
			r, _ := h.DynMethod(name, args, nil)
			return r
		}, true
	}
	return nil, false
}

func (h *Hash2) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "update":
		data, ok := bytesOf(argAt(args, 0))
		if !ok {
			if _, isStr := Unbox(argAt(args, 0)).(string); isStr {
				panic(newError("TypeError", "Strings must be encoded before hashing"))
			}
			panic(newError("TypeError", "object supporting the buffer API required"))
		}
		h.h.Write([]byte(data))
		return nil, true
	case "digest":
		return Bytes(h.h.Sum(nil)), true
	case "hexdigest":
		return hex.EncodeToString(h.h.Sum(nil)), true
	case "copy":
		n := &Hash2{name: h.name, mk: h.mk, block: h.block}
		// Replay is not possible, so marshal the state.
		if m, ok := h.h.(interface{ MarshalBinary() ([]byte, error) }); ok {
			data, _ := m.MarshalBinary()
			n.h = h.mk()
			n.h.(interface{ UnmarshalBinary([]byte) error }).UnmarshalBinary(data)
			return n, true
		}
		panic(newError("TypeError", "cannot copy this hash"))
	}
	return nil, false
}

func init() {
	RegisterNative("hashlib.new", func(args []any, kw []KW) any {
		name := strings.ToLower(Str(args[0]))
		name = strings.ReplaceAll(name, "-", "_")
		mk, ok := hashMakers[name]
		if !ok {
			panic(valueError("unsupported hash type " + Str(args[0])))
		}
		h := &Hash2{name: name, h: mk.mk(), mk: mk.mk, block: mk.block}
		if len(args) > 1 && !IsNone(args[1]) {
			h.DynMethod("update", []any{args[1]}, nil)
		}
		return h
	})
	RegisterNative("hashlib.algorithms", func(args []any, kw []KW) any {
		var out []string
		for n := range hashMakers {
			out = append(out, n)
		}
		sortStrings(out)
		return strList(out)
	})
}

func sortStrings(xs []string) {
	for i := 1; i < len(xs); i++ {
		for j := i; j > 0 && xs[j] < xs[j-1]; j-- {
			xs[j], xs[j-1] = xs[j-1], xs[j]
		}
	}
}
