package runtime

import (
	"errors"
	"os"
	"strings"
	"syscall"
	"unicode/utf8"
	"unsafe"
)

// File is what open() returns, and sys.stdout, sys.stderr and sys.stdin: a
// buffered binary file over a descriptor (BufferedReader, BufferedWriter,
// BufferedRandom or, unbuffered, FileIO), or a text file (TextIOWrapper)
// over one. The text file keeps its binary file in inner, which is what its
// .buffer attribute is, and does all its I/O through it.
type File struct {
	inner *File // the buffer of a text file

	fd      int
	name    any // a str, bytes or the int descriptor the file was opened from
	mode    string
	rd, wr  bool
	app     bool
	closed  bool
	closefd bool
	std     bool // the standard streams stay open when closed
	rawIO   bool // buffering=0: a FileIO
	tty     bool
	isOut   bool // sys.stdout: flushed before anything reads standard input

	rbuf   []byte
	rpos   int
	wbuf   []byte
	bufMax int

	// text files
	text         bool
	encoding     string
	codec        string
	errs         string
	nlMode       int // nlUniversal, nlEmpty, nlLF, nlCR or nlCRLF
	lineBuf      bool
	writeThrough bool
	seenNL       int
	bom          bool   // a utf-8-sig file still has to skip a BOM
	pending      []byte // encoded text not yet handed to the buffer
	given        string // the encoding name as the program spelled it
	slurped      bool   // a stateful encoding was decoded as a whole
	slurp        []rune
	slurpAt      int
}

const (
	nlUniversal = iota // newline=None: any ending reads as \n
	nlEmpty            // newline="": any ending, untranslated
	nlLF
	nlCR
	nlCRLF
)

const defaultBufferSize = 8192

var openFiles = map[*File]struct{}{}

// Stdout, Stderr and Stdin are the standard streams as file objects.
var (
	Stdout = newStdStream(1, "<stdout>", "w")
	Stderr = newStdStream(2, "<stderr>", "w")
	Stdin  = newStdStream(0, "<stdin>", "r")
)

func newStdStream(fd int, name, mode string) *File {
	inner := &File{fd: fd, name: int64(fd), mode: mode + "b", rd: mode == "r", wr: mode == "w", std: true, bufMax: defaultBufferSize}
	inner.tty = isatty(fd)
	f := &File{inner: inner, fd: fd, name: name, mode: mode, rd: inner.rd, wr: inner.wr, std: true, text: true,
		encoding: "utf-8", codec: "utf-8", given: "utf-8", errs: "surrogateescape", nlMode: nlUniversal}
	if fd == 2 {
		f.errs = "backslashreplace"
	}
	f.isOut = fd == 1
	if mode == "w" {
		f.lineBuf = inner.tty || fd == 2
		f.writeThrough = fd == 2
		if inner.tty || fd == 2 {
			inner.bufMax = 0
		}
	}
	return f
}

func isatty(fd int) bool {
	var t syscall.Termios
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(ioctlGetTermios), uintptr(unsafe.Pointer(&t)))
	return e == 0
}

// FlushAll writes what the open files still hold: what CPython does when the
// interpreter finalizes. The standard streams come last.
func FlushAll() {
	for f := range openFiles {
		func() {
			defer func() { recover() }()
			f.flush()
		}()
	}
	for _, f := range []*File{Stdout, Stderr} {
		func() {
			defer func() { recover() }()
			f.flush()
		}()
	}
}

func (f *File) PyClassName() string {
	switch {
	case f.text:
		return "TextIOWrapper"
	case f.rawIO:
		return "FileIO"
	case f.rd && f.wr:
		return "BufferedRandom"
	case f.wr:
		return "BufferedWriter"
	}
	return "BufferedReader"
}

func (f *File) PyRepr() string {
	name := ""
	if f.name != nil {
		name = " name=" + Repr(f.name)
	}
	switch {
	case f.text:
		return "<_io.TextIOWrapper" + name + " mode=" + quote(f.mode) + " encoding=" + quote(f.given) + ">"
	case f.rawIO:
		cf := "True"
		if !f.closefd {
			cf = "False"
		}
		return "<_io.FileIO" + name + " mode=" + quote(f.mode) + " closefd=" + cf + ">"
	}
	return "<_io." + f.PyClassName() + name + ">"
}

var fileMethodNames = names("read read1 readline readlines readinto readinto1 peek write writelines close flush seek tell truncate fileno isatty readable writable seekable detach __enter__ __exit__ __iter__ __next__")

func (f *File) PyDynTable() map[string]bool { return fileMethodNames }

func (f *File) PyAttrNames() []string { return []string{"name", "mode"} }

