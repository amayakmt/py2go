package runtime

import (
	"os"
	"syscall"
)

func unameFields() [5]any {
	get := func(name string) string { s, _ := syscall.Sysctl(name); return s }
	host, _ := os.Hostname()
	return [5]any{get("kern.ostype"), host, get("kern.osrelease"), get("kern.version"), get("hw.machine")}
}
