package websocket

import (
	"net"
	"strconv"
	"strings"
)

// check Check host address format
// xxx.xxx.xxx.xxx:xxx will return true
// else return false.
func HostAddrCheck(addr string) bool {
	items := strings.Split(addr, ":")
	if items == nil || len(items) != 2 {
		return false
	}

	a := net.ParseIP(items[0])
	if a == nil {
		return false
	}

	i, err := strconv.Atoi(items[1])
	if err != nil {
		return false
	}
	if i < 0 || i > 65535 {
		return false
	}

	return true
}

//check port range
func checkPort(i int) bool {
	return i > 0 && i <= 65535
}

func EqualWithoutIndex(a1, b1 []string) bool {
	if len(a1) != len(b1) {
		return false
	}
	a := make([]string, len(a1))
	b := make([]string, len(b1))
	copy(a, a1)
	copy(b, b1)
re:
	for _, va := range a {
		for kb, vb := range b {
			if va == vb {
				b = append(b[0:kb], b[kb+1:]...)
				continue re
			}
		}
		return false
	}
	return true
}