func (f *File) PyGetAttr(name string) (any, bool) {
	switch name {
	case "name":
		return f.name, true
	case "mode":
		return f.mode, true
	case "closed":
		return f.isClosed(), true
	case "closefd":
		if f.inner != nil {
			return f.inner.closefd, true
		}
		return f.closefd, true
	case "encoding":
		if f.text {
			return f.given, true
		}
		return nil, false
	case "errors":
		if f.text {
			return f.errs, true
		}
		return nil, false
	case "newlines":
		if !f.text {
			return nil, false
		}
		return f.newlinesAttr(), true
	case "buffer":
		if f.inner != nil {
			return f.inner, true
		}
		return nil, false
	case "raw":
		if !f.text && !f.rawIO {
			return &File{fd: f.fd, name: f.name, mode: strings.ReplaceAll(strings.ReplaceAll(f.mode, "b", ""), "t", "") + "b", rd: f.rd, wr: f.wr, rawIO: true, bufMax: defaultBufferSize}, true
		}
		return nil, false
	case "line_buffering":
		return f.lineBuf, f.text
	case "write_through":
		return f.writeThrough, f.text
	case "_CHUNK_SIZE":
		return int64(defaultBufferSize), f.text
	}
	if fileMethodNames[name] {
		return func(args ...any) any {
			r, _ := f.DynMethod(name, args, nil)
			return r
		}, true
	}
	return nil, false
}

// PySetAttr lets the program name a file whose name it knows better, as
// tempfile does for the descriptor it opened.
func (f *File) PySetAttr(name string, v any) bool {
	if name != "name" {
		return false
	}
	f.name = v
	if f.inner != nil {
		f.inner.name = v
	}
	return true
}

func (f *File) isClosed() bool {
	if f.inner != nil {
		return f.inner.closed
	}
	return f.closed
}

func (f *File) newlinesAttr() any {
	if f.nlMode != nlUniversal && f.nlMode != nlEmpty {
		return nil
	}
	switch f.seenNL {
	case 0:
		return nil
	case 1:
		return "\n"
	case 2:
		return "\r"
	case 4:
		return "\r\n"
	}
	var parts []any
	if f.seenNL&2 != 0 {
		parts = append(parts, "\r")
	}
	if f.seenNL&1 != 0 {
		parts = append(parts, "\n")
	}
	if f.seenNL&4 != 0 {
		parts = append(parts, "\r\n")
	}
	return NewVarTuple(parts...)
}

func unsupported(msg string) *PyError { return newError("io_pm_UnsupportedOperation", msg) }

func (f *File) check(op string) {
	if !f.isClosed() {
		return
	}
	if !f.text && op == "read" {
		panic(valueError("read of closed file"))
	}
	if !f.text && op == "write" {
		panic(valueError("write to closed file"))
	}
	panic(valueError("I/O operation on closed file."))
}

// ---- the binary layer ----

func (f *File) unread() int { return len(f.rbuf) - f.rpos }

// rawRead reads from the descriptor, letting other threads run meanwhile.
func rawRead(fd int, p []byte) (int, error) {
	var n int
	var err error
	Blocking(func() {
		for {
			n, err = syscall.Read(fd, p)
			if err != syscall.EINTR {
				break
			}
		}
	})
	return n, err
}

func rawWrite(fd int, p []byte) (int, error) {
	var n int
	var err error
	Blocking(func() {
		for {
			n, err = syscall.Write(fd, p)
			if err != syscall.EINTR {
				break
			}
		}
	})
	return n, err
}

// fill reads more into the read buffer; false at the end of the file.
func (f *File) fill() bool {
	if f.isOut || f.fd == 0 {
		Stdout.flush()
	}
	if len(f.wbuf) > 0 {
		f.flushBuf()
	}
	if f.rpos > 0 && f.rpos == len(f.rbuf) {
		f.rbuf, f.rpos = f.rbuf[:0], 0
	} else if f.rpos > 4096 {
		f.rbuf = append(f.rbuf[:0], f.rbuf[f.rpos:]...)
		f.rpos = 0
	}
	size := f.bufMax
	if size < defaultBufferSize {
		size = defaultBufferSize
	}
	tmp := make([]byte, size)
	n, err := rawRead(f.fd, tmp)
	if err != nil {
		panic(f.sysErr(err))
	}
	if n == 0 {
		return false
	}
	f.rbuf = append(f.rbuf, tmp[:n]...)
	return true
}

func (f *File) sysErr(err error) *PyError {
	return sysError(err)
}

// flushBuf writes the pending bytes.
func (f *File) flushBuf() {
	for len(f.wbuf) > 0 {
		n, err := rawWrite(f.fd, f.wbuf)
		if err != nil {
			f.wbuf = f.wbuf[:0]
			panic(f.sysErr(err))
		}
		f.wbuf = f.wbuf[n:]
	}
	f.wbuf = f.wbuf[:0]
}

// syncRead gives back the bytes read ahead, so that a write or a seek
// happens where the program thinks the position is.
func (f *File) syncRead() {
	if n := f.unread(); n > 0 {
		syscall.Seek(f.fd, int64(-n), 1)
	}
	f.rbuf, f.rpos = f.rbuf[:0], 0
}

