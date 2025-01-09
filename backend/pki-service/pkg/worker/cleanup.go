package worker

import (
	"context"
	"time"

	"github.com/hm-edu/pki-service/ent"
	"github.com/hm-edu/pki-service/ent/certificate"
	"go.uber.org/zap"
)

// Cleanup checks for expired certificates and marks them as expired
func Cleanup(logger *zap.Logger, db *ent.Client) error {

	certs, err := db.Certificate.Query().Where(certificate.And(certificate.StatusEQ(certificate.StatusIssued), certificate.NotAfterLT(time.Now()))).All(context.Background())
	if err != nil {
		return err
	}
	for _, cert := range certs {
		// Mark certificate as expired
		if _, err := db.Certificate.UpdateOneID(cert.ID).SetStatus(certificate.StatusExpired).Save(context.Background()); err != nil {
			return err
		}
		logger.Info("Certificate expired", zap.String("common_name", cert.CommonName), zap.String("serial_number", cert.Serial))
	}
	return nil
}
