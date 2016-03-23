package realip

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

func (m *module) ServeHTTP(w http.ResponseWriter, req *http.Request) (int, error) {
	validSource := false
	host, port, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		if m.Strict {
			return 403, fmt.Errorf("Error reading remote addr: %s", req.RemoteAddr)
		}
		return m.next.ServeHTTP(w, req) // Change nothing and let next deal with it.
	}
	reqIP := net.ParseIP(host)
	if reqIP == nil {
		if m.Strict {
			return 403, fmt.Errorf("Error parsing remote addr: %s", host)
		}
		return m.next.ServeHTTP(w, req)
	}
	for _, from := range m.From {
		if from.Contains(reqIP) {
			validSource = true
			break
		}
	}
	if !validSource && m.Strict {
		return 403, fmt.Errorf("Unrecognized proxy ip address: %s", reqIP)
	}
	if hVal := req.Header.Get(m.Header); validSource && hVal != "" {
		//restore original host:port format
		leftMost := strings.Split(hVal, ",")[0]
		if net.ParseIP(leftMost) != nil {
			req.RemoteAddr = leftMost + ":" + port
		}
	}
	return m.next.ServeHTTP(w, req)
}