func (f *File) readBytes(n int) []byte {
	if f.inner != nil {
		return f.inner.readBytes(n)
	}
	if !f.rd {
		if f.rawIO {
			panic(unsupported("File not open for reading"))
		}
		panic(unsupported("read"))
	}
	if f.rawIO {
		if n < 0 {
			var out []byte
			for {
				tmp := make([]byte, 65536)
				got, err := rawRead(f.fd, tmp)
				if err != nil {
					panic(f.sysErr(err))
				}
				if got == 0 {
					return out
				}
				out = append(out, tmp[:got]...)
			}
		}
		tmp := make([]byte, n)
		got, err := rawRead(f.fd, tmp)
		if err != nil {
			panic(f.sysErr(err))
		}
		return tmp[:got]
	}
	for n < 0 || f.unread() < n {
		if !f.fill() {
			break
		}
	}
	if n < 0 || n > f.unread() {
		n = f.unread()
	}
	out := append([]byte(nil), f.rbuf[f.rpos:f.rpos+n]...)
	f.rpos += n
	return out
}

// readByteSlice is read1: at most one system call.
func (f *File) read1(n int) []byte {
	if f.unread() == 0 {
		if n == 0 || !f.fill() {
			return nil
		}
	}
	if n < 0 || n > f.unread() {
		n = f.unread()
	}
	out := append([]byte(nil), f.rbuf[f.rpos:f.rpos+n]...)
	f.rpos += n
	return out
}

func (f *File) readLineBytes(limit int) []byte {
	var out []byte
	for {
		avail := f.rbuf[f.rpos:]
		if len(avail) == 0 {
			if !f.fill() {
				return out
			}
			continue
		}
		if limit >= 0 && len(out)+len(avail) > limit {
			avail = avail[:limit-len(out)]
		}
		if i := indexByte(avail, '\n'); i >= 0 {
			out = append(out, avail[:i+1]...)
			f.rpos += i + 1
			return out
		}
		out = append(out, avail...)
		f.rpos += len(avail)
		if limit >= 0 && len(out) >= limit {
			return out
		}
	}
}

func (f *File) writeBytes(data []byte) int {
	if f.inner != nil {
		return f.inner.writeBytes(data)
	}
	if !f.wr {
		if f.rawIO {
			panic(unsupported("File not open for writing"))
		}
		panic(unsupported("write"))
	}
	if f.rawIO {
		n, err := rawWrite(f.fd, data)
		if err != nil {
			panic(f.sysErr(err))
		}
		return n
	}
	if f.unread() > 0 || len(f.rbuf) > 0 {
		f.syncRead()
	}
	if f.app && f.rd {
		// appending to a file opened a+: writes go to the end
		syscall.Seek(f.fd, 0, 2)
	}
	f.wbuf = append(f.wbuf, data...)
	if len(f.wbuf) >= f.bufMax {
		f.flushBuf()
	}
	return len(data)
}

// pushPending hands the encoded text a text file holds to its buffer.
func (f *File) pushPending() {
	if len(f.pending) > 0 {
		data := f.pending
		f.pending = nil
		f.inner.writeBytes(data)
	}
}

func (f *File) flush() {
	if f.inner != nil {
		f.pushPending()
		f.inner.flush()
		return
	}
	if len(f.wbuf) > 0 {
		f.flushBuf()
	}
}

func (f *File) seekRaw(off int64, whence int) int64 {
	if f.inner != nil {
		return f.inner.seekRaw(off, whence)
	}
	f.flush()
	if whence == 1 {
		off -= int64(f.unread())
	}
	f.rbuf, f.rpos = f.rbuf[:0], 0
	pos, err := syscall.Seek(f.fd, off, whence)
	if err != nil {
		panic(f.sysErr(err))
	}
	return pos
}

func (f *File) tellRaw() int64 {
	if f.inner != nil {
		return f.inner.tellRaw()
	}
	pos, err := syscall.Seek(f.fd, 0, 1)
	if err != nil {
		panic(f.sysErr(err))
	}
	return pos - int64(f.unread()) + int64(len(f.wbuf))
}

func (f *File) seekable() bool {
	if f.inner != nil {
		return f.inner.seekable()
	}
	_, err := syscall.Seek(f.fd, 0, 1)
	return err == nil
}

func (f *File) closeFile() {
	if f.isClosed() {
		return
	}
	if f.inner != nil {
		var err any
		func() {
			defer func() { err = recover() }()
			f.pushPending()
			f.inner.closeFile()
		}()
		f.closed = true
		delete(openFiles, f)
		if err != nil {
			panic(err)
		}
		return
	}
	var ferr any
	func() {
		defer func() { ferr = recover() }()
		f.flush()
	}()
	f.closed = true
	delete(openFiles, f)
	if !f.std && f.closefd {
		syscall.Close(f.fd)
	}
	if ferr != nil {
		panic(ferr)
	}
}

// ---- the text layer ----

// charLen is the length in bytes of the character starting at b[0], by the
// lead byte (1 for anything that is not one).
func charLen(codec string, b byte) int {
	if codec != "utf-8" {
		return 1
	}
	switch {
	case b < 0xc2:
		return 1
	case b < 0xe0:
		return 2
	case b < 0xf0:
		return 3
	case b < 0xf5:
		return 4
	}
	return 1
}

func (f *File) stateless() bool {
	return f.codec == "utf-8" || f.codec == "ascii" || f.codec == "latin-1" || f.codec == "utf-8-sig"
}

func (f *File) decodeChunk(b []byte) string {
	if f.codec == "utf-8" && utf8.Valid(b) {
		return string(b)
	}
	return DecodeBytes(string(b), f.codec, f.errs)
}

