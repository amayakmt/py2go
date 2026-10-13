package link

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/amayakmt/py2go/stdlib"
)

var (
	stdlibOnce sync.Once
	stdlibPath string
)

// stdlibRoot is a directory with the modules py2go ships, written out once
// from the embedded tree so that the Python that parses them can read them.
// It is named after the content, so a changed module gets a fresh copy.
func stdlibRoot() string {
	stdlibOnce.Do(func() {
		h := sha256.New()
		type file struct {
			path string
			data []byte
		}
		var files []file
		fs.WalkDir(stdlib.Files, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			data, err := stdlib.Files.ReadFile(p)
			if err != nil {
				return nil
			}
			h.Write([]byte(p))
			h.Write(data)
			files = append(files, file{p, data})
			return nil
		})
		dir := filepath.Join(os.TempDir(), "py2go-stdlib-"+hex.EncodeToString(h.Sum(nil))[:16])
		if _, err := os.Stat(filepath.Join(dir, ".complete")); err != nil {
			tmp, err := os.MkdirTemp(os.TempDir(), "py2go-stdlib-tmp-")
			if err != nil {
				return
			}
			for _, f := range files {
				p := filepath.Join(tmp, f.path)
				os.MkdirAll(filepath.Dir(p), 0o755)
				os.WriteFile(p, f.data, 0o644)
			}
			os.WriteFile(filepath.Join(tmp, ".complete"), nil, 0o644)
			if err := os.Rename(tmp, dir); err != nil {
				os.RemoveAll(tmp) // another process won
			}
		}
		stdlibPath = dir
	})
	return stdlibPath
}
