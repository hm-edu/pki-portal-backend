package acme

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

const (
	// maximum time the DNS client can be off from the server for an update to succeed
	clockSkew = 300

	// maximum size of a UDP transport message in the DNS protocol
	udpMaxMsgSize = 512

	// TTL of the published challenge records
	challengeTTL = 60
)

// DNSProvider publishes DNS-01 challenges using RFC2136 dynamic updates.
// The zone, nameserver and TSIG key are selected per challenge domain based
// on the DNS validation configuration.
type DNSProvider struct {
	cfg    *DNSConfig
	logger *zap.Logger
}

// NewDNSProvider returns a new RFC2136 based DNS-01 challenge provider.
func NewDNSProvider(cfg *DNSConfig, logger *zap.Logger) *DNSProvider {
	return &DNSProvider{cfg: cfg, logger: logger}
}

// Present publishes the TXT record for the given challenge.
func (p *DNSProvider) Present(ctx context.Context, domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)
	zone, rr, err := p.challengeRecord(info)
	if err != nil {
		return err
	}
	p.logger.Info("Publishing DNS-01 challenge",
		zap.String("fqdn", info.EffectiveFQDN),
		zap.String("zone", zone.Zone),
		zap.String("nameserver", zone.Nameserver))
	m := new(dns.Msg)
	m.SetUpdate(dns.Fqdn(zone.Zone))
	m.Insert([]dns.RR{rr})
	return p.sendMessage(ctx, zone, m)
}

// CleanUp removes the TXT record of the given challenge again.
func (p *DNSProvider) CleanUp(ctx context.Context, domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)
	zone, rr, err := p.challengeRecord(info)
	if err != nil {
		return err
	}
	p.logger.Info("Removing DNS-01 challenge",
		zap.String("fqdn", info.EffectiveFQDN),
		zap.String("zone", zone.Zone))
	m := new(dns.Msg)
	m.SetUpdate(dns.Fqdn(zone.Zone))
	m.Remove([]dns.RR{rr})
	return p.sendMessage(ctx, zone, m)
}

// Timeout tells lego how long to wait for the challenge records to propagate
// to all authoritative nameservers (e.g. secondaries).
func (p *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 10 * time.Minute, 10 * time.Second
}

func (p *DNSProvider) challengeRecord(info dns01.ChallengeInfo) (*Zone, dns.RR, error) {
	zone := p.cfg.ZoneFor(info.EffectiveFQDN)
	if zone == nil {
		return nil, nil, fmt.Errorf("no DNS zone configured for %s", info.EffectiveFQDN)
	}
	rr := &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   dns.Fqdn(strings.ToLower(info.EffectiveFQDN)),
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    challengeTTL,
		},
		Txt: []string{info.Value},
	}
	return zone, rr, nil
}

func (p *DNSProvider) sendMessage(ctx context.Context, zone *Zone, msg *dns.Msg) error {
	c := new(dns.Client)
	c.Timeout = 10 * time.Second
	c.TsigSecret = map[string]string{zone.TsigKeyName: zone.TsigSecret}
	msg.SetTsig(zone.TsigKeyName, zone.TsigAlgorithm, clockSkew, time.Now().Unix())
	if zone.TsigAlgorithm == dns.HmacMD5 { //nolint:staticcheck // Still required for legacy zones.
		c.TsigProvider = md5provider(zone.TsigSecret)
	}
	if msg.Len() > udpMaxMsgSize {
		c.Net = "tcp"
	}

	resp, _, err := c.ExchangeContext(ctx, msg, zone.Nameserver)
	if err != nil {
		return fmt.Errorf("DNS update for zone %s failed: %w", zone.Zone, err)
	}
	if resp != nil && resp.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("DNS update for zone %s failed: %s", zone.Zone, dns.RcodeToString[resp.Rcode])
	}
	return nil
}
