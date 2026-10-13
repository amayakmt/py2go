package runtime

import (
	"hash/crc32"
	"strconv"
	"strings"
)

// binascii: conversions between binary data and ASCII.

const (
	binErr        = "binascii_pm_Error"
	binIncomplete = "binascii_pm_Incomplete"
)

func binError(msg string) *PyError { return newError(binErr, msg) }

// asciiOrBytes reads a str of ASCII characters or a bytes-like object.
func asciiOrBytes(v any) []byte {
	v = Unbox(v)
	if s, ok := v.(string); ok {
		for i := 0; i < len(s); i++ {
			if s[i] >= 0x80 {
				panic(valueError("string argument should contain only ASCII characters"))
			}
		}
		return []byte(s)
	}
	b, ok := bytesOf(v)
	if !ok {
		panic(newError("TypeError", "argument should be bytes, buffer or ASCII string, not '"+TypeName(v)+"'"))
	}
	return []byte(b)
}

func bytesArg(v any) []byte {
	b, ok := bytesOf(v)
	if !ok {
		panic(newError("TypeError", "a bytes-like object is required, not '"+TypeName(Unbox(v))+"'"))
	}
	return []byte(b)
}

const hexDigits = "0123456789abcdef"

func hexlify(data []byte, sep []byte, per int) []byte {
	if len(sep) == 0 || per == 0 || len(data) == 0 {
		out := make([]byte, 0, len(data)*2)
		for _, c := range data {
			out = append(out, hexDigits[c>>4], hexDigits[c&15])
		}
		return out
	}
	// groups are counted from the right when per > 0, from the left when per < 0
	n := per
	if n < 0 {
		n = -n
	}
	var out []byte
	for i := range data {
		if i > 0 {
			if per > 0 && (len(data)-i)%n == 0 || per < 0 && i%n == 0 {
				out = append(out, sep...)
			}
		}
		out = append(out, hexDigits[data[i]>>4], hexDigits[data[i]&15])
	}
	return out
}

