package realip

import (
	"net"
	"net/http"

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
		next.ServeHTTP(w, req) // invalid remote ip, not sure what to do here. Maybe an error would be more appropriate if we only expect requests from known sources
	}
	reqIP := net.ParseIP(host)
	if reqIP == nil {
		next.ServeHTTP(w, req) //same as above
	}
	for _, from := range r.from {
		if from.Contains(reqIP) {
			validSource = true
		}
	}
	//TODO: reject if not from known source? Probably best lest to ipfilter
	if hVal := req.Header.Get(r.header); validSource && hVal != "" {
		//restore original host:port format
		//TODO: check header format validity
		req.RemoteAddr = hVal + ":" + port
	}
	return next.ServeHTTP(w, req)
}
