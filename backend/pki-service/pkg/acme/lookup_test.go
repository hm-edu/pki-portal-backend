package acme

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// startDNSServer serves handler over UDP and TCP on the same loopback port
// and returns the address of the server.
func startDNSServer(t *testing.T, handler dns.Handler) string {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := pc.LocalAddr().String()
	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	udp := &dns.Server{PacketConn: pc, Handler: handler}
	tcp := &dns.Server{Listener: l, Handler: handler}
	go func() { _ = udp.ActivateAndServe() }()
	go func() { _ = tcp.ActivateAndServe() }()
	t.Cleanup(func() {
		_ = udp.Shutdown()
		_ = tcp.Shutdown()
	})
	return addr
}

// startSilentServer binds a UDP port that never answers.
func startSilentServer(t *testing.T) string {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pc.Close() })
	return pc.LocalAddr().String()
}

func cnameRR(name, target string) dns.RR {
	return &dns.CNAME{
		Hdr:    dns.RR_Header{Name: name, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 60},
		Target: target,
	}
}

func testHandler(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	q := r.Question[0]
	switch strings.ToLower(q.Name) {
	case "_acme-challenge.cname.example.":
		m.Answer = append(m.Answer, cnameRR(q.Name, "target.example."))
	case "_acme-challenge.other.example.":
		// A CNAME for a different owner must not be reported.
		m.Answer = append(m.Answer, cnameRR("other.example.", "target.example."))
	case "_acme-challenge.nx.example.":
		m.Rcode = dns.RcodeNameError
	case "_acme-challenge.servfail.example.":
		m.Rcode = dns.RcodeServerFailure
	case "_acme-challenge.truncated.example.":
		if w.RemoteAddr().Network() == "udp" {
			m.Truncated = true
		} else {
			m.Answer = append(m.Answer, cnameRR(q.Name, "tcp-target.example."))
		}
	case "_acme-challenge.nodata.example.":
		// NOERROR without answer.
	}
	_ = w.WriteMsg(m)
}

func TestLookupCNAME(t *testing.T) {
	addr := startDNSServer(t, dns.HandlerFunc(testHandler))
	tests := []struct {
		name    string
		query   string
		want    string
		wantErr bool
	}{
		{name: "cname", query: "_acme-challenge.cname.example", want: "target.example."},
		{name: "cname with trailing dot", query: "_acme-challenge.cname.example.", want: "target.example."},
		{name: "cname case insensitive", query: "_ACME-Challenge.CNAME.example", want: "target.example."},
		{name: "cname of other owner", query: "_acme-challenge.other.example", want: ""},
		{name: "nxdomain", query: "_acme-challenge.nx.example", want: ""},
		{name: "nodata", query: "_acme-challenge.nodata.example", want: ""},
		{name: "truncated falls back to tcp", query: "_acme-challenge.truncated.example", want: "tcp-target.example."},
		{name: "servfail", query: "_acme-challenge.servfail.example", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lookupCNAME(context.Background(), tt.query, []string{addr}, 2*time.Second)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestLookupCNAME_FallbackNameserver(t *testing.T) {
	silent := startSilentServer(t)
	addr := startDNSServer(t, dns.HandlerFunc(testHandler))
	got, err := lookupCNAME(context.Background(), "_acme-challenge.cname.example", []string{silent, addr}, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if got != "target.example." {
		t.Errorf("expected CNAME from second nameserver, got %q", got)
	}
}

func TestLookupCNAME_AllNameserversFail(t *testing.T) {
	silent := startSilentServer(t)
	_, err := lookupCNAME(context.Background(), "_acme-challenge.cname.example", []string{silent, silent}, 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected error when no nameserver answers")
	}
}

func TestLookupCNAME_CancelledContext(t *testing.T) {
	addr := startDNSServer(t, dns.HandlerFunc(testHandler))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := lookupCNAME(ctx, "_acme-challenge.cname.example", []string{addr}, 2*time.Second)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
