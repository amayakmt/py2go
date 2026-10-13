package runtime

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
	"syscall"
)

// File is the object open() returns, and sys.stdout, sys.stderr and
// sys.stdin: a text or binary file over an *os.File.
type File struct {
	f      *os.File
	r      *bufio.Reader
	name   string
	mode   string
	binary bool
	closed bool
	std    bool // standard streams stay open when closed
	// dirty is set once something was written, so that read after write
	// seeks to the end of what was written.
	wrote bool
}

// Stdout, Stderr and Stdin are the standard streams as file objects.
var (
	Stdout = &File{f: os.Stdout, name: "<stdout>", mode: "w", std: true}
	Stderr = &File{f: os.Stderr, name: "<stderr>", mode: "w", std: true}
	Stdin  = &File{f: os.Stdin, name: "<stdin>", mode: "r", std: true}
)

func (f *File) PyClassName() string {
	if f.binary {
		return "BufferedReader"
	}
	return "TextIOWrapper"
}

func (f *File) PyRepr() string {
	return "<_io.TextIOWrapper name=" + quote(f.name) + " mode=" + quote(f.mode) + " encoding='UTF-8'>"
}

var fileMethodNames = names("read readline readlines write writelines close flush seek tell truncate fileno isatty readable writable seekable __enter__ __exit__ __iter__ __next__")

func (f *File) PyDynTable() map[string]bool { return fileMethodNames }

func (f *File) PyGetAttr(name string) (any, bool) {
	switch name {
	case "name":
		return f.name, true
	case "mode":
		return f.mode, true
	case "closed":
		return f.closed, true
	case "encoding":
		if f.binary {
			return nil, false
		}
		return "UTF-8", true
	case "errors":
		return "strict", true
	case "newlines":
		return nil, true
	}
	if fileMethodNames[name] {
		return func(args ...any) any {
			r, _ := f.DynMethod(name, args, nil)
			return r
		}, true
	}
	return nil, false
}

func (f *File) check() { f.checkOp("") }

// checkOp raises the ValueError for an operation on a closed file; a binary
// file names the operation.
func (f *File) checkOp(op string) {
	if !f.closed {
		return
	}
	if f.binary && op == "read" {
		panic(valueError("read of closed file"))
	}
	if f.binary && op == "write" {
		panic(valueError("write to closed file"))
	}
	panic(valueError("I/O operation on closed file."))
}

func (f *File) reader() *bufio.Reader {
	if f.r == nil {
		f.r = bufio.NewReader(f.f)
	}
	return f.r
}

func (f *File) wrap(s string) any {
	if f.binary {
		return Bytes(s)
	}
	return universalNewlines(s)
}

// universalNewlines is the text of a file opened in text mode: \r\n and \r
// are read as \n.
func universalNewlines(s string) string {
	if !strings.Contains(s, "\r") {
		return s
	}
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}

func (f *File) readLine(limit int) string {
	r := f.reader()
	var sb strings.Builder
	for limit < 0 || sb.Len() < limit {
		b, err := r.ReadByte()
		if err != nil {
			break
		}
		sb.WriteByte(b)
		if b == '\n' {
			break
		}
		if b == '\r' && !f.binary {
			if next, err := r.Peek(1); err == nil && next[0] == '\n' {
				r.ReadByte()
			}
			return sb.String()[:sb.Len()-1] + "\n"
		}
	}
	return sb.String()
}

func (f *File) IterAny() Cursor[any] {
	f.check()
	return &fileCursor{f: f}
}

type fileCursor struct {
	f   *File
	cur any
}

func (c *fileCursor) Next() bool {
	c.f.check()
	line := c.f.readLine(-1)
	if line == "" {
		return false
	}
	if c.f.binary {
		c.cur = Bytes(line)
	} else {
		c.cur = line
	}
	return true
}

func (c *fileCursor) Value() any { return c.cur }

