package precheck

import (
	"fmt"
	"net"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

func CheckDNS(domain, fqdn, value string, _ dns01.PreCheckFunc) (bool, error) {
	for _, ns := range []string{"1.1.1.1", "8.8.8.8"} {
		data, err := LookupTxt(fqdn, ns)
		if err != nil {
			return false, err
		}
		found := false
		for _, txt := range data {
			if txt == value {
				found = true
				continue
			}
		}
		if !found {
			return false, fmt.Errorf("TXT record not found")
		}
	}
	zap.L().Sugar().Infof("TXT record found for %s", fqdn)
	return true, nil
}

var timeouts = []time.Duration{(time.Second * 1), (time.Second * 1), (time.Second * 2), (time.Second * 4), (time.Second * 2)}

func ResolveWithTimeout(name, resolver string, qtype, qclass uint16) (*dns.Msg, error) {
	client := new(dns.Client)
	msg := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:               dns.Id(),
			RecursionDesired: true,
		},
		Question: []dns.Question{{Name: dns.Fqdn(name), Qtype: qtype, Qclass: qclass}},
	}
	msg.AuthenticatedData = true
	msg.SetEdns0(4096, true)

	for i := 0; i < len(timeouts); i++ {

		client.Timeout = timeouts[i]
		resp, _, err := client.Exchange(msg, fmt.Sprintf("%s:53", resolver))
		if err == nil && resp.Truncated {
			tcpConn, _ := dns.Dial("tcp", fmt.Sprintf("%s:53", resolver))
			resp, _, err = client.ExchangeWithConn(msg, tcpConn)
		}
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				zap.L().Sugar().Warnf("Timeout querying %s records '%s' after %v", dns.TypeToString[qtype], name, timeouts[i])
				continue
			}
			return nil, err
		}

		return resp, nil

	}
	return nil, &net.DNSError{
		Name:      name,
		Err:       "Final timeout.",
		IsTimeout: true,
	}
}

func LookupTxt(name, resolver string) ([]string, error) {
	zap.L().Sugar().Infof("Using custom resolver %s for lookup of %s", resolver, name)
	resp, err := ResolveWithTimeout(name, resolver, dns.TypeTXT, dns.ClassINET)
	if err != nil {
		zap.L().Sugar().Warnf("Failed to lookup %s: %v", name, err)
		return nil, err
	}
	data := []string{}

	// Check if TXT records are present
	for _, answer := range resp.Answer {
		if txt, ok := answer.(*dns.TXT); ok {
			zap.L().Sugar().Infof("Resolved TXT records for %s: %s", name, txt.Txt)
			data = append(data, txt.Txt...)
		}
	}
	return data, nil
}
