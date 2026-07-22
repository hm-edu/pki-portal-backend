package acme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/miekg/dns"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dns.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadDNSConfig(t *testing.T) {
	path := writeConfig(t, `
zones:
  - zone: hm.edu
    nameserver: ns1.hm.edu
    tsig_key_name: acme-hm
    tsig_algorithm: hmac-sha256
    tsig_secret: c2VjcmV0
  - zone: cs.hm.edu.
    nameserver: ns1.cs.hm.edu:5353
    tsig_key_name: acme-cs.
    tsig_algorithm: HMAC-SHA512
    tsig_secret: c2VjcmV0
`)
	cfg, err := LoadDNSConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Zones) != 2 {
		t.Fatalf("expected 2 zones, got %d", len(cfg.Zones))
	}
	if cfg.Zones[0].Nameserver != "ns1.hm.edu:53" {
		t.Errorf("expected default port to be added, got %s", cfg.Zones[0].Nameserver)
	}
	if cfg.Zones[1].Nameserver != "ns1.cs.hm.edu:5353" {
		t.Errorf("expected port to be kept, got %s", cfg.Zones[1].Nameserver)
	}
	if cfg.Zones[0].TsigKeyName != "acme-hm." {
		t.Errorf("expected TSIG key name to be a FQDN, got %s", cfg.Zones[0].TsigKeyName)
	}
	if cfg.Zones[0].TsigAlgorithm != dns.HmacSHA256 {
		t.Errorf("unexpected algorithm %s", cfg.Zones[0].TsigAlgorithm)
	}
	if cfg.Zones[1].Zone != "cs.hm.edu" {
		t.Errorf("expected trailing dot to be stripped, got %s", cfg.Zones[1].Zone)
	}
	if cfg.Zones[1].TsigAlgorithm != dns.HmacSHA512 {
		t.Errorf("expected case insensitive algorithm, got %s", cfg.Zones[1].TsigAlgorithm)
	}
}

func TestLoadDNSConfigInvalid(t *testing.T) {
	cases := map[string]string{
		"no zones":    `zones: []`,
		"no name":     "zones:\n  - nameserver: ns1.hm.edu\n    tsig_key_name: a\n    tsig_algorithm: hmac-sha256\n    tsig_secret: b",
		"no ns":       "zones:\n  - zone: hm.edu\n    tsig_key_name: a\n    tsig_algorithm: hmac-sha256\n    tsig_secret: b",
		"no key":      "zones:\n  - zone: hm.edu\n    nameserver: ns1.hm.edu\n    tsig_algorithm: hmac-sha256",
		"bad alg":     "zones:\n  - zone: hm.edu\n    nameserver: ns1.hm.edu\n    tsig_key_name: a\n    tsig_algorithm: hmac-sha384\n    tsig_secret: b",
		"invalid yml": `{{`,
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadDNSConfig(writeConfig(t, content)); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestZoneFor(t *testing.T) {
	cfg := &DNSConfig{Zones: []Zone{
		{Zone: "hm.edu"},
		{Zone: "cs.hm.edu"},
	}}
	cases := []struct {
		domain string
		want   string
	}{
		{"hm.edu", "hm.edu"},
		{"www.hm.edu", "hm.edu"},
		{"WWW.HM.EDU", "hm.edu"},
		{"www.hm.edu.", "hm.edu"},
		{"cs.hm.edu", "cs.hm.edu"},
		{"host.cs.hm.edu", "cs.hm.edu"},
		{"_acme-challenge.host.cs.hm.edu.", "cs.hm.edu"},
		{"*.cs.hm.edu", "cs.hm.edu"},
		{"foohm.edu", ""},
		{"example.com", ""},
	}
	for _, c := range cases {
		zone := cfg.ZoneFor(c.domain)
		got := ""
		if zone != nil {
			got = zone.Zone
		}
		if got != c.want {
			t.Errorf("ZoneFor(%q) = %q, want %q", c.domain, got, c.want)
		}
	}
}

func TestCovers(t *testing.T) {
	cfg := &DNSConfig{Zones: []Zone{{Zone: "hm.edu"}}}
	if !cfg.Covers([]string{"hm.edu", "www.hm.edu", "*.hm.edu"}) {
		t.Error("expected domains to be covered")
	}
	if cfg.Covers([]string{"www.hm.edu", "example.com"}) {
		t.Error("expected domains not to be covered")
	}
}
