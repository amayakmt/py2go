package runtime

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"hash/adler32"
	"hash/crc32"
	"io"
	"strconv"
)

// zlib on top of Go's compress packages. The streams decompress as zlib's do;
// the compressed bytes are not zlib's (the match finder differs).

const zlibErr = "zlib_pm_error"

func zlibError(code int, where, msg string) *PyError {
	text := "Error " + strconv.Itoa(code) + " while " + where
	if msg != "" {
		text += ": " + msg
	}
	return newError(zlibErr, text)
}

func zlibLevel(level int) int {
	switch {
	case level == -1:
		return flate.DefaultCompression
	case level >= 0 && level <= 9:
		return level
	}
	panic(newError(zlibErr, "Bad compression level"))
}

func zlibCompress(data []byte, level, wbits int) []byte {
	var buf bytes.Buffer
	lv := zlibLevel(level)
	switch {
	case wbits < 0:
		w, _ := flate.NewWriter(&buf, lv)
		w.Write(data)
		w.Close()
	case wbits > 15:
		w, _ := gzip.NewWriterLevel(&buf, lv)
		w.Write(data)
		w.Close()
	default:
		w, _ := zlib.NewWriterLevel(&buf, lv)
		w.Write(data)
		w.Close()
	}
	return buf.Bytes()
}

func zlibDecodeError(err error, where string) *PyError {
	switch {
	case errors.Is(err, zlib.ErrHeader), errors.Is(err, gzip.ErrHeader):
		return zlibError(-3, where, "incorrect header check")
	case errors.Is(err, zlib.ErrChecksum), errors.Is(err, gzip.ErrChecksum):
		return zlibError(-3, where, "incorrect data check")
	case errors.Is(err, io.ErrUnexpectedEOF), errors.Is(err, io.EOF):
		return zlibError(-5, where, "incomplete or truncated stream")
	case errors.Is(err, zlib.ErrDictionary):
		return zlibError(2, where, "")
	}
	var ce flate.CorruptInputError
	if errors.As(err, &ce) {
		return zlibError(-3, where, "invalid stored block lengths")
	}
	return zlibError(-3, where, err.Error())
}

// zlibDecompress returns the data, what was left over after the end of the
// stream, and whether the stream ended.
func zlibDecompress(data []byte, wbits int) (out, unused []byte, eof bool, err error) {
	src := bytes.NewReader(data)
	var r io.Reader
	switch {
	case wbits < 0:
		r = flate.NewReader(src)
	case wbits >= 32:
		// zlib or gzip, by the header
		if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
			gr, e := gzip.NewReader(src)
			if e != nil {
				return nil, nil, false, e
			}
			gr.Multistream(false)
			r = gr
		} else {
			zr, e := zlib.NewReader(src)
			if e != nil {
				return nil, nil, false, e
			}
			r = zr
		}
	case wbits > 15:
		gr, e := gzip.NewReader(src)
		if e != nil {
			return nil, nil, false, e
		}
		gr.Multistream(false)
		r = gr
	default:
		zr, e := zlib.NewReader(src)
		if e != nil {
			return nil, nil, false, e
		}
		r = zr
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err == nil {
		eof = true
		unused = data[len(data)-src.Len():]
		if br, ok := r.(interface{ Close() error }); ok {
			br.Close()
		}
	}
	return buf.Bytes(), unused, eof, err
}

type zlibCompressObj struct {
	level, wbits int
	buf          bytes.Buffer
	w            interface {
		io.Writer
		Flush() error
		Close() error
	}
	done bool
}

func (z *zlibCompressObj) PyClassName() string { return "Compress" }
func (z *zlibCompressObj) PyRepr() string      { return "<zlib.Compress object at " + addr(z) + ">" }

var zlibCompressNames = names("compress flush copy")

func (z *zlibCompressObj) PyDynTable() map[string]bool { return zlibCompressNames }

func (z *zlibCompressObj) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "compress":
		if z.done {
			panic(zlibError(-2, "compressing data", "inconsistent stream state"))
		}
		z.w.Write([]byte(bytesArg(argAt(args, 0))))
		out := append([]byte(nil), z.buf.Bytes()...)
		z.buf.Reset()
		return Bytes(out), true
	case "flush":
		mode := int64(4)
		if len(args) > 0 {
			mode = DynInt(args[0])
		}
		if z.done {
			if mode == 4 {
				panic(zlibError(-2, "flushing", "inconsistent stream state"))
			}
			return Bytes(""), true
		}
		if mode == 4 {
			z.w.Close()
			z.done = true
		} else if mode != 0 {
			z.w.Flush()
		}
		out := append([]byte(nil), z.buf.Bytes()...)
		z.buf.Reset()
		return Bytes(out), true
	}
	return nil, false
}

