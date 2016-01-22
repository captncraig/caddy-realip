package realip

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mholt/caddy/middleware"
)

func TestRealIP(t *testing.T) {
	for i, test := range []struct {
		actualIP   string
		headerVal  string
		path       string
		expectedIP string
	}{
		{"1.2.3.4:123", "", "foo", "1.2.3.4:123"},
		{"4.4.255.255:123", "", "xyz", "4.4.255.255:123"},
		{"4.5.0.0:123", "1.2.3.4", "xyz", "1.2.3.4:123"},
		{"4.5.2.3:123", "1.2.6.7,5.6.7.8,111.111.111.111", "xyz", "1.2.6.7:123"},
		{"4.5.5.5:123", "NOTANIP", "xyz", "4.5.5.5:123"},
		{"aaaaaa", "1.2.3.4", "xyz", "aaaaaa"},
		{"aaaaaa:123", "1.2.3.4", "xyz", "aaaaaa:123"},
	} {
		remoteAddr := ""
		_, ipnet, err := net.ParseCIDR("4.5.0.0/16") // "4.5.x.x"
		if err != nil {
			t.Fatal(err)
		}
		he := &module{
			next: middleware.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
				remoteAddr = r.RemoteAddr
				return 0, nil
			}),
			rules: []*rule{
				&rule{path: "/xyz", header: "X-Real-IP", from: []*net.IPNet{ipnet}},
			},
		}

		req, err := http.NewRequest("GET", "http://foo.tld/"+test.path, nil)
		if err != nil {
			t.Fatalf("Test %d: Could not create HTTP request: %v", i, err)
		}
		req.RemoteAddr = test.actualIP
		if test.headerVal != "" {
			req.Header.Set("X-Real-IP", test.headerVal)
		}

		rec := httptest.NewRecorder()
		he.ServeHTTP(rec, req)

		if remoteAddr != test.expectedIP {
			t.Errorf("Test %d: Expected '%s', but found '%s'", i, test.expectedIP, remoteAddr)
		}
	}
}
