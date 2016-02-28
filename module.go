package realip

import (
	"net"
	"net/http"
	"strings"
)

func (m *module) ServeHTTP(w http.ResponseWriter, req *http.Request) (int, error) {
	validSource := false
	host, port, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return m.next.ServeHTTP(w, req) // invalid remote ip. Change nothing and let next deal with it.
	}
	reqIP := net.ParseIP(host)
	if reqIP == nil {
		return m.next.ServeHTTP(w, req) //same as above.
	}
	for _, from := range m.From {
		if from.Contains(reqIP) {
			validSource = true
		}
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