func newZlibCompressObj(level, wbits int) *zlibCompressObj {
	z := &zlibCompressObj{level: level, wbits: wbits}
	lv := zlibLevel(level)
	switch {
	case wbits < 0:
		w, _ := flate.NewWriter(&z.buf, lv)
		z.w = w
	case wbits > 15:
		w, _ := gzip.NewWriterLevel(&z.buf, lv)
		z.w = w
	default:
		w, _ := zlib.NewWriterLevel(&z.buf, lv)
		z.w = w
	}
	return z
}

type zlibDecompressObj struct {
	wbits    int
	input    []byte
	returned int
	eof      bool
	unused   []byte
}

func (z *zlibDecompressObj) PyClassName() string { return "Decompress" }
func (z *zlibDecompressObj) PyRepr() string      { return "<zlib.Decompress object at " + addr(z) + ">" }

var zlibDecompressNames = names("decompress flush copy")

func (z *zlibDecompressObj) PyDynTable() map[string]bool { return zlibDecompressNames }

func (z *zlibDecompressObj) PyGetAttr(name string) (any, bool) {
	switch name {
	case "eof":
		return z.eof, true
	case "unused_data":
		return Bytes(z.unused), true
	case "unconsumed_tail":
		return Bytes(""), true
	}
	if zlibDecompressNames[name] {
		return func(args ...any) any {
			r, _ := z.DynMethod(name, args, nil)
			return r
		}, true
	}
	return nil, false
}

func (z *zlibDecompressObj) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "decompress":
		data := bytesArg(argAt(args, 0))
		if z.eof {
			z.unused = append(z.unused, data...)
			return Bytes(""), true
		}
		z.input = append(z.input, data...)
		out, unused, eof, err := zlibDecompress(z.input, z.wbits)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
			panic(zlibDecodeError(err, "decompressing data"))
		}
		res := out[z.returned:]
		z.returned = len(out)
		if eof {
			z.eof = true
			z.unused = append([]byte(nil), unused...)
		}
		return Bytes(append([]byte(nil), res...)), true
	case "flush":
		return Bytes(""), true
	}
	return nil, false
}

func init() {
	intArg := func(args []any, kw []KW, i int, name string, def int) int {
		v := def
		if i < len(args) {
			v = int(DynInt(args[i]))
		}
		for _, k := range kw {
			if k.Name == name {
				v = int(DynInt(k.Val))
			}
		}
		return v
	}
	RegisterNative("zlib.compress", func(args []any, kw []KW) any {
		if len(args) < 1 {
			panic(newError("TypeError", "compress() takes at least 1 argument (0 given)"))
		}
		return Bytes(zlibCompress(bytesArg(args[0]), intArg(args, kw, 1, "level", -1), intArg(args, kw, 2, "wbits", 15)))
	})
	RegisterNative("zlib.decompress", func(args []any, kw []KW) any {
		if len(args) < 1 {
			panic(newError("TypeError", "decompress() takes at least 1 argument (0 given)"))
		}
		out, _, eof, err := zlibDecompress(bytesArg(args[0]), intArg(args, kw, 1, "wbits", 15))
		if err != nil {
			panic(zlibDecodeError(err, "decompressing data"))
		}
		if !eof {
			panic(zlibError(-5, "decompressing data", "incomplete or truncated stream"))
		}
		return Bytes(out)
	})
	RegisterNative("zlib.crc32", func(args []any, kw []KW) any {
		nativeArgs("crc32", args, 1, 2)
		var crc uint32
		if len(args) > 1 {
			crc = uint32(DynInt(args[1]))
		}
		return int64(crc32.Update(crc, crc32.IEEETable, bytesArg(args[0])))
	})
	RegisterNative("zlib.adler32", func(args []any, kw []KW) any {
		nativeArgs("adler32", args, 1, 2)
		data := bytesArg(args[0])
		a, b := uint32(1), uint32(0)
		if len(args) > 1 {
			v := uint32(DynInt(args[1]))
			a, b = v&0xffff, v>>16
		}
		const mod = 65521
		for _, c := range data {
			a = (a + uint32(c)) % mod
			b = (b + a) % mod
		}
		_ = adler32.Size
		return int64(b<<16 | a)
	})
	RegisterNative("zlib.compressobj", func(args []any, kw []KW) any {
		return newZlibCompressObj(intArg(args, kw, 0, "level", -1), intArg(args, kw, 2, "wbits", 15))
	})
	RegisterNative("zlib.decompressobj", func(args []any, kw []KW) any {
		return &zlibDecompressObj{wbits: intArg(args, kw, 0, "wbits", 15)}
	})
}