// peekByte is the next byte, reading more when the buffer is empty.
func (in *File) peekByte() (byte, bool) {
	if in.unread() == 0 && !in.fill() {
		return 0, false
	}
	return in.rbuf[in.rpos], true
}

func (f *File) skipBOM() {
	if f.bom {
		f.bom = false
		in := f.inner
		for in.unread() < 3 {
			if !in.fill() {
				break
			}
		}
		if in.unread() >= 3 && string(in.rbuf[in.rpos:in.rpos+3]) == "\xef\xbb\xbf" {
			in.rpos += 3
		}
	}
}

// slurpAll decodes the rest of a file whose encoding keeps state as a whole.
func (f *File) slurpAll() {
	if f.slurped {
		return
	}
	data := f.inner.readBytes(-1)
	f.slurp = []rune(DecodeBytes(string(data), f.codec, f.errs))
	f.slurpAt = 0
	f.slurped = true
}

// readText is read(n): up to n characters, or all that is left for n < 0.
func (f *File) readText(n int) string {
	if !f.rd {
		panic(unsupported("not readable"))
	}
	f.pushPending()
	if f.stateful() {
		f.slurpAll()
		rest := f.slurp[f.slurpAt:]
		if n < 0 || n > len(rest) {
			n = len(rest)
		}
		out := string(rest[:n])
		f.slurpAt += n
		return f.translate(out)
	}
	f.skipBOM()
	in := f.inner
	if n < 0 {
		data := in.readBytes(-1)
		return f.translate(f.decodeChunk(data))
	}
	var sb strings.Builder
	got := 0
	for got < n {
		b, ok := in.peekByte()
		if !ok {
			break
		}
		if f.nlMode == nlUniversal && b == '\r' {
			in.rpos++
			if nb, ok := in.peekByte(); ok && nb == '\n' {
				in.rpos++
				f.seenNL |= 4
			} else {
				f.seenNL |= 2
			}
			sb.WriteByte('\n')
			got++
			continue
		}
		if b == '\n' {
			f.seenNL |= 1
		}
		l := charLen(f.codec, b)
		for in.unread() < l {
			if !in.fill() {
				break
			}
		}
		if l > in.unread() {
			l = in.unread()
		}
		sb.WriteString(f.decodeChunk(in.rbuf[in.rpos : in.rpos+l]))
		in.rpos += l
		got++
	}
	return sb.String()
}

func (f *File) stateful() bool { return !f.stateless() }

// translate applies the newline mode to decoded text.
func (f *File) translate(s string) string {
	if f.nlMode == nlUniversal && strings.IndexByte(s, '\r') >= 0 {
		if strings.Contains(s, "\r\n") {
			f.seenNL |= 4
		}
		s = strings.ReplaceAll(s, "\r\n", "\n")
		if strings.IndexByte(s, '\r') >= 0 {
			f.seenNL |= 2
			s = strings.ReplaceAll(s, "\r", "\n")
		}
	}
	if strings.IndexByte(s, '\n') >= 0 {
		f.seenNL |= 1
	}
	return s
}

// readLineText is readline(limit) of a text file.
func (f *File) readLineText(limit int) string {
	if !f.rd {
		panic(unsupported("not readable"))
	}
	f.pushPending()
	if f.stateful() {
		f.slurpAll()
		var sb strings.Builder
		for f.slurpAt < len(f.slurp) && (limit < 0 || sb.Len() < limit) {
			r := f.slurp[f.slurpAt]
			f.slurpAt++
			sb.WriteRune(r)
			if r == '\n' {
				break
			}
			if r == '\r' && f.nlMode != nlLF {
				if f.slurpAt < len(f.slurp) && f.slurp[f.slurpAt] == '\n' {
					f.slurpAt++
					sb.WriteRune('\n')
				}
				break
			}
		}
		return f.translate(sb.String())
	}
	f.skipBOM()
	in := f.inner
	var line []byte
	chars := 0
	for limit < 0 || chars < limit {
		b, ok := in.peekByte()
		if !ok {
			break
		}
		switch {
		case b == '\n' && f.nlMode != nlCR && f.nlMode != nlCRLF:
			in.rpos++
			line = append(line, '\n')
			return f.finishLine(line)
		case b == '\r' && (f.nlMode == nlUniversal || f.nlMode == nlEmpty):
			in.rpos++
			line = append(line, '\r')
			if nb, ok := in.peekByte(); ok && nb == '\n' {
				in.rpos++
				line = append(line, '\n')
			}
			return f.finishLine(line)
		case b == '\r' && f.nlMode == nlCR:
			in.rpos++
			line = append(line, '\r')
			return f.finishLine(line)
		case b == '\r' && f.nlMode == nlCRLF:
			in.rpos++
			line = append(line, '\r')
			if nb, ok := in.peekByte(); ok && nb == '\n' {
				in.rpos++
				line = append(line, '\n')
				return f.finishLine(line)
			}
			chars++
			continue
		}
		l := charLen(f.codec, b)
		for in.unread() < l {
			if !in.fill() {
				break
			}
		}
		if l > in.unread() {
			l = in.unread()
		}
		line = append(line, in.rbuf[in.rpos:in.rpos+l]...)
		in.rpos += l
		chars++
	}
	return f.finishLine(line)
}

