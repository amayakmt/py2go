package runtime

import "syscall"

func statFields(st *syscall.Stat_t) []any {
	at := int64(st.Atimespec.Sec)*1_000_000_000 + int64(st.Atimespec.Nsec)
	mt := int64(st.Mtimespec.Sec)*1_000_000_000 + int64(st.Mtimespec.Nsec)
	ct := int64(st.Ctimespec.Sec)*1_000_000_000 + int64(st.Ctimespec.Nsec)
	return []any{int64(st.Mode), int64(st.Ino), int64(st.Dev), int64(st.Nlink), int64(st.Uid), int64(st.Gid), st.Size,
		int64(st.Atimespec.Sec), int64(st.Mtimespec.Sec), int64(st.Ctimespec.Sec),
		float64(at) / 1e9, float64(mt) / 1e9, float64(ct) / 1e9, at, mt, ct,
		int64(st.Blksize), st.Blocks, int64(st.Rdev)}
}