func (f *File) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "__enter__":
		f.check()
		return f, true
	case "__exit__":
		f.closeFile()
		return false, true
	case "close":
		f.closeFile()
		return nil, true
	case "flush":
		f.check()
		return nil, true
	case "fileno":
		return int64(f.f.Fd()), true
	case "isatty":
		return false, true
	case "readable":
		return strings.ContainsAny(f.mode, "r+"), true
	case "writable":
		return strings.ContainsAny(f.mode, "wax+"), true
	case "seekable":
		return !f.std, true
	case "read":
		f.checkOp("read")
		if !strings.ContainsAny(f.mode, "r+") {
			panic(newError("OSError", "not readable"))
		}
		n := int64(-1)
		if len(args) > 0 && !IsNone(args[0]) {
			n = DynInt(args[0])
		}
		r := f.reader()
		if n < 0 {
			data, _ := io.ReadAll(r)
			return f.wrap(string(data)), true
		}
		buf := make([]byte, n)
		got, _ := io.ReadFull(r, buf)
		return f.wrap(string(buf[:got])), true
	case "readline":
		f.check()
		limit := -1
		if len(args) > 0 && !IsNone(args[0]) {
			limit = int(DynInt(args[0]))
		}
		return f.wrap2(f.readLine(limit)), true
	case "readlines":
		f.check()
		out := NewList[any]()
		for {
			line := f.readLine(-1)
			if line == "" {
				break
			}
			out.Append(f.wrap2(line))
		}
		return out, true
	case "__next__":
		f.check()
		line := f.readLine(-1)
		if line == "" {
			panic(newError("StopIteration", ""))
		}
		return f.wrap2(line), true
	case "write":
		f.checkOp("write")
		if !strings.ContainsAny(f.mode, "wax+") {
			panic(newError("OSError", "not writable"))
		}
		countArgs("write", args, 1, 1)
		var data string
		switch x := Unbox(args[0]).(type) {
		case string:
			if f.binary {
				panic(newError("TypeError", "a bytes-like object is required, not 'str'"))
			}
			data = x
		default:
			bs, ok := bytesOf(x)
			if !ok || !f.binary {
				if f.binary {
					panic(newError("TypeError", "a bytes-like object is required, not '"+TypeName(x)+"'"))
				}
				panic(newError("TypeError", "write() argument must be str, not "+TypeName(x)))
			}
			data = bs
		}
		f.syncForWrite()
		f.f.WriteString(data)
		if f.binary {
			return int64(len(data)), true
		}
		return StrLen(data), true
	case "writelines":
		f.check()
		for c := DynIter(args[0]).Iter(); c.Next(); {
			f.DynMethod("write", []any{c.Value()}, nil)
		}
		return nil, true
	case "seek":
		f.check()
		off := DynInt(argAt(args, 0))
		whence := int(argInt(args, 1, 0))
		if f.r != nil && whence == 1 {
			off -= int64(f.r.Buffered())
		}
		pos, err := f.f.Seek(off, whence)
		if err != nil {
			panic(newError("OSError", err.Error()))
		}
		f.r = nil
		return pos, true
	case "tell":
		f.check()
		pos, _ := f.f.Seek(0, 1)
		if f.r != nil {
			pos -= int64(f.r.Buffered())
		}
		return pos, true
	case "truncate":
		f.check()
		size := int64(0)
		if len(args) > 0 && !IsNone(args[0]) {
			size = DynInt(args[0])
		} else {
			size, _ = f.f.Seek(0, 1)
		}
		f.f.Truncate(size)
		return size, true
	}
	return nil, false
}

func (f *File) wrap2(line string) any {
	if f.binary {
		return Bytes(line)
	}
	return line
}

// syncForWrite makes a write after buffered reads land at the logical
// position.
func (f *File) syncForWrite() {
	if f.r != nil {
		pos, _ := f.f.Seek(0, 1)
		pos -= int64(f.r.Buffered())
		f.f.Seek(pos, 0)
		f.r = nil
	}
	f.wrote = true
}

func (f *File) closeFile() {
	if f.closed {
		return
	}
	f.closed = true
	if !f.std {
		f.f.Close()
	}
}

func (f *File) PyTruth() bool { return true }

func init() {
	RegisterNative("builtins.open", func(args []any, kw []KW) any {
		file := argAt(args, 0)
		if v, ok := kwValue(kw, "file"); ok {
			file = v
		}
		mode := "r"
		if len(args) > 1 {
			mode = argStr(args, 1, "r")
		}
		if v, ok := kwValue(kw, "mode"); ok {
			mode, _ = Unbox(v).(string)
		}
		name, ok := Unbox(file).(string)
		if !ok {
			if bs, isBytes := bytesOf(file); isBytes {
				name = bs
			} else {
				panic(newError("TypeError", "expected str, bytes or os.PathLike object, not "+TypeName(Unbox(file))))
			}
		}
		binary := strings.Contains(mode, "b")
		base := strings.NewReplacer("b", "", "t", "").Replace(mode)
		var flag int
		switch base {
		case "r":
			flag = os.O_RDONLY
		case "w":
			flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		case "a":
			flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
		case "x":
			flag = os.O_WRONLY | os.O_CREATE | os.O_EXCL
		case "r+":
			flag = os.O_RDWR
		case "w+":
			flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
		case "a+":
			flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
		case "x+":
			flag = os.O_RDWR | os.O_CREATE | os.O_EXCL
		default:
			panic(valueError("invalid mode: '" + mode + "'"))
		}
		fh, err := os.OpenFile(name, flag, 0o666)
		if err != nil {
			panic(osError(err, name))
		}
		if st, serr := fh.Stat(); serr == nil && st.IsDir() {
			fh.Close()
			panic(osErrorCode("IsADirectoryError", 21, "Is a directory", name))
		}
		return &File{f: fh, name: name, mode: mode, binary: binary}
	})
	RegisterNative("builtins.input", func(args []any, kw []KW) any {
		if len(args) > 0 {
			Stdout.f.WriteString(Str(args[0]))
		}
		line := Stdin.readLine(-1)
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
		switch errno {
		case syscall.ENOENT:
			return osErrorCode("FileNotFoundError", 2, "No such file or directory", name)
		case syscall.EEXIST:
			return osErrorCode("FileExistsError", 17, "File exists", name)
		case syscall.EACCES:
			return osErrorCode("PermissionError", 13, "Permission denied", name)
		case syscall.EPERM:
			return osErrorCode("PermissionError", 1, "Operation not permitted", name)
		case syscall.EISDIR:
			return osErrorCode("IsADirectoryError", 21, "Is a directory", name)
		case syscall.ENOTDIR:
			return osErrorCode("NotADirectoryError", 20, "Not a directory", name)
		case syscall.ENOTEMPTY:
			return osErrorCode("OSError", 39, "Directory not empty", name)
		}
		text := errno.Error()
		return osErrorCode("OSError", int(errno), strings.ToUpper(text[:1])+text[1:], name)
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
	e := newError(class, "[Errno "+itoa(code)+"] "+text+": "+quote(name))
	e.Attrs = map[string]any{"errno": int64(code), "strerror": text, "filename": name}
	return e
}