func (f *File) finishLine(line []byte) string {
	s := f.decodeChunk(line)
	if f.nlMode == nlEmpty {
		switch {
		case strings.HasSuffix(s, "\r\n"):
			f.seenNL |= 4
		case strings.HasSuffix(s, "\n"):
			f.seenNL |= 1
		case strings.HasSuffix(s, "\r"):
			f.seenNL |= 2
		}
		return s
	}
	return f.translate(s)
}

// writeStr is writeText without the count of characters.
func (f *File) writeStr(s string) {
	if f.isClosed() {
		panic(valueError("I/O operation on closed file."))
	}
	if f.codec == "utf-8" && f.nlMode != nlCR && f.nlMode != nlCRLF && f.wr {
		f.queue([]byte(s), s)
		return
	}
	f.writeText(s)
}

// queue holds encoded text until a chunk is full, or the file is flushed
// or, when line buffered, a newline was written.
func (f *File) queue(data []byte, s string) {
	in := f.inner
	if in.unread() > 0 {
		in.syncRead()
	}
	f.pending = append(f.pending, data...)
	if f.writeThrough || len(f.pending) >= defaultBufferSize || (f.lineBuf && (strings.IndexByte(s, '\n') >= 0 || strings.IndexByte(s, '\r') >= 0)) {
		f.pushPending()
		if f.writeThrough || f.lineBuf {
			in.flush()
		}
	}
}

// writeText is write(s) of a text file.
func (f *File) writeText(s string) int {
	if !f.wr {
		panic(unsupported("not writable"))
	}
	n := StrLen(s)
	text := s
	switch f.nlMode {
	case nlCR:
		text = strings.ReplaceAll(text, "\n", "\r")
	case nlCRLF:
		text = strings.ReplaceAll(text, "\n", "\r\n")
	}
	var data []byte
	switch f.codec {
	case "utf-8":
		data = []byte(text)
	case "utf-8-sig":
		data = []byte(EncodeStr(text, "utf-8", f.errs))
	default:
		data = []byte(EncodeStr(text, f.codec, f.errs))
	}
	f.queue(data, s)
	return int(n)
}

// ---- iteration ----

func (f *File) IterAny() Cursor[any] {
	f.check("")
	return &fileCursor{f: f}
}

type fileCursor struct {
	f   *File
	cur any
}

func (c *fileCursor) Next() bool {
	c.f.check("")
	if c.f.text {
		line := c.f.readLineText(-1)
		if line == "" {
			return false
		}
		c.cur = line
		return true
	}
	line := c.f.readLineBytes(-1)
	if len(line) == 0 {
		return false
	}
	c.cur = Bytes(line)
	return true
}

func (c *fileCursor) Value() any { return c.cur }

func (f *File) PyTruth() bool { return true }

func bytesLike(v any) ([]byte, bool) {
	switch x := Unbox(v).(type) {
	case Bytes:
		return []byte(x), true
	case *ByteArray:
		return append([]byte(nil), x.b...), true
	}
	if s, ok := bytesOf(v); ok {
		return []byte(s), true
	}
	return nil, false
}

