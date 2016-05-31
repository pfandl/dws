package dws

var (
	// General errors
	ErrCannotConvertData = NewError("could not convert data to byte array")

	// Configuration network errors
	ErrConfigNetworkNoneAvailable     = NewError("no networks configured")
	ErrConfigNetworkNameEmpty         = NewError("no network name given")
	ErrConfigNetworkNameUsed          = NewError("network name is already used")
	ErrConfigNetworkNameNotFound      = NewError("network not found")
	ErrConfigNetworkIpV4GatewayEmpty  = NewError("no ipv4 gateway for network given")
	ErrConfigNetworkIpV4SubnetEmpty   = NewError("no ipv4 subnet for network given")
	ErrConfigNetworkIpV4GatewaySyntax = NewError("ipv4 gateway for network is not a valid address")
	ErrConfigNetworkIpV4SubnetSyntax  = NewError("ipv4 subnet for network is not a valid address")
	ErrConfigNetworkTypeInvalid       = NewError("network type invalid")

	// Configuration host errors
	ErrConfigHostNoneAvailable        = NewError("no hosts configured")
	ErrConfigHostNameEmpty            = NewError("no host name given")
	ErrConfigHostNameUsed             = NewError("host name is already used")
	ErrConfigHostNameNotFound         = NewError("host not found")
	ErrConfigHostIpV4AddressEmpty     = NewError("no ipv4 address for host given")
	ErrConfigHostIpV4AddressSyntax    = NewError("ipv4 address for host is not valid")
	ErrConfigHostIpV4AddressUsed      = NewError("host ipv4 address is already used")
	ErrConfigHostIpV4MacAddressEmpty  = NewError("no ipv4 mac address for host given")
	ErrConfigHostIpV4MacAddressSyntax = NewError("ipv4 mac address for host is not valid")
	ErrConfigHostIpV4MacAddressUsed   = NewError("host ipv4 mac address is already used")
	ErrConfigHostUtsNameEmpty         = NewError("no utsname for host given")
	ErrConfigHostUtsNameSyntax        = NewError("uts name for host is not valid")
	ErrConfigHostUtsNameUsed          = NewError("host uts name is already used")

	// LXC errors
	ErrLxcNoTemplatesInstalled = NewError("lxc-templates not installed")
	ErrLxcNoTemplatesFound     = NewError("no templates found")
)

// Error represents a basic error that implies the error interface.
type Error struct {
	Message string
}

// NewError creates a new error with the given msg argument.
func NewError(msg string) error {
	return &Error{
		Message: msg,
	}
}

func (e *Error) Error() string {
	return e.Message
}
