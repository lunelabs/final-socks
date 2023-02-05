package final_socks

import (
	"bufio"
	"github.com/pkg/errors"
	"io"
	"net"
)

type Request struct {
	DestAddr   *AddrSpec
	LocalAddr  net.TCPAddr
	RemoteAddr net.TCPAddr
	Version    uint8
	Command    uint8
	BufConn    *bufio.Reader
}

func ReadSocksVersion(bufConn *bufio.Reader) (uint8, error) {
	version := []byte{0}

	if _, err := bufConn.Read(version); err != nil {
		return 0, errors.Wrap(err, "failed to get version byte")
	}

	return version[0], nil
}

func ReadAuthenticateMethods(bufConn *bufio.Reader) ([]byte, error) {
	header := []byte{0}

	if _, err := bufConn.Read(header); err != nil {
		return nil, errors.Wrap(err, "failed to get auth methods")
	}

	numMethods := int(header[0])
	methods := make([]byte, numMethods)
	_, err := io.ReadAtLeast(bufConn, methods, numMethods)

	return methods, errors.Wrap(err, "failed to get auth methods")
}

func ReadRequest(bufConn *bufio.Reader) (*Request, error) {
	header := []byte{0, 0, 0}

	if _, err := io.ReadAtLeast(bufConn, header, 3); err != nil {
		return nil, errors.Wrap(err, "failed to read header")
	}

	addressType := []byte{0}

	if _, err := bufConn.Read(addressType); err != nil {
		return nil, errors.Wrap(err, "failed read address type")
	}

	if addressType[0] != AddressIpv4 && addressType[0] != AddressIpv6 && addressType[0] != AddressFqdn {
		return nil, errors.New("cant read address type")
	}

	dest := &AddrSpec{}

	switch addressType[0] {
	case AddressIpv4:
		addr := make([]byte, 4)

		if _, err := io.ReadAtLeast(bufConn, addr, len(addr)); err != nil {
			return nil, err
		}

		dest.IP = addr

	case AddressIpv6:
		addr := make([]byte, 16)

		if _, err := io.ReadAtLeast(bufConn, addr, len(addr)); err != nil {
			return nil, err
		}

		dest.IP = addr

	case AddressFqdn:
		if _, err := bufConn.Read(addressType); err != nil {
			return nil, err
		}

		addrLen := int(addressType[0])
		fqdn := make([]byte, addrLen)

		if _, err := io.ReadAtLeast(bufConn, fqdn, addrLen); err != nil {
			return nil, err
		}

		dest.FQDN = string(fqdn)
	}

	port := []byte{0, 0}

	if _, err := io.ReadAtLeast(bufConn, port, 2); err != nil {
		return nil, err
	}

	dest.Port = (int(port[0]) << 8) | int(port[1])

	request := &Request{
		Version:  header[0],
		Command:  header[1],
		DestAddr: dest,
		BufConn:  bufConn,
	}

	return request, nil
}
