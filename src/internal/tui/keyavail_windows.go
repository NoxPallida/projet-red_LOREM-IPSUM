//go:build windows

package tui

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modkernel32                       = windows.NewLazyDLL("kernel32.dll")
	procGetNumberOfConsoleInputEvents = modkernel32.NewProc("GetNumberOfConsoleInputEvents")
)

// keyAvail : événements console en attente (touches... mais aussi
// relâchements qui ne produisent aucun octet : PollKey ne bloquera
// pas dessus, il rend juste have=false après un court délai).
func keyAvail(fd uintptr) bool {
	var n uint32
	r1, _, _ := procGetNumberOfConsoleInputEvents.Call(fd, uintptr(unsafe.Pointer(&n)))
	if r1 == 0 {
		// Pas une console (pipe redirigé) : laisse Read trancher (EOF -> quitte).
		return true
	}
	return n > 0
}
