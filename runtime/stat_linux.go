package runtime

import "syscall"

// statFields is the tuple stat() returns: mode, ino, dev, nlink, uid, gid,
// size, the three times as integer seconds, as floats and as nanoseconds,
// blksize, blocks and rdev.
func statFields(st *syscall.Stat_t) []any {
	at := int64(st.Atim.Sec)*1_000_000_000 + int64(st.Atim.Nsec)
	mt := int64(st.Mtim.Sec)*1_000_000_000 + int64(st.Mtim.Nsec)
	ct := int64(st.Ctim.Sec)*1_000_000_000 + int64(st.Ctim.Nsec)
	return []any{int64(st.Mode), int64(st.Ino), int64(st.Dev), int64(st.Nlink), int64(st.Uid), int64(st.Gid), st.Size,
		int64(st.Atim.Sec), int64(st.Mtim.Sec), int64(st.Ctim.Sec),
		float64(at) / 1e9, float64(mt) / 1e9, float64(ct) / 1e9, at, mt, ct,
		int64(st.Blksize), st.Blocks, int64(st.Rdev)}
}
