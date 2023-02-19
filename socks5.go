package final_socks

const (
	VersionSocks5 = uint8(5)
)

const (
	AuthNoAuth       = uint8(0)
	AuthUserPass     = uint8(2)
	AuthVersion      = uint8(1)
	AuthNoAcceptable = uint8(255)
	AuthSuccess      = uint8(0)
	AuthFailure      = uint8(1)
)

const (
	AddressIpv4 = uint8(1)
	AddressFqdn = uint8(3)
	AddressIpv6 = uint8(4)
)

const (
	CommandConnect   = uint8(1)
	CommandAssociate = uint8(3)
)

const (
	ReplySucceeded                     = uint8(0)
	ReplyGeneralServerFailure          = uint8(1)
	ReplyConnectionNotAllowedByRuleset = uint8(2)
	ReplyNetworkUnreachable            = uint8(3)
	ReplyHostUnreachable               = uint8(4)
	ReplyConnectionRefused             = uint8(5)
	ReplyConnectionTTLExpired          = uint8(6)
	ReplyCommandNotSupported           = uint8(7)
	ReplyAddressTypeNotSupported       = uint8(8)
)
