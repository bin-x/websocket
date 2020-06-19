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