func (f *File) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "__enter__":
		f.check("")
		return f, true
	case "__exit__":
		f.closeFile()
		return nil, true
	case "close":
		f.closeFile()
		return nil, true
	case "flush":
		f.check("")
		f.flush()
		return nil, true
	case "fileno":
		f.check("")
		if f.inner != nil {
			return int64(f.inner.fd), true
		}
		return int64(f.fd), true
	case "isatty":
		f.check("")
		if f.inner != nil {
			return isatty(f.inner.fd), true
		}
		return isatty(f.fd), true
	case "readable":
		f.check("")
		return f.rd, true
	case "writable":
		f.check("")
		return f.wr, true
	case "seekable":
		f.check("")
		return f.seekable(), true
	case "detach":
		panic(unsupported("detach"))
	case "read", "read1":
		f.check("read")
		if !f.rd {
			switch {
			case f.text:
				panic(unsupported("not readable"))
			case f.rawIO:
				panic(unsupported("File not open for reading"))
			}
			panic(unsupported("read"))
		}
		n := -1
		if len(args) > 0 && !IsNone(args[0]) {
			n = int(DynInt(args[0]))
		}
		if f.text {
			return f.readText(n), true
		}
		if name == "read1" {
			return Bytes(f.read1(n)), true
		}
		return Bytes(f.readBytes(n)), true
	case "readinto", "readinto1":
		f.check("read")
		if f.text {
			return nil, false
		}
		dst, ok := Unbox(argAt(args, 0)).(*ByteArray)
		if !ok {
			panic(newError("TypeError", "readinto() argument must be read-write bytes-like object, not "+TypeName(argAt(args, 0))))
		}
		data := f.readBytes(int(dst.Len()))
		if name == "readinto1" {
			data = f.read1(int(dst.Len()))
		}
		copy(dst.b, data)
		return int64(len(data)), true
	case "peek":
		f.check("read")
		if f.text {
			return nil, false
		}
		if f.unread() == 0 {
			f.fill()
		}
		return Bytes(append([]byte(nil), f.rbuf[f.rpos:]...)), true
	case "readline":
		f.check("read")
		limit := -1
		if len(args) > 0 && !IsNone(args[0]) {
			limit = int(DynInt(args[0]))
		}
		if f.text {
			return f.readLineText(limit), true
		}
		if !f.rd {
			panic(unsupported("not readable"))
		}
		return Bytes(f.readLineBytes(limit)), true
	case "readlines":
		f.check("read")
		hint := int64(-1)
		if len(args) > 0 && !IsNone(args[0]) {
			hint = DynInt(args[0])
		}
		out := NewList[any]()
		total := int64(0)
		for {
			var item any
			if f.text {
				line := f.readLineText(-1)
				if line == "" {
					break
				}
				item = line
				total += StrLen(line)
			} else {
				line := f.readLineBytes(-1)
				if len(line) == 0 {
					break
				}
				item = Bytes(line)
				total += int64(len(line))
			}
			out.Append(item)
			if hint > 0 && total >= hint {
				break
			}
		}
		return out, true
	case "__next__":
		f.check("")
		if f.text {
			line := f.readLineText(-1)
			if line == "" {
				panic(newError("StopIteration", ""))
			}
			return line, true
		}
		line := f.readLineBytes(-1)
		if len(line) == 0 {
			panic(newError("StopIteration", ""))
		}
		return Bytes(line), true
	case "__iter__":
		f.check("")
		return f, true
	case "write":
		f.check("write")
		countArgs("write", args, 1, 1)
		if !f.wr {
			switch {
			case f.text:
				panic(unsupported("not writable"))
			case f.rawIO:
				panic(unsupported("File not open for writing"))
			}
			panic(unsupported("write"))
		}
		if f.text {
			s, ok := Unbox(args[0]).(string)
			if !ok {
				panic(newError("TypeError", "write() argument must be str, not "+TypeName(Unbox(args[0]))))
			}
			return int64(f.writeText(s)), true
		}
		data, ok := bytesLike(args[0])
		if !ok {
			panic(newError("TypeError", "a bytes-like object is required, not '"+TypeName(Unbox(args[0]))+"'"))
		}
		return int64(f.writeBytes(data)), true
	case "writelines":
		f.check("write")
		for c := DynIter(args[0]).Iter(); c.Next(); {
			f.DynMethod("write", []any{c.Value()}, nil)
		}
		return nil, true
	case "seek":
		f.check("")
		off := DynInt(argAt(args, 0))
		whence := int(argInt(args, 1, 0))
		if whence < 0 || whence > 2 {
			panic(valueError("invalid whence (" + itoa(whence) + ", should be 0, 1 or 2)"))
		}
		if f.text {
			f.pushPending()
			if whence != 0 && off != 0 {
				if whence == 1 {
					panic(unsupported("can't do nonzero cur-relative seeks"))
				}
				panic(unsupported("can't do nonzero end-relative seeks"))
			}
			if whence == 0 && off < 0 {
				panic(valueError("negative seek position " + Str(off)))
			}
			f.slurped = false
			f.slurpAt = 0
			if whence == 1 && f.stateful() {
				return f.inner.tellRaw(), true
			}
			f.bom = false
			return f.inner.seekRaw(off, whence), true
		}
		if f.rawIO && f.wr {
			pos, err := syscall.Seek(f.fd, off, whence)
			if err != nil {
				panic(f.sysErr(err))
			}
			return pos, true
		}
		return f.seekRaw(off, whence), true
	case "tell":
		f.check("")
		if f.text {
			f.pushPending()
			return f.inner.tellRaw(), true
		}
		if f.rawIO {
			pos, err := syscall.Seek(f.fd, 0, 1)
			if err != nil {
				panic(f.sysErr(err))
			}
			return pos, true
		}
		return f.tellRaw(), true
	case "truncate":
		f.check("")
		in := f
		if f.inner != nil {
			f.pushPending()
			in = f.inner
		}
		if !in.wr {
			panic(unsupported("File not open for writing"))
		}
		in.flush()
		size := int64(0)
		if len(args) > 0 && !IsNone(args[0]) {
			size = DynInt(args[0])
		} else {
			size = in.tellRaw()
		}
		if err := syscall.Ftruncate(in.fd, size); err != nil {
			panic(in.sysErr(err))
		}
		return size, true
	}
	return nil, false
}

// ---- open ----

type openArgs struct {
	file      any
	mode      string
	buffering int
	encoding  any
	errors    any
	newline   any
	closefd   bool
}

