package core

import "github.com/miekg/dns"

// UpdateSet holds the data that shal be removed and added.
type UpdateSet struct {
	Old []dns.RR
	New []dns.RR
}

// DNSProvider provides the basic interface for an
type DNSProvider interface {
	List(string) ([]dns.RR, error)
	Add(string, []dns.RR) error
	Update(string, []UpdateSet) error
	Delete(string, []dns.RR) error
}
