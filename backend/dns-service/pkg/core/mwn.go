package core

import (
	"fmt"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

const (

	// maximum time DNS client can be off from server for an update to succeed
	clockSkew = 300

	// maximum size of a UDP transport message in DNS protocol
	udpMaxMsgSize = 512
)

// MwnProviderConfig is the configuration for the MWN provider.
type MwnProviderConfig struct {
	ReadNameserver  string `mapstructure:"read_nameserver"`
	WriteNameserver string `mapstructure:"write_nameserver"`
	TsigKeyName     string `mapstructure:"tsig_key_name"`
	TsigSecret      string `mapstructure:"tsig_secret"`
	TsigSecretAlg   string `mapstructure:"tsig_secret_alg"`
}

// MwnProvider is an DNSProvider that executes updates using the RFC2136 AXFR protocol.
// It is adapted to the setup in the MWN and thus reads from a different nameserver than writing.
// Also reading is done without an TSIG-Key. Instead this is done using an IP-ACL.
type MwnProvider struct {
	log *zap.Logger
	cfg *MwnProviderConfig
}

// NewMwnProvider returns a new MWN provider.
func NewMwnProvider(logger *zap.Logger, cfg *MwnProviderConfig) *MwnProvider {
	tsigAlgs := map[string]string{
		"hmac-md5":    dns.HmacMD5,
		"hmac-sha1":   dns.HmacSHA1,
		"hmac-sha256": dns.HmacSHA256,
		"hmac-sha512": dns.HmacSHA512,
	}
	cfg.TsigSecretAlg = tsigAlgs[cfg.TsigSecretAlg]
	return &MwnProvider{
		log: logger,
		cfg: cfg,
	}
}

// List returns the current list of records.
func (r MwnProvider) List(zoneName string) ([]dns.RR, error) {

	r.log.Sugar().Debugf("Fetching records for '%s'", zoneName)

	m := new(dns.Msg)
	m.SetAxfr(zoneName)
	t := new(dns.Transfer)
	env, err := t.In(m, r.cfg.ReadNameserver)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch records: %v", err)
	}

	records := make([]dns.RR, 0)
	for e := range env {
		if e.Error != nil {
			if e.Error == dns.ErrSoa {
				r.log.Sugar().Errorf("unexpected response received from the server")
			} else {
				r.log.Sugar().Errorf("%v", e.Error)
			}
			continue
		}
		records = append(records, e.RR...)
	}

	return records, nil
}

// Add adds the given records to the zone.
func (r MwnProvider) Add(zoneName string, entries []dns.RR) error {
	m := new(dns.Msg)
	m.SetUpdate(zoneName)
	m.Insert(entries)
	return r.sendMessage(m)

}

// Delete removes the given records from the zone.
func (r MwnProvider) Delete(zoneName string, entries []dns.RR) error {
	m := new(dns.Msg)
	m.SetUpdate(zoneName)
	m.Remove(entries)
	return r.sendMessage(m)
}

// Update updates the given records in the zone.
func (r MwnProvider) Update(zoneName string, entries []UpdateSet) error {
	m := new(dns.Msg)
	m.SetUpdate(zoneName)

	for _, entry := range entries {
		m.Remove(entry.Old)
		m.Insert(entry.New)
	}

	return r.sendMessage(m)
}

func (r MwnProvider) sendMessage(msg *dns.Msg) error {

	c := new(dns.Client)
	c.SingleInflight = true

	c.TsigSecret = map[string]string{r.cfg.TsigKeyName: r.cfg.TsigSecret}
	msg.SetTsig(r.cfg.TsigKeyName, r.cfg.TsigSecretAlg, clockSkew, time.Now().Unix())
	if r.cfg.TsigSecretAlg == dns.HmacMD5 {
		c.TsigProvider = md5provider(r.cfg.TsigSecret)
	}
	if msg.Len() > udpMaxMsgSize {
		c.Net = "tcp"
	}

	resp, _, err := c.Exchange(msg, r.cfg.WriteNameserver)
	if err != nil {
		if resp != nil && resp.Rcode != dns.RcodeSuccess {
			log.Infof("error in Exchange: %s", err)
			return err
		}
		log.Warnf("warn in Exchange: %s", err)
	}
	if resp != nil && resp.Rcode != dns.RcodeSuccess {
		log.Infof("Bad Exchange response: %s", resp)
		return fmt.Errorf("bad return code: %s", dns.RcodeToString[resp.Rcode])
	}

	return nil
}