func parseOpenArgs(args []any, kw []KW) openArgs {
	o := openArgs{mode: "r", buffering: -1, closefd: true}
	names := []string{"file", "mode", "buffering", "encoding", "errors", "newline", "closefd", "opener"}
	vals := make([]any, len(names))
	set := make([]bool, len(names))
	if len(args) > len(names) {
		panic(newError("TypeError", "open() takes at most 8 arguments ("+itoa(len(args))+" given)"))
	}
	for i, a := range args {
		vals[i], set[i] = a, true
	}
	for _, k := range kw {
		found := false
		for i, n := range names {
			if n == k.Name {
				if set[i] {
					panic(newError("TypeError", "argument for open() given by name ('"+n+"') and position ("+itoa(i+1)+")"))
				}
				vals[i], set[i], found = k.Val, true, true
			}
		}
		if !found {
			panic(newError("TypeError", "open() got an unexpected keyword argument '"+k.Name+"'"))
		}
	}
	if !set[0] {
		panic(newError("TypeError", "open() missing required argument 'file' (pos 1)"))
	}
	o.file = vals[0]
	if set[1] {
		m, ok := Unbox(vals[1]).(string)
		if !ok {
			panic(newError("TypeError", "open() argument 'mode' must be str, not "+TypeName(Unbox(vals[1]))))
		}
		o.mode = m
	}
	if set[2] {
		o.buffering = int(DynInt(vals[2]))
	}
	if set[3] && !IsNone(vals[3]) {
		if _, ok := Unbox(vals[3]).(string); !ok {
			panic(newError("TypeError", "open() argument 'encoding' must be str or None, not "+TypeName(Unbox(vals[3]))))
		}
		o.encoding = vals[3]
	}
	if set[4] && !IsNone(vals[4]) {
		if _, ok := Unbox(vals[4]).(string); !ok {
			panic(newError("TypeError", "open() argument 'errors' must be str or None, not "+TypeName(Unbox(vals[4]))))
		}
		o.errors = vals[4]
	}
	if set[5] && !IsNone(vals[5]) {
		if _, ok := Unbox(vals[5]).(string); !ok {
			panic(newError("TypeError", "open() argument 'newline' must be str or None, not "+TypeName(Unbox(vals[5]))))
		}
		o.newline = vals[5]
	}
	if set[6] {
		o.closefd = Truth(vals[6])
	}
	return o
}

// OpenFile is open(...).
func OpenFile(args []any, kw []KW) *File {
	o := parseOpenArgs(args, kw)
	var path string
	fd := -1
	fileArg := Unbox(o.file)
	switch v := fileArg.(type) {
	case int64:
		fd = int(v)
		if fd < 0 {
			panic(valueError("negative file descriptor"))
		}
	case bool:
		fd = 0
		if v {
			fd = 1
		}
	case string:
		path = v
	case Bytes:
		path = string(v)
	default:
		if m, ok := lookupMethod(fileArg, "__fspath__"); ok {
			r := Unbox(m())
			switch p := r.(type) {
			case string:
				path = p
			case Bytes:
				path = string(p)
			default:
				panic(newError("TypeError", "expected "+TypeName(fileArg)+".__fspath__() to return str or bytes, not "+TypeName(r)))
			}
			o.file = path
		} else {
			panic(newError("TypeError", "expected str, bytes or os.PathLike object, not "+TypeName(fileArg)))
		}
	}
	if fd < 0 && strings.IndexByte(path, 0) >= 0 {
		panic(valueError("embedded null byte"))
	}
	// the mode
	seen := map[rune]bool{}
	for _, c := range o.mode {
		if !strings.ContainsRune("rwxabt+U", c) || seen[c] {
			panic(valueError("invalid mode: '" + o.mode + "'"))
		}
		seen[c] = true
	}
	if seen['U'] {
		panic(valueError("invalid mode: '" + o.mode + "'"))
	}
	if seen['t'] && seen['b'] {
		panic(valueError("can't have text and binary mode at once"))
	}
	n := 0
	for _, c := range "rwxa" {
		if seen[c] {
			n++
		}
	}
	if n != 1 {
		panic(valueError("must have exactly one of create/read/write/append mode"))
	}
	binary := seen['b']
	plus := seen['+']
	if binary && o.encoding != nil {
		panic(valueError("binary mode doesn't take an encoding argument"))
	}
	if binary && o.errors != nil {
		panic(valueError("binary mode doesn't take an errors argument"))
	}
	if binary && o.newline != nil {
		panic(valueError("binary mode doesn't take a newline argument"))
	}
	if o.buffering == 0 && !binary {
		panic(valueError("can't have unbuffered text I/O"))
	}
	if !o.closefd && fd < 0 {
		panic(valueError("Cannot use closefd=False with file name"))
	}
	var flags int
	rd, wr, app := false, false, false
	switch {
	case seen['r']:
		rd = true
		flags = syscall.O_RDONLY
	case seen['w']:
		wr = true
		flags = syscall.O_WRONLY | syscall.O_CREAT | syscall.O_TRUNC
	case seen['x']:
		wr = true
		flags = syscall.O_WRONLY | syscall.O_CREAT | syscall.O_EXCL
	case seen['a']:
		wr, app = true, true
		flags = syscall.O_WRONLY | syscall.O_CREAT | syscall.O_APPEND
	}
	if plus {
		rd, wr = true, true
		flags = flags&^(syscall.O_RDONLY|syscall.O_WRONLY) | syscall.O_RDWR
	}
	if fd < 0 {
		var err error
		fd, err = syscall.Open(path, flags|syscall.O_CLOEXEC, 0o666)
		for err == syscall.EINTR {
			fd, err = syscall.Open(path, flags|syscall.O_CLOEXEC, 0o666)
		}
		if err != nil {
			panic(sysError(err, path))
		}
		var st syscall.Stat_t
		if syscall.Fstat(fd, &st) == nil && st.Mode&syscall.S_IFMT == syscall.S_IFDIR {
			syscall.Close(fd)
			panic(sysError(syscall.EISDIR, path))
		}
	} else {
		var st syscall.Stat_t
		if err := syscall.Fstat(fd, &st); err != nil {
			panic(sysError(err))
		}
	}
	// the mode text FileIO reports
	rawMode := "wb"
	switch {
	case seen['x']:
		rawMode = "xb"
		if plus {
			rawMode = "xb+"
		}
	case seen['a']:
		rawMode = "ab"
		if plus {
			rawMode = "ab+"
		}
	case rd && wr:
		rawMode = "rb+"
	case rd:
		rawMode = "rb"
	}
	var nameVal any = o.file
	if bs, ok := fileArg.(Bytes); ok {
		nameVal = bs
	}
	raw := &File{fd: fd, name: nameVal, mode: rawMode, rd: rd, wr: wr, app: app, closefd: o.closefd, rawIO: true, bufMax: defaultBufferSize}
	raw.tty = isatty(fd)
	if app {
		syscall.Seek(fd, 0, 2)
	}
	if o.buffering == 0 {
		return raw
	}
	size := defaultBufferSize
	if o.buffering > 1 {
		size = o.buffering
	}
	lineBuf := o.buffering == 1 || (o.buffering < 0 && raw.tty)
	raw.rawIO = false
	raw.bufMax = size
	openFiles[raw] = struct{}{}
	if binary {
		return raw
	}
	enc := "utf-8"
	if o.encoding != nil {
		enc = Unbox(o.encoding).(string)
	}
	codec := canonicalCodec(enc)
	errs := "strict"
	if o.errors != nil {
		errs = Unbox(o.errors).(string)
		checkHandler(errs)
	}
	nl := nlUniversal
	if o.newline != nil {
		switch Unbox(o.newline).(string) {
		case "":
			nl = nlEmpty
		case "\n":
			nl = nlLF
		case "\r":
			nl = nlCR
		case "\r\n":
			nl = nlCRLF
		default:
			panic(valueError("illegal newline value: " + Str(Unbox(o.newline))))
		}
	}
	tf := &File{inner: raw, fd: fd, name: nameVal, mode: o.mode, rd: rd, wr: wr, app: app, text: true,
		encoding: codec, codec: codec, given: enc, errs: errs, nlMode: nl, lineBuf: lineBuf}
	if o.encoding == nil {
		tf.given = "UTF-8"
	}
	openFiles[tf] = struct{}{}
	if codec == "utf-8-sig" {
		tf.bom = rd
		if wr && !app {
			raw.writeBytes([]byte("\xef\xbb\xbf"))
		}
	}
	return tf
}

