package helper

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	portalCommon "github.com/hm-edu/portal-common/helper"
	"github.com/miekg/dns"
	"gopkg.in/yaml.v3"
)

const (
	// maximum time DNS client can be off from server for an update to succeed
	clockSkew = 300

	// maximum size of a UDP transport message in DNS protocol
	udpMaxMsgSize = 512
)

type ProviderConfig struct {
	BaseDomain      string `yaml:"base_domain"`
	ReadNameserver  string `yaml:"read_nameserver"`
	WriteNameserver string `yaml:"write_nameserver"`
	TsigKeyName     string `yaml:"tsig_key_name"`
	TsigSecret      string `yaml:"tsig_secret"`
	TsigSecretAlg   string `yaml:"tsig_secret_alg"`
}

type DNSProvider struct {
	Configs []*ProviderConfig
}

func NewDNSProvider(config string) (*DNSProvider, error) {
	data, err := os.ReadFile(config)
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		return nil, err
	}

	providerConfigs := make([]*ProviderConfig, 0)
	err = yaml.Unmarshal(data, &providerConfigs)

	return &DNSProvider{}, nil
}

// List returns the current list of records.
func (r ProviderConfig) List() ([]dns.RR, error) {
	m := new(dns.Msg)
	m.SetAxfr(r.BaseDomain)
	t := new(dns.Transfer)
	t.TsigSecret = map[string]string{r.TsigKeyName: r.TsigSecret}
	m.SetTsig(r.TsigKeyName, r.TsigSecretAlg, clockSkew, time.Now().Unix())
	if r.TsigSecretAlg == dns.HmacMD5 {
		t.TsigProvider = portalCommon.Md5provider(r.TsigSecret)
	}
	env, err := t.In(m, r.ReadNameserver)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch records: %v", err)
	}

	records := make([]dns.RR, 0)
	for e := range env {
		if e.Error != nil {
			continue
		}
		records = append(records, e.RR...)
	}

	return records, nil
}

// Add adds the given records to the zone.
func (r ProviderConfig) Add(entries []dns.RR) error {
	m := new(dns.Msg)
	m.SetUpdate(r.BaseDomain)
	m.Insert(entries)
	return r.sendMessage(m)

}

// Delete removes the given records from the zone.
func (r ProviderConfig) Delete(entries []dns.RR) error {
	m := new(dns.Msg)
	m.SetUpdate(r.BaseDomain)
	m.Remove(entries)
	return r.sendMessage(m)
}

func (r ProviderConfig) sendMessage(msg *dns.Msg) error {

	c := new(dns.Client)

	c.TsigSecret = map[string]string{r.TsigKeyName: r.TsigSecret}
	msg.SetTsig(r.TsigKeyName, r.TsigSecretAlg, clockSkew, time.Now().Unix())
	if r.TsigSecretAlg == dns.HmacMD5 {
		c.TsigProvider = portalCommon.Md5provider(r.TsigSecret)
	}
	if msg.Len() > udpMaxMsgSize {
		c.Net = "tcp"
	}

	resp, _, err := c.Exchange(msg, r.WriteNameserver)
	if err != nil {
		if resp != nil && resp.Rcode != dns.RcodeSuccess {
			return err
		}
	}
	if resp != nil && resp.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("bad return code: %s", dns.RcodeToString[resp.Rcode])
	}

	return nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)
	// Get the DNS Provider with the best matching domain
	// Check if the currently selected config is more specific than the previous one
	matchingConfig, err := d.matchingProvider(info, domain)
	if err != nil {
		return err
	}

	// Use the matching DNS provider to create the TXT record
	rr := new(dns.TXT)
	rr.Hdr = dns.RR_Header{Name: info.FQDN, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60}
	rr.Txt = []string{info.Value}

	err = matchingConfig.Add([]dns.RR{rr})
	if err != nil {
		return fmt.Errorf("error adding TXT record: %v", err)
	}

	return nil
}

func (d *DNSProvider) matchingProvider(info dns01.ChallengeInfo, domain string) (*ProviderConfig, error) {
	var matchingConfig *ProviderConfig
	matchingConfig = nil
	for _, config := range d.Configs {
		if strings.HasSuffix(info.EffectiveFQDN, fmt.Sprintf(".%s", config.BaseDomain)) {
			if matchingConfig == nil {
				matchingConfig = config
				continue
			}

			if len(strings.Split(config.BaseDomain, ".")) > len(strings.Split(matchingConfig.BaseDomain, ".")) {
				matchingConfig = config
			}
		}
	}

	if matchingConfig == nil {
		return nil, fmt.Errorf("no matching DNS provider found for domain %s", domain)
	}
	return matchingConfig, nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	// clean up any state you created in Present, like removing the TXT record
	info := dns01.GetChallengeInfo(domain, keyAuth)
	// Get the DNS Provider with the best matching domain
	provider, err := d.matchingProvider(info, domain)
	if err != nil {
	}
	rr := new(dns.TXT)
	rr.Hdr = dns.RR_Header{Name: info.FQDN, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60}
	rr.Txt = []string{info.Value}

	err = provider.Delete([]dns.RR{rr})
	if err != nil {
		return fmt.Errorf("error deleting TXT record: %v", err)
	}

	return nil
}
