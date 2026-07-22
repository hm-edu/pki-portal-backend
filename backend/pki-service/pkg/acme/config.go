// Package acme provides certificate issuance using an ACME CA (e.g.
// Let's Encrypt). Domain validation is performed using DNS-01 challenges
// that are published via RFC2136 dynamic updates signed with per-zone
// TSIG keys.
package acme

import (
	"fmt"
	"os"
	"strings"

	"github.com/miekg/dns"
	"gopkg.in/yaml.v3"
)

// tsigAlgorithms maps the algorithm names used in the configuration file to
// the constants of the dns library.
var tsigAlgorithms = map[string]string{
	"hmac-md5":    dns.HmacMD5, //nolint:staticcheck // Still required for legacy zones.
	"hmac-sha1":   dns.HmacSHA1,
	"hmac-sha256": dns.HmacSHA256,
	"hmac-sha512": dns.HmacSHA512,
}

// Zone describes a single DNS zone that can be used for DNS-01 validation.
type Zone struct {
	// Zone is the name of the DNS zone. All (sub-)domains of this zone are
	// validated using this entry; the most specific zone wins.
	Zone string `yaml:"zone"`
	// Nameserver is the server receiving the dynamic updates (host[:port],
	// port defaults to 53).
	Nameserver string `yaml:"nameserver"`
	// TsigKeyName is the name of the TSIG key used to sign the updates.
	TsigKeyName string `yaml:"tsig_key_name"`
	// TsigAlgorithm is one of hmac-md5, hmac-sha1, hmac-sha256 or hmac-sha512.
	TsigAlgorithm string `yaml:"tsig_algorithm"`
	// TsigSecret is the base64 encoded shared secret.
	TsigSecret string `yaml:"tsig_secret"`
}

// DNSConfig is the content of the DNS validation configuration file. It maps
// domains/subdomains to the TSIG keys that are used to publish the DNS-01
// challenges.
type DNSConfig struct {
	Zones []Zone `yaml:"zones"`
}

// LoadDNSConfig reads and validates the DNS validation configuration file.
func LoadDNSConfig(path string) (*DNSConfig, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is provided by the operator
	if err != nil {
		return nil, fmt.Errorf("reading DNS config %s: %w", path, err)
	}
	var cfg DNSConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing DNS config %s: %w", path, err)
	}
	if len(cfg.Zones) == 0 {
		return nil, fmt.Errorf("DNS config %s contains no zones", path)
	}
	for i := range cfg.Zones {
		zone := &cfg.Zones[i]
		if zone.Zone == "" {
			return nil, fmt.Errorf("DNS config %s: zone %d has no zone name", path, i)
		}
		zone.Zone = normalizeDomain(zone.Zone)
		if zone.Nameserver == "" {
			return nil, fmt.Errorf("DNS config %s: zone %s has no nameserver", path, zone.Zone)
		}
		if !strings.Contains(zone.Nameserver, ":") {
			zone.Nameserver += ":53"
		}
		if zone.TsigKeyName == "" || zone.TsigSecret == "" {
			return nil, fmt.Errorf("DNS config %s: zone %s has no TSIG key", path, zone.Zone)
		}
		zone.TsigKeyName = dns.Fqdn(strings.ToLower(zone.TsigKeyName))
		alg, ok := tsigAlgorithms[strings.ToLower(zone.TsigAlgorithm)]
		if !ok {
			return nil, fmt.Errorf("DNS config %s: zone %s has unsupported TSIG algorithm %q", path, zone.Zone, zone.TsigAlgorithm)
		}
		zone.TsigAlgorithm = alg
	}
	return &cfg, nil
}

// normalizeDomain lower-cases a domain and strips wildcard prefixes and
// trailing dots so it can be compared label-wise.
func normalizeDomain(domain string) string {
	domain = strings.ToLower(strings.TrimSuffix(domain, "."))
	domain = strings.TrimPrefix(domain, "*.")
	return domain
}

// ZoneFor returns the most specific configured zone for the given domain or
// nil if the domain is not covered by the configuration.
func (c *DNSConfig) ZoneFor(domain string) *Zone {
	domain = normalizeDomain(domain)
	var best *Zone
	for i := range c.Zones {
		zone := &c.Zones[i]
		if domain != zone.Zone && !strings.HasSuffix(domain, "."+zone.Zone) {
			continue
		}
		if best == nil || len(zone.Zone) > len(best.Zone) {
			best = zone
		}
	}
	return best
}

// Covers reports whether all given domains can be validated with the
// configured zones.
func (c *DNSConfig) Covers(domains []string) bool {
	for _, domain := range domains {
		if c.ZoneFor(domain) == nil {
			return false
		}
	}
	return true
}
