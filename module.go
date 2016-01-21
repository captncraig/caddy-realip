package realip

import (
	"net"
	"net/http"
	"strings"

	"github.com/mholt/caddy/middleware"
)

func (m *module) ServeHTTP(w http.ResponseWriter, req *http.Request) (int, error) {
	for _, r := range m.rules {
		if middleware.Path(req.URL.Path).Matches(r.path) {
			return r.handle(w, req, m.next)
		}
	}
	return m.next.ServeHTTP(w, req)
}

func (r *rule) handle(w http.ResponseWriter, req *http.Request, next middleware.Handler) (int, error) {
	validSource := false
	host, port, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return next.ServeHTTP(w, req) // invalid remote ip. Change nothing and let next deal with it.
	}
	reqIP := net.ParseIP(host)
	if reqIP == nil {
		return next.ServeHTTP(w, req) //same as above.
	}
	for _, from := range r.from {
		if from.Contains(reqIP) {
			validSource = true
		}
	}
	if hVal := req.Header.Get(r.header); validSource && hVal != "" {
		//restore original host:port format
		leftMost := strings.Split(hVal, ",")[0]
		if net.ParseIP(leftMost) != nil {
			req.RemoteAddr = leftMost + ":" + port
		}
	}
	return next.ServeHTTP(w, req)
}