func unhexDigit(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

const b64chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

var b64rev [256]int8

func init() {
	for i := range b64rev {
		b64rev[i] = -1
	}
	for i := 0; i < 64; i++ {
		b64rev[b64chars[i]] = int8(i)
	}
}

func a2bBase64(data []byte, strict bool) []byte {
	out := []byte{}
	quad, left, pads := 0, 0, 0
	paddingStarted := false
	if strict && len(data) > 0 && data[0] == '=' {
		panic(binError("Leading padding not allowed"))
	}
	for i, c := range data {
		if c == '=' {
			paddingStarted = true
			if quad >= 2 {
				pads++
				if quad+pads >= 4 {
					if strict && i+1 < len(data) {
						panic(binError("Excess data after padding"))
					}
					return out
				}
			}
			continue
		}
		v := b64rev[c]
		if v < 0 {
			if strict {
				panic(binError("Only base64 data is allowed"))
			}
			continue
		}
		if strict && paddingStarted {
			panic(binError("Discontinuous padding not allowed"))
		}
		pads = 0
		switch quad {
		case 0:
			quad, left = 1, int(v)
		case 1:
			quad = 2
			out = append(out, byte(left<<2|int(v)>>4))
			left = int(v) & 0x0f
		case 2:
			quad = 3
			out = append(out, byte(left<<4|int(v)>>2))
			left = int(v) & 0x03
		case 3:
			quad = 0
			out = append(out, byte(left<<6|int(v)))
			left = 0
		}
	}
	if quad != 0 {
		if quad == 1 {
			panic(binError("Invalid base64-encoded string: number of data characters (" + strconv.Itoa(len(out)/3*4+1) + ") cannot be 1 more than a multiple of 4"))
		}
		panic(binError("Incorrect padding"))
	}
	return out
}

func b2aBase64(data []byte, newline bool) []byte {
	out := make([]byte, 0, (len(data)+2)/3*4+1)
	for i := 0; i < len(data); i += 3 {
		var chunk [3]byte
		n := copy(chunk[:], data[i:])
		v := int(chunk[0])<<16 | int(chunk[1])<<8 | int(chunk[2])
		out = append(out, b64chars[v>>18&63], b64chars[v>>12&63])
		if n > 1 {
			out = append(out, b64chars[v>>6&63])
		} else {
			out = append(out, '=')
		}
		if n > 2 {
			out = append(out, b64chars[v&63])
		} else {
			out = append(out, '=')
		}
	}
	if newline {
		out = append(out, '\n')
	}
	return out
}

func b2aUU(data []byte, backtick bool) []byte {
	if len(data) > 45 {
		panic(binError("At most 45 bytes at once"))
	}
	enc := func(v int) byte {
		if v == 0 && backtick {
			return '`'
		}
		return byte(v + ' ')
	}
	out := []byte{enc(len(data))}
	for i := 0; i < len(data); i += 3 {
		var chunk [3]byte
		copy(chunk[:], data[i:])
		v := int(chunk[0])<<16 | int(chunk[1])<<8 | int(chunk[2])
		out = append(out, enc(v>>18&63), enc(v>>12&63), enc(v>>6&63), enc(v&63))
	}
	return append(out, '\n')
}

func a2bUU(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	n := int(data[0]-' ') & 077
	data = data[1:]
	var out []byte
	acc, bits := 0, 0
	for len(out) < n && len(data) > 0 {
		c := data[0]
		data = data[1:]
		if c == '\n' || c == '\r' {
			c = ' '
		}
		if c < ' ' || c > ' '+64 {
			panic(binError("Illegal char"))
		}
		acc = acc<<6 | (int(c-' ') & 077)
		bits += 6
		if bits >= 8 {
			bits -= 8
			out = append(out, byte(acc>>bits))
			acc &= 1<<bits - 1
		}
	}
	for len(out) < n {
		out = append(out, 0)
	}
	out = out[:n]
	for _, c := range data {
		if c != ' ' && c != ' '+64 && c != '\n' && c != '\r' {
			panic(binError("Trailing garbage"))
		}
	}
	return out
}

func b2aQP(data []byte, quotetabs, istext, header bool) []byte {
	const maxLine = 76
	crlf := false
	if i := indexByte(data, '\n'); i > 0 && data[i-1] == '\r' {
		crlf = true
	}
	var out []byte
	hex := func(c byte) { out = append(out, "0123456789ABCDEF"[c>>4], "0123456789ABCDEF"[c&15]) }
	eol := func() {
		if crlf {
			out = append(out, '\r')
		}
		out = append(out, '\n')
	}
	in, linelen := 0, 0
	n := len(data)
	for in < n {
		c := data[in]
		special := c > 126 || c == '=' || header && c == '_' ||
			c == '.' && linelen == 0 && (in+1 == n || data[in+1] == '\n' || data[in+1] == '\r' || data[in+1] == 0) ||
			!istext && (c == '\r' || c == '\n') ||
			(c == '\t' || c == ' ') && in+1 == n ||
			c < 33 && c != '\r' && c != '\n' && (quotetabs || c != '\t' && c != ' ')
		if special {
			if linelen+3 >= maxLine {
				out = append(out, '=')
				eol()
				linelen = 0
			}
			out = append(out, '=')
			hex(c)
			in++
			linelen += 3
			continue
		}
		if istext && (c == '\n' || in+1 < n && c == '\r' && data[in+1] == '\n') {
			linelen = 0
			if len(out) > 0 && (out[len(out)-1] == ' ' || out[len(out)-1] == '\t') {
				ch := out[len(out)-1]
				out[len(out)-1] = '='
				hex(ch)
			}
			eol()
			if c == '\r' {
				in += 2
			} else {
				in++
			}
			continue
		}
		if in+1 != n && data[in+1] != '\n' && linelen+1 >= maxLine {
			out = append(out, '=')
			eol()
			linelen = 0
		}
		linelen++
		if header && c == ' ' {
			out = append(out, '_')
		} else {
			out = append(out, c)
		}
		in++
	}
	return out
}

func indexByte(b []byte, c byte) int {
	for i, x := range b {
		if x == c {
			return i
		}
	}
	return -1
}

func a2bQP(data []byte, header bool) []byte {
	var out []byte
	n := len(data)
	for in := 0; in < n; {
		c := data[in]
		switch {
		case c == '=':
			in++
			if in >= n {
				return out
			}
			if data[in] == '\n' || data[in] == '\r' {
				if data[in] != '\n' {
					for in < n && data[in] != '\n' {
						in++
					}
				}
				if in < n {
					in++
				}
			} else if data[in] == '=' {
				out = append(out, '=')
				in++
			} else if in+1 < n && unhexDigit(data[in]) >= 0 && unhexDigit(data[in+1]) >= 0 {
				out = append(out, byte(unhexDigit(data[in])<<4|unhexDigit(data[in+1])))
				in += 2
			} else {
				out = append(out, '=')
			}
		case header && c == '_':
			out = append(out, ' ')
			in++
		default:
			out = append(out, c)
			in++
		}
	}
	return out
}

func crcHqx(data []byte, crc int64) int64 {
	c := uint16(crc)
	for _, b := range data {
		c ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if c&0x8000 != 0 {
				c = c<<1 ^ 0x1021
			} else {
				c <<= 1
			}
		}
	}
	return int64(c)
}

