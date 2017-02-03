package realip

import (
	"net"
	"strconv"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("realip", caddy.Plugin{
		ServerType: "http",
		Action:     Setup,
	})
}

type module struct {
	next   httpserver.Handler
	From   []*net.IPNet
	Header string

	// MaxHops configures the maxiumum number of hops or IPs to be found in a forward header.
	// It's purpose is to prevent abuse and/or DOS attacks from long forward-chains, since each one
	// must be parsed and checked against a list of subnets.
	// The default is 5, -1 to disable. If set to 0, any request with a forward header will be rejected
	MaxHops int
	Strict  bool
}

func Setup(c *caddy.Controller) error {
	var m *module
	for c.Next() {
		if m != nil {
			return c.Err("cannot specify realip more than once")
		}
		m = &module{
			Header:  "X-Forwarded-For",
			MaxHops: 5,
		}
		if err := parse(m, c); err != nil {
			return err
		}
	}
	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		m.next = next
		return m
	})
	return nil
}

func parse(m *module, c *caddy.Controller) (err error) {
	args := c.RemainingArgs()
	if len(args) == 1 && args[0] == "cloudflare" {
		addCloudflareIps(m)
		if c.NextBlock() {
			return c.Err("No realip subblocks allowed if using preset.")
		}
	} else if len(args) != 0 {
		return c.ArgErr()
	}
	for c.NextBlock() {
		var err error
		switch c.Val() {
		case "header":
			m.Header, err = StringArg(c)
		case "from":
			var cidr *net.IPNet
			cidr, err = CidrArg(c)
			m.From = append(m.From, cidr)
		case "strict":
			m.Strict, err = BoolArg(c)
		case "maxhops":
			m.MaxHops, err = IntArg(c)
		default:
			return c.Errf("Unknown realip arg: %s", c.Val())
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func addCloudflareIps(m *module) {
	// from https://www.cloudflare.com/ips/
	var cfPresets = []string{
		"103.21.244.0/22",
		"103.22.200.0/22",
		"103.31.4.0/22",
		"104.16.0.0/12",
		"108.162.192.0/18",
		"131.0.72.0/22",
		"141.101.64.0/18",
		"162.158.0.0/15",
		"172.64.0.0/13",
		"173.245.48.0/20",
		"188.114.96.0/20",
		"190.93.240.0/20",
		"197.234.240.0/22",
		"198.41.128.0/17",
		"199.27.128.0/21",
		"2400:cb00::/32",
		"2405:8100::/32",
		"2405:b500::/32",
		"2606:4700::/32",
		"2803:f800::/32",
		"2a06:98c0::/29",
		"2c0f:f248::/32",
	}
	for _, c := range cfPresets {
		_, cidr, err := net.ParseCIDR(c)
		if err != nil {
			panic(err)
		}
		m.From = append(m.From, cidr)
	}
}

///////
// Helpers below here could potentially be methods on *caddy.Contoller for convenience

// IntArg check's there is only one arg, parses, and returns it
func IntArg(c *caddy.Controller) (int, error) {
	args := c.RemainingArgs()
	if len(args) != 1 {
		return 0, c.ArgErr()
	}
	return strconv.Atoi(args[0])
}

// Assert only one arg and return it
func StringArg(c *caddy.Controller) (string, error) {
	args := c.RemainingArgs()
	if len(args) != 1 {
		return "", c.ArgErr()
	}
	return args[0], nil
}

// Assert only one arg is a valid cidr notation
func CidrArg(c *caddy.Controller) (*net.IPNet, error) {
	a, err := StringArg(c)
	if err != nil {
		return nil, err
	}
	_, cidr, err := net.ParseCIDR(a)
	if err != nil {
		return nil, err
	}
	return cidr, nil
}

func BoolArg(c *caddy.Controller) (bool, error) {
	args := c.RemainingArgs()
	if len(args) > 1 {
		return false, c.ArgErr()
	}
	if len(args) == 0 {
		return true, nil
	}
	switch args[0] {
	case "false":
		return false, nil
	case "true":
		return true, nil
	default:
		return false, c.Errf("Unexpected bool value: %s", args[0])
	}
}

func NoArgs(c *caddy.Controller) error {
	if len(c.RemainingArgs()) != 0 {
		return c.ArgErr()
	}
	return nil
}