// canonicalName is the name Python reports for an encoding: the codec name
// without the dashes' differences (utf-8, ascii, latin-1 is iso8859-1).
func canonicalName(given, codec string) string {
	if codec == "latin-1" {
		return "iso8859-1"
	}
	return codec
}

// lookupMethod finds a method of a user object by name.
func lookupMethod(obj any, name string) (func() any, bool) {
	var v any
	if Catch(func() { v = DynGetAttr(obj, name) }) != nil {
		return nil, false
	}
	return func() any { return DynCall(v) }, true
}

func init() {
	RegisterNative("builtins.open", func(args []any, kw []KW) any {
		return OpenFile(args, kw)
	})
	RegisterNative("builtins.input", func(args []any, kw []KW) any {
		if len(args) > 0 {
			writeStdout(Str(args[0]))
		}
		Stdout.flush()
		Stderr.flush()
		line := Stdin.readLineText(-1)
		if line == "" {
			panic(newError("EOFError", "EOF when reading a line"))
		}
		return strings.TrimSuffix(line, "\n")
	})
}

// osError is the exception for a failed system call, with the class and the
// message CPython uses: FileNotFoundError: [Errno 2] No such file or directory: 'x'.
func osError(err error, name string) *PyError {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return sysError(errno, name)
	}
	switch {
	case os.IsNotExist(err):
		return osErrorCode("FileNotFoundError", 2, "No such file or directory", name)
	case os.IsExist(err):
		return osErrorCode("FileExistsError", 17, "File exists", name)
	case os.IsPermission(err):
		return osErrorCode("PermissionError", 13, "Permission denied", name)
	}
	return newError("OSError", err.Error())
}

func osErrorCode(class string, code int, text, name string) *PyError {
	e := &PyError{Type: class, Args: []any{int64(code), text}}
	e.Attrs = map[string]any{"errno": int64(code), "strerror": text, "filename": nil, "filename2": nil}
	if name != "" {
		e.Attrs["filename"] = name
	}
	return e
}

// WithTempFile runs f on a file that nothing else holds, and closes the
// file afterwards, as CPython does when the object goes away.
func WithTempFile(file any, f func(any) any) any {
	if fl, ok := Unbox(file).(*File); ok {
		defer func() {
			if r := recover(); r != nil {
				func() { defer func() { recover() }(); fl.closeFile() }()
				panic(r)
			}
			fl.closeFile()
		}()
	}
	return f(file)
}
