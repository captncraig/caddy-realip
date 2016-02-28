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
	defer catchError(c, &err)
	if args := c.RemainingArgs(); len(args) != 0 {
		return c.ArgErr()
	}
	for c.NextBlock() {
		switch c.Val() {
		case "header":
			m.Header = StringArg(c)
		case "from":
			m.From = append(m.From, CidrArg(c))
		default:
			return c.Errf("Unknown realip arg: %s", c.Val())
		}
	}
	return nil
}

///////
// Helpers below here could potentially be methods on *setup.Contoller for convenience

// helper to recover panics and translate into error.
// lets helers simply panic(err) to make control flow simpler.
// I generally dislike this approach, but in this case I think a tighter parse function
// may be worth it.
func catchError(c *setup.Controller, err *error) {
	if e := recover(); e != nil {
		if er, ok := e.(error); ok {
			*err = er
		} else {
			*err = c.Err("Unknown error occurred")
		}
	}
}

// Assert only one arg and return it
func StringArg(c *setup.Controller) string {
	args := c.RemainingArgs()
	if len(args) != 1 {
		panic(c.ArgErr())
	}
	return args[1]
}

// Assert only one arg is a valid cidr notation
func CidrArg(c *setup.Controller) *net.IPNet {
	_, cidr, err := net.ParseCIDR(StringArg(c))
	if err != nil {
		panic(err)
	}
	return cidr
}

//other possible helpers:
// IntArg
// BoolArg
// TwoStrings, ThreeStrings, FourStrings (for conevenience like `a,b := c.TwoStrings()`)
// IPAddrArg
