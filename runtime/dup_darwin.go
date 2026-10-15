//go:build darwin

package runtime

import "syscall"

func sysDup2(old, new int) error { return syscall.Dup2(old, new) }
