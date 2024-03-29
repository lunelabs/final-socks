package final_socks

import (
	"fmt"
	"net"
	"strconv"
)

type AddrSpec struct {
	FQDN string
	IP   net.IP
	Port int
}

func (a *AddrSpec) String() string {
	if a.FQDN != "" {
		return fmt.Sprintf("%s (%s):%d", a.FQDN, a.IP, a.Port)
	}

	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

func (a *AddrSpec) Address() string {
	if 0 != len(a.IP) {
		return net.JoinHostPort(a.IP.String(), strconv.Itoa(a.Port))
	}

	return net.JoinHostPort(a.FQDN, strconv.Itoa(a.Port))
}

func (a *AddrSpec) Host() string {
	if 0 != len(a.IP) {
		return a.IP.String()
	}

	return a.FQDN
}
