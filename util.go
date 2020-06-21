package websocket

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net"
	"strconv"
	"strings"
)

// HostAddrCheck() Check host address format
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

func Ip2long(ip net.IP) uint32 {
	a := uint32(ip[12])
	b := uint32(ip[13])
	c := uint32(ip[14])
	d := uint32(ip[15])
	return uint32(a<<24 | b<<16 | c<<8 | d)
}

func Long2ip(ip uint32) net.IP {
	a := byte((ip >> 24) & 0xFF)
	b := byte((ip >> 16) & 0xFF)
	c := byte((ip >> 8) & 0xFF)
	d := byte(ip & 0xFF)
	return net.IPv4(a, b, c, d)
}

func AddressToClientId(ip string, port uint16, id uint32) string {
	ipInt := Ip2long(net.ParseIP(ip))
	ipBin := make([]byte, 4)
	portBin := make([]byte, 2)
	idBin := make([]byte, 4)
	binary.BigEndian.PutUint32(ipBin, ipInt)
	binary.BigEndian.PutUint16(portBin, uint16(port))
	binary.BigEndian.PutUint32(idBin, uint32(id))
	b := append(append(ipBin, portBin...), idBin...)
	s := hex.EncodeToString(b)
	return s
}

func ClientIdToAddress(clientId string) (ip string, port uint16, id uint32, err error) {
	if len(clientId) != 20 {
		err = errors.New("client Id not validated")
		return "", 0, 0, err
	}
	b, err := hex.DecodeString(clientId)
	if err != nil {
		return "", 0, 0, nil
	}
	ipBin := b[:4]
	portBin := b[4:6]
	idBin := b[6:]

	ipInt := binary.BigEndian.Uint32(ipBin)
	port = binary.BigEndian.Uint16(portBin)
	id = binary.BigEndian.Uint32(idBin)
	ip = Long2ip(ipInt).String()
	return
}