func init() {
	hexFn := func(name string) {
		RegisterNative("binascii."+name, func(args []any, kw []KW) any {
			if len(args) < 1 || len(args) > 3 {
				panic(newError("TypeError", name+"() takes 1 to 3 arguments"))
			}
			data := bytesArg(args[0])
			var sep []byte
			per := 1
			vals := []any{nil, nil}
			copy(vals, args[1:])
			for _, k := range kw {
				switch k.Name {
				case "sep":
					vals[0] = k.Val
				case "bytes_per_sep":
					vals[1] = k.Val
				}
			}
			if vals[0] != nil {
				sv := Unbox(vals[0])
				switch x := sv.(type) {
				case string:
					sep = []byte(x)
				default:
					sep = bytesArg(sv)
				}
				if len(sep) != 1 {
					panic(valueError("sep must be length 1."))
				}
			}
			if vals[1] != nil {
				per = int(DynInt(vals[1]))
			}
			return Bytes(hexlify(data, sep, per))
		})
	}
	hexFn("hexlify")
	hexFn("b2a_hex")
	unhex := func(name string) {
		RegisterNative("binascii."+name, func(args []any, kw []KW) any {
			nativeArgs(name, args, 1, 1)
			d := asciiOrBytes(args[0])
			if len(d)%2 != 0 {
				panic(binError("Odd-length string"))
			}
			out := make([]byte, len(d)/2)
			for i := range out {
				hi, lo := unhexDigit(d[2*i]), unhexDigit(d[2*i+1])
				if hi < 0 || lo < 0 {
					panic(binError("Non-hexadecimal digit found"))
				}
				out[i] = byte(hi<<4 | lo)
			}
			return Bytes(out)
		})
	}
	unhex("unhexlify")
	unhex("a2b_hex")
	RegisterNative("binascii.a2b_base64", func(args []any, kw []KW) any {
		nativeArgs("a2b_base64", args, 1, 2)
		strict := false
		if len(args) > 1 {
			strict = DynBool(args[1])
		}
		for _, k := range kw {
			if k.Name == "strict_mode" {
				strict = DynBool(k.Val)
			}
		}
		return Bytes(a2bBase64(asciiOrBytes(args[0]), strict))
	})
	RegisterNative("binascii.b2a_base64", func(args []any, kw []KW) any {
		nativeArgs("b2a_base64", args, 1, 2)
		nl := true
		if len(args) > 1 {
			nl = DynBool(args[1])
		}
		for _, k := range kw {
			if k.Name == "newline" {
				nl = DynBool(k.Val)
			}
		}
		return Bytes(b2aBase64(bytesArg(args[0]), nl))
	})
	RegisterNative("binascii.crc32", func(args []any, kw []KW) any {
		nativeArgs("crc32", args, 1, 2)
		var crc uint32
		if len(args) > 1 {
			crc = uint32(DynInt(args[1]))
		}
		return int64(crc32.Update(crc, crc32.IEEETable, bytesArg(args[0])))
	})
	RegisterNative("binascii.crc_hqx", func(args []any, kw []KW) any {
		nativeArgs("crc_hqx", args, 2, 2)
		return crcHqx(bytesArg(args[0]), DynInt(args[1]))
	})
	RegisterNative("binascii.b2a_uu", func(args []any, kw []KW) any {
		nativeArgs("b2a_uu", args, 1, 2)
		bt := true
		if len(args) > 1 {
			bt = DynBool(args[1])
		}
		for _, k := range kw {
			if k.Name == "backtick" {
				bt = DynBool(k.Val)
			}
		}
		return Bytes(b2aUU(bytesArg(args[0]), bt))
	})
	RegisterNative("binascii.a2b_uu", func(args []any, kw []KW) any {
		nativeArgs("a2b_uu", args, 1, 1)
		return Bytes(a2bUU(asciiOrBytes(args[0])))
	})
	kwBool := func(args []any, kw []KW, names ...string) []bool {
		out := make([]bool, len(names))
		for i := range names {
			if i+1 < len(args) {
				out[i] = DynBool(args[i+1])
			}
		}
		for _, k := range kw {
			for i, n := range names {
				if k.Name == n {
					out[i] = DynBool(k.Val)
				}
			}
		}
		return out
	}
	RegisterNative("binascii.b2a_qp", func(args []any, kw []KW) any {
		o := kwBool(args, kw, "quotetabs", "istext", "header")
		istext := true
		set := false
		for _, k := range kw {
			if k.Name == "istext" {
				set = true
			}
		}
		if len(args) > 2 || set {
			istext = o[1]
		}
		return Bytes(b2aQP(bytesArg(args[0]), o[0], istext, o[2]))
	})
	RegisterNative("binascii.a2b_qp", func(args []any, kw []KW) any {
		o := kwBool(args, kw, "header")
		return Bytes(a2bQP(asciiOrBytes(args[0]), o[0]))
	})
	_ = strings.Join
}
