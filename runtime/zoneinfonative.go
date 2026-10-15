package runtime

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// The natives of zoneinfo, over the zone database Go reads (the system's, or
// the one embedded in the program).

var (
	zoneMu    sync.Mutex
	zoneCache = map[string]*time.Location{}
)

func loadZone(key string) *time.Location {
	zoneMu.Lock()
	defer zoneMu.Unlock()
	if l, ok := zoneCache[key]; ok {
		return l
	}
	if key == "" || key == "Local" || strings.HasPrefix(key, "/") || strings.Contains(key, "..") {
		return nil
	}
	l, err := time.LoadLocation(key)
	if err != nil {
		zoneCache[key] = nil
		return nil
	}
	zoneCache[key] = l
	return l
}

// zoneAt is the offset, daylight saving amount and abbreviation at an instant.
func zoneAt(l *time.Location, t time.Time) (offset int, dst int, abbr string) {
	t = t.In(l)
	abbr, offset = t.Zone()
	if !t.IsDST() {
		return offset, 0, abbr
	}
	// the standard offset is the one of the nearest period without DST
	std := offset
	for _, d := range []int{30, 60, 90, 120, 180, 270} {
		for _, sign := range []int{1, -1} {
			c := t.AddDate(0, 0, sign*d)
			if !c.IsDST() {
				_, o := c.Zone()
				std = o
				return offset, offset - std, abbr
			}
		}
	}
	return offset, 3600, abbr
}

func init() {
	RegisterNative("zoneinfo.exists", func(args []any, kw []KW) any {
		return loadZone(Str(args[0])) != nil
	})
	RegisterNative("zoneinfo.at_utc", func(args []any, kw []KW) any {
		l := loadZone(Str(args[0]))
		if l == nil {
			panic(newError("KeyError", "zone"))
		}
		o, d, a := zoneAt(l, time.Unix(DynInt(args[1]), 0))
		return NewVarTuple[any](int64(o), int64(d), a)
	})
	// zoneinfo.at_local(key, wall_seconds, fold): the offset that applies to
	// a wall-clock time (seconds since the epoch as if it were UTC).
	RegisterNative("zoneinfo.at_local", func(args []any, kw []KW) any {
		l := loadZone(Str(args[0]))
		if l == nil {
			panic(newError("KeyError", "zone"))
		}
		w := DynInt(args[1])
		fold := DynInt(args[2])
		before, _, _ := zoneAt(l, time.Unix(w-86400, 0))
		after, _, _ := zoneAt(l, time.Unix(w+86400, 0))
		valid := func(off int) bool {
			o, _, _ := zoneAt(l, time.Unix(w-int64(off), 0))
			return o == off
		}
		off := before
		switch {
		case before == after:
			off = before
		case valid(before) && valid(after):
			// ambiguous: the first occurrence has fold 0
			if fold == 1 {
				off = after
			}
		case valid(before):
			off = before
		case valid(after):
			off = after
		default:
			// in a gap: fold 0 uses the offset before the transition
			if fold == 1 {
				off = after
			}
		}
		u := time.Unix(w-int64(off), 0)
		o, d, a := zoneAt(l, u)
		if o != off {
			// a gap: report the offset the fold chose, with the names of its side
			_, d, a = zoneAt(l, time.Unix(w-int64(off), 0).Add(time.Duration(off-o)*time.Second))
			o = off
		}
		return NewVarTuple[any](int64(o), int64(d), a)
	})
	RegisterNative("zoneinfo.available", func(args []any, kw []KW) any {
		out := NewList[any]()
		var names []string
		for _, root := range []string{"/usr/share/zoneinfo", "/usr/lib/zoneinfo", "/usr/share/lib/zoneinfo", "/etc/zoneinfo"} {
			if st, err := os.Stat(root); err != nil || !st.IsDir() {
				continue
			}
			filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				rel := strings.TrimPrefix(p, root+"/")
				if strings.HasPrefix(rel, "posix/") || strings.HasPrefix(rel, "right/") || strings.Contains(rel, ".") && !strings.Contains(rel, "/") && rel != "UTC" {
					return nil
				}
				if loadZone(rel) != nil {
					names = append(names, rel)
				}
				return nil
			})
			break
		}
		sort.Strings(names)
		for _, n := range names {
			out.Append(n)
		}
		return out
	})
}
