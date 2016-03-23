package realip

import (
	"net"

	"github.com/mholt/caddy/caddy/setup"
	"github.com/mholt/caddy/middleware"
)

type module struct {
	next   middleware.Handler
	From   []*net.IPNet
	Header string
	Strict bool
}

func Setup(c *setup.Controller) (middleware.Middleware, error) {
	var m *module
	for c.Next() {
		if m != nil {
			return nil, c.Err("cannot specify realip more than once")
		}
		m = &module{
			Header: "X-Forwarded-For",
		}
		if err := parse(m, c); err != nil {
			return nil, err
		}
	}
	return func(next middleware.Handler) middleware.Handler {
		m.next = next
		return m
	}, nil
}

func parse(m *module, c *setup.Controller) (err error) {
	args := c.RemainingArgs()
	if len(args) == 1 && args[0] == "cloudflare" {
		addCloudflareIps(m)
		if c.NextBlock() {
			return c.Err("No realip subblocks allowed if using preset.")
		}
	}
	if len(args) != 0 {
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
// Helpers below here could potentially be methods on *setup.Contoller for convenience

// Assert only one arg and return it
func StringArg(c *setup.Controller) (string, error) {
	args := c.RemainingArgs()
	if len(args) != 1 {
		return "", c.ArgErr()
	}
	return args[0], nil
}

// Assert only one arg is a valid cidr notation
func CidrArg(c *setup.Controller) (*net.IPNet, error) {
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

func BoolArg(c *setup.Controller) (bool, error) {
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

func NoArgs(c *setup.Controller) error {
	if len(c.RemainingArgs()) != 0 {
		return c.ArgErr()
	}
	return nil
}
