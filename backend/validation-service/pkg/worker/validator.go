package worker

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/sectigo-client/sectigo"
	"github.com/hm-edu/sectigo-client/sectigo/domain"
	"go.uber.org/zap"
)

// DomainValidator holds the required clients/data for the domain validation.
type DomainValidator struct {
	Client     *sectigo.Client
	DNSService pb.DNSServiceClient
	Force      bool
	Domains    []string
}

// ValidateDomains validates the domains returned from sectigo.
func (v *DomainValidator) ValidateDomains() {
	logger := zap.L()
	validations, err := v.Client.DomainValidationService.List()
	if err != nil {
		logger.Fatal("fetching domain validations failed", zap.Error(err))
	}

	if len(v.Domains) == 0 {
		logger.Warn("No filter given. Validating all domains")
	}

	for _, validation := range *validations {
		logger.Debug("Handling existing validation", zap.String("domain", validation.Domain), zap.Time("expires", validation.ExpirationDate.Time))

		if len(v.Domains) != 0 && !helper.Contains(v.Domains, validation.Domain) {
			logger.Info("Domain not in filter", zap.String("domain", validation.Domain))
			continue
		} else if len(v.Domains) == 0 && strings.Count(validation.Domain, ".") != 1 {
			logger.Info("Skipping sub-domain", zap.String("domain", validation.Domain))
			continue
		}

		if validation.ExpirationDate.Time.Before(time.Now().Add(7 * 24 * time.Hour)) {
			logger.Info("Domain validation is not required", zap.String("domain", validation.Domain))
			if !v.Force {
				continue
			}
			logger.Info("Forcing domain validation", zap.String("domain", validation.Domain))
		}
		data, err := v.Client.DomainValidationService.StartCNAME(validation.Domain)
		if err != nil {
			logger.Error("Failed to start validation", zap.Error(err), zap.String("domain", validation.Domain))
			continue
		}
		logger.Info("Started validation", zap.String("domain", validation.Domain))
		_, err = v.DNSService.Add(context.Background(), &pb.AddRequest{
			Zone: fmt.Sprintf("%s.", validation.Domain),
			Records: []*pb.DNSRecord{
				{Ttl: 3600, Type: "CNAME", Name: data.Host, Content: data.Point},
			}})
		if err != nil {
			logger.Error("Failed to add validation record", zap.Error(err), zap.String("domain", validation.Domain))
			continue
		}
		logger.Info("Added validation record", zap.String("domain", validation.Domain))
		resp, err := v.Client.DomainValidationService.SubmitCNAME(validation.Domain)
		if err != nil {
			logger.Error("Failed to submit validation", zap.Error(err), zap.String("domain", validation.Domain))
			continue
		}
		logger.Info("Submitted validation", zap.String("domain", validation.Domain), zap.Any("response", resp))
		err = helper.WaitFor(10*time.Minute, 10*time.Second, func() (bool, error) {
			status, err := v.Client.DomainValidationService.Status(validation.Domain)
			if err != nil {
				logger.Error("Failed to get validation status", zap.Error(err), zap.String("domain", validation.Domain))
				return false, nil
			}
			if status.Status == domain.Validated {
				logger.Info("Validation succeeded", zap.String("domain", validation.Domain))
				return true, nil
			}
			logger.Info("Validation not yet finished", zap.String("domain", validation.Domain))
			return false, nil
		})
		if err != nil {
			logger.Error("Validation failed", zap.Error(err), zap.String("domain", validation.Domain))
			continue
		}

		records, err := v.DNSService.List(context.Background(), &pb.ListRequest{Zone: fmt.Sprintf("%s.", validation.Domain)})
		if err != nil {
			logger.Error("Failed to list dns records", zap.Error(err), zap.String("domain", validation.Domain))
			continue
		}
		for _, record := range records.Records {
			if record.Type == "CNAME" && record.Name == data.Host && record.Content == data.Point {
				logger.Info("Deleting validation record", zap.String("domain", validation.Domain))

				_, err = v.DNSService.Delete(context.Background(), &pb.DeleteRequest{
					Zone:    fmt.Sprintf("%s.", validation.Domain),
					Records: []*pb.DNSRecord{record}})
				if err != nil {
					logger.Error("Failed to delete validation record", zap.Error(err), zap.String("domain", validation.Domain))
					continue
				}
				logger.Debug("Deleted validation record", zap.String("domain", validation.Domain))
			}
		}
	}
}
