//go:build linux

package runtime

import "syscall"

// sysDup2 is dup2; arm64 Linux has only dup3.
func sysDup2(old, new int) error {
	if old == new {
		var st syscall.Stat_t
		return syscall.Fstat(old, &st)
	}
	return syscall.Dup3(old, new, 0)
}
