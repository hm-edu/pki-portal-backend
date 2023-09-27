package worker

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/hm-edu/pki-service/ent"
	"github.com/hm-edu/pki-service/ent/certificate"
	"github.com/hm-edu/pki-service/ent/domain"
	"go.uber.org/zap"
)

// Notifier holds all required information for mail notifications related to certificate expiry
type Notifier struct {
	MailFrom  string
	MailHost  string
	MailPort  int
	MailTo    string
	MailToBcc string
	Db        *ent.Client
}

// Notify triggers an email for each
func (w *Notifier) Notify(logger *zap.Logger) error {

	domains, err := w.Db.Domain.Query().All(context.Background())
	if err != nil {
		return err
	}
	for _, d := range domains {
		certificate, err := w.Db.Certificate.Query().WithDomains().Where(certificate.HasDomainsWith(domain.FqdnEQ(d.Fqdn)),
			certificate.StatusNEQ(certificate.StatusRevoked),
			certificate.StatusNEQ(certificate.StatusInvalid)).Order(ent.Desc(certificate.FieldNotAfter)).Limit(1).All(context.Background())
		if err != nil {
			return err
		}
		if len(certificate) == 0 {
			continue
		}
		if certificate[0].NotAfter.Before(time.Now().AddDate(0, 0, 30)) && certificate[0].NotAfter.After(time.Now()) {
			days := time.Until(certificate[0].NotAfter).Hours() / 24
			if int(days)%7 == 0 {
				// TODO: send notification
				logger.Info(fmt.Sprintf("Certificate for %s expires in %d days, sending notification.", d.Fqdn, int(days)))
				certDomains := certificate[0].Edges.Domains[0].Fqdn
				for _, x := range certificate[0].Edges.Domains[1:] {
					certDomains = fmt.Sprintf("%s, %s", certDomains, x.Fqdn)
				}
				to := []string{strings.Split(*certificate[0].IssuedBy, " ")[0]}
				if w.MailTo != "" {
					to = []string{w.MailTo}
				}
				if w.MailToBcc != "" {
					to = append(to, w.MailToBcc)
				}
				err = smtp.SendMail(fmt.Sprintf("%s:%d", w.MailHost, w.MailPort), nil, w.MailFrom, to, []byte(fmt.Sprintf(`
				From: PKI <%s>
				To: %s
				Subject: Zertifikatsablauf %s

				Sehr geehrte(r) Nutzer(in) des PKI-Portals,

				Das letzte Zertifikat für die Domain %s wird am %s ablaufen.
				Bitte erneuern Sie dieses Zertifikat zeitnah.
				Das betreffende Zertifikat ist für folgende (weitere) Domains ausgestellt:

				%s.

				Sollten Sie Fragen haben, wenden Sie sich bitte an den Support.

				Mit freundlichen Grüßen,
				Ihre Zentrale IT
				`, w.MailFrom, strings.Split(*certificate[0].IssuedBy, " ")[0], d.Fqdn, d.Fqdn, certificate[0].NotAfter.Format("02.01.2006"), certDomains)))
				if err != nil {
					logger.Error("Error sending mail", zap.Error(err))
				}
			}
		}
	}
	return nil
}
