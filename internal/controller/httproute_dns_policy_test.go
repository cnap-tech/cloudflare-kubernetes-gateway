package controller

import (
	"os"
	"strings"
	"testing"
)

func TestHTTPRouteControllerDoesNotManagePerHostnameDNSRecords(t *testing.T) {
	source, err := os.ReadFile("httproute_controller.go")
	if err != nil {
		t.Fatal(err)
	}

	content := string(source)
	for _, forbidden := range []string{"api.DNS.Records", "FindZoneID"} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("HTTPRoute controller must not manage per-hostname DNS records; found %q", forbidden)
		}
	}
}
