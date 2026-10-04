package acme

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// recursiveNameservers are the public resolvers used for every lookup that
// has to reflect the public DNS view (propagation checks, authoritative
// nameserver discovery, CNAME checks). The system resolver may expose an
// internal view in split-DNS environments that is not visible to the ACME CA.
var recursiveNameservers = []string{"1.1.1.1:53", "8.8.8.8:53"}

// lookupTimeout bounds a single DNS exchange with one nameserver.
const lookupTimeout = 5 * time.Second

// LookupCNAME returns the target of the CNAME record owned by name as seen
// by the public recursive resolvers. It returns an empty string if no CNAME
// exists at that name; NXDOMAIN and NODATA responses are not errors. CNAME
// chains are not followed. An error is returned only if none of the
// resolvers delivered a usable answer.
func LookupCNAME(ctx context.Context, name string) (string, error) {
	return lookupCNAME(ctx, name, recursiveNameservers, lookupTimeout)
}

func lookupCNAME(ctx context.Context, name string, nameservers []string, timeout time.Duration) (string, error) {
	fqdn := dns.Fqdn(strings.ToLower(name))
	m := new(dns.Msg)
	m.SetQuestion(fqdn, dns.TypeCNAME)
	m.RecursionDesired = true

	var errs []error
	for _, ns := range nameservers {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			break
		}
		resp, err := exchange(ctx, m, ns, timeout)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", ns, err))
			continue
		}
		switch resp.Rcode {
		case dns.RcodeSuccess:
			return extractCNAME(resp, fqdn), nil
		case dns.RcodeNameError:
			return "", nil
		default:
			errs = append(errs, fmt.Errorf("%s: %s", ns, dns.RcodeToString[resp.Rcode]))
		}
	}
	return "", fmt.Errorf("CNAME lookup for %s failed: %w", fqdn, errors.Join(errs...))
}

// exchange sends m to the nameserver over UDP and repeats the query over TCP
// if the response was truncated.
func exchange(ctx context.Context, m *dns.Msg, nameserver string, timeout time.Duration) (*dns.Msg, error) {
	c := &dns.Client{Timeout: timeout}
	resp, _, err := c.ExchangeContext(ctx, m, nameserver)
	if err != nil {
		return nil, err
	}
	if resp.Truncated {
		c.Net = "tcp"
		resp, _, err = c.ExchangeContext(ctx, m, nameserver)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

// extractCNAME returns the target of the CNAME record in the answer section
// that is owned by fqdn, or an empty string if there is none.
func extractCNAME(msg *dns.Msg, fqdn string) string {
	for _, rr := range msg.Answer {
		if cn, ok := rr.(*dns.CNAME); ok && strings.EqualFold(cn.Hdr.Name, fqdn) {
			return cn.Target
		}
	}
	return ""
}
