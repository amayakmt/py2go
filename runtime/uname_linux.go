package runtime

import "syscall"

// unameFields is sysname, nodename, release, version and machine.
func unameFields() [5]any {
	var u syscall.Utsname
	if err := syscall.Uname(&u); err != nil {
		panic(sysError(err))
	}
	cs := func(a [65]int8) string {
		b := make([]byte, 0, 65)
		for _, c := range a {
			if c == 0 {
				break
			}
			b = append(b, byte(c))
		}
		return string(b)
	}
	return [5]any{cs(u.Sysname), cs(u.Nodename), cs(u.Release), cs(u.Version), cs(u.Machine)}
}
