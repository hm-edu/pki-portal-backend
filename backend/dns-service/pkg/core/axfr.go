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

// AxfrProvider is an DNSProvider that executes updates using the RFC2136 AXFR protocol.
type AxfrProvider struct {
	log           *zap.Logger
	nameserver    string
	zoneName      string
	tsigKeyName   string
	tsigSecret    string
	tsigSecretAlg string
	axfr          bool
}

func (r AxfrProvider) incomeTransfer(m *dns.Msg, _ string) (env chan *dns.Envelope, err error) {
	t := new(dns.Transfer)
	t.TsigSecret = map[string]string{r.tsigKeyName: r.tsigSecret}

	return t.In(m, r.nameserver)
}

// List returns the current list of records.
func (r AxfrProvider) List() ([]dns.RR, error) {

	r.log.Sugar().Debugf("Fetching records for '%s'", r.zoneName)

	m := new(dns.Msg)
	m.SetAxfr(r.zoneName)
	m.SetTsig(r.tsigKeyName, r.tsigSecretAlg, clockSkew, time.Now().Unix())

	env, err := r.incomeTransfer(m, r.nameserver)
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
func (r AxfrProvider) Add(entries []dns.RR) error {
	m := new(dns.Msg)
	m.SetUpdate(r.zoneName)
	m.Insert(entries)
	return r.sendMessage(m)

}

// Delete removes the given records from the zone.
func (r AxfrProvider) Delete(entries []dns.RR) error {
	m := new(dns.Msg)
	m.SetUpdate(r.zoneName)
	m.Remove(entries)
	return r.sendMessage(m)
}

// Update updates the given records in the zone.
func (r AxfrProvider) Update(entries []UpdateSet) error {
	m := new(dns.Msg)
	m.SetUpdate(r.zoneName)

	for _, entry := range entries {
		m.Remove(entry.Old)
		m.Insert(entry.New)
	}

	return r.sendMessage(m)
}

func (r AxfrProvider) sendMessage(msg *dns.Msg) error {

	c := new(dns.Client)
	c.SingleInflight = true

	c.TsigSecret = map[string]string{r.tsigKeyName: r.tsigSecret}
	msg.SetTsig(r.tsigKeyName, r.tsigSecretAlg, clockSkew, time.Now().Unix())

	if msg.Len() > udpMaxMsgSize {
		c.Net = "tcp"
	}

	resp, _, err := c.Exchange(msg, r.nameserver)
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
