//go:build !windows

package tui

import "golang.org/x/sys/unix"

// keyAvail dit si le descripteur a des octets à lire, sans bloquer.
// Sondage immédiat (timeout nul) : pas d'attente.
func keyAvail(fd uintptr) bool {
	var rfds unix.FdSet
	rfds.Set(int(fd))
	tv := unix.Timeval{}
	n, err := unix.Select(int(fd)+1, &rfds, nil, nil, &tv)
	if err != nil || n <= 0 {
		return false
	}
	return rfds.IsSet(int(fd))
}
