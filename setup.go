package realip

import (
	"net"

	"github.com/mholt/caddy/caddy/setup"
	"github.com/mholt/caddy/middleware"
)

type module struct {
	next  middleware.Handler
	rules []*rule
}

type rule struct {
	path   string
	from   []*net.IPNet
	header string
}

func Setup(c *setup.Controller) (middleware.Middleware, error) {
	m, err := parse(c)
	if err != nil {
		return nil, err
	}
	return func(next middleware.Handler) middleware.Handler {
		m.next = next
		return m
	}, nil
}

func parse(c *setup.Controller) (*module, error) {
	m := &module{}
	for c.Next() {
		r := &rule{
			header: "X-Forwarded-For",
		}
		args := c.RemainingArgs()
		if len(args) != 0 {
			return nil, c.Errf("realip directive takes no arguments")
		}
		for c.NextBlock() {
			switch c.Val() {
			case "from":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return nil, c.ArgErr()
				}
				_, net, err := net.ParseCIDR(args[0])
				if err != nil {
					return nil, c.Errf("realip: %s", err)
				}
				r.from = append(r.from, net)
			case "header":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return nil, c.ArgErr()
				}
				r.header = args[0]
			default:
				return nil, c.Errf("unknown realip config item: %s", c.Val())
			}
		}

		m.rules = append(m.rules, r)
	}
	return m, nil
}
