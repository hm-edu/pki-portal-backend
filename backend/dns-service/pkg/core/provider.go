package core

import "github.com/miekg/dns"

// UpdateSet holds the data that shal be removed and added.
type UpdateSet struct {
	Old []dns.RR
	New []dns.RR
}

// DNSProvider provides the basic interface for an
type DNSProvider interface {
	List() ([]dns.RR, error)
	Add([]dns.RR) error
	Update([]UpdateSet) error
	Delete([]dns.RR) error
}
