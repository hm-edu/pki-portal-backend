package worker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/sectigo-client/sectigo"
	"github.com/hm-edu/sectigo-client/sectigo/dcv"
	"github.com/hm-edu/sectigo-client/sectigo/domain"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
	"go.uber.org/zap"
)

// DomainValidator holds the required clients/data for the domain validation.
type DomainValidator struct {
	Client     *sectigo.Client
	DNSService pb.DNSServiceClient
	Domains    []string

	observerLock          *sync.RWMutex
	observerValueToReport map[string]time.Duration
}

var meter = global.MeterProvider().Meter("validation-service")

// ValidateDomains validates the domains returned from sectigo.
func (v *DomainValidator) ValidateDomains() {

	logger := zap.L()
	v.observerLock = new(sync.RWMutex)
	v.observerValueToReport = make(map[string]time.Duration)
	gaugeObserver, err := meter.AsyncInt64().Gauge("remaining_days", instrument.WithDescription("The remaining validation days per domain"), instrument.WithUnit(unit.Unit("days")))
	if err != nil {
		logger.Panic("failed to initialize instrument: %v", zap.Error(err))
	}
	_ = meter.RegisterCallback([]instrument.Asynchronous{gaugeObserver}, func(ctx context.Context) {
		(*v.observerLock).RLock()
		data := v.observerValueToReport
		(*v.observerLock).RUnlock()
		for domain, duration := range data {
			gaugeObserver.Observe(ctx, int64(duration.Hours()/24), attribute.String("domain", domain))
		}
	})

	validations, err := v.Client.DomainValidationService.List()
	if err != nil {
		logger.Fatal("fetching domain validations failed", zap.Error(err))
	}

	if len(v.Domains) == 0 {
		logger.Warn("No filter given. Validating all domains")
	}
	pending := make(map[time.Duration][]string)
	x := time.Now().Add(7 * 24 * time.Hour)
	for _, validation := range *validations {
		logger.Debug("Handling existing validation", zap.String("domain", validation.Domain), zap.Time("expires", validation.ExpirationDate.Time))

		if len(v.Domains) != 0 && !helper.Contains(v.Domains, validation.Domain) {
			logger.Info("Domain not in filter", zap.String("domain", validation.Domain))
			continue
		} else if len(v.Domains) == 0 && strings.Count(validation.Domain, ".") != 1 {
			logger.Info("Skipping sub-domain", zap.String("domain", validation.Domain))
			continue
		}

		var duration time.Duration
		if validation.ExpirationDate.Time.Before(x) {
			duration = 1 * time.Hour
		} else {
			duration = validation.ExpirationDate.Time.Sub(x)
		}
		pending[duration] = append(pending[duration], validation.Domain)
		(*v.observerLock).Lock()
		v.observerValueToReport[validation.Domain] = time.Until(validation.ExpirationDate.Time)
		(*v.observerLock).Unlock()
	}

	for duration, domains := range pending {
		go v.validateDomains(duration, domains)
	}

}

func (v *DomainValidator) validateDomains(duration time.Duration, domains []string) {
	time.Sleep(duration)
	pending := make(map[time.Duration][]string)
	x := time.Now().Add(7 * 24 * time.Hour)
	for _, d := range domains {
		logger := zap.L().With(zap.String("domain", d))
		data, err := v.Client.DomainValidationService.StartCNAME(d)
		if err != nil {
			logger.Error("Failed to start validation", zap.Error(err))
			continue
		}
		logger.Info("Started validation")
		_, err = v.DNSService.Add(context.Background(), &pb.AddRequest{
			Zone: fmt.Sprintf("%s.", d),
			Records: []*pb.DNSRecord{
				{Ttl: 3600, Type: "CNAME", Name: data.Host, Content: data.Point},
			}})
		if err != nil {
			logger.Error("Failed to add validation record", zap.Error(err))
			continue
		}
		logger.Info("Added validation record")
		resp, err := v.Client.DomainValidationService.SubmitCNAME(d)
		if err != nil {
			logger.Error("Failed to submit validation", zap.Error(err))
			continue
		}
		logger.Info("Submitted validation", zap.Any("response", resp))
		var status *dcv.StatusResponse
		err = helper.WaitFor(10*time.Minute, 10*time.Second, func() (bool, error) {
			status, err = v.Client.DomainValidationService.Status(d)
			if err != nil {
				logger.Error("Failed to get validation status", zap.Error(err))
				return false, nil
			}
			if status.Status == domain.Validated {
				logger.Info("Validation succeeded")
				return true, nil
			}
			logger.Info("Validation not yet finished")
			return false, nil
		})
		if err != nil {
			logger.Error("Validation failed", zap.Error(err))
			pending[1*24*time.Hour] = append(pending[1*24*time.Hour], d)
			continue
		}

		records, err := v.DNSService.List(context.Background(), &pb.ListRequest{Zone: fmt.Sprintf("%s.", d)})
		if err != nil {
			logger.Error("Failed to list dns records", zap.Error(err))
			continue
		}
		for _, record := range records.Records {
			if record.Type == "CNAME" && record.Name == data.Host && record.Content == data.Point {
				logger.Info("Deleting validation record")

				_, err = v.DNSService.Delete(context.Background(), &pb.DeleteRequest{
					Zone:    fmt.Sprintf("%s.", d),
					Records: []*pb.DNSRecord{record}})
				if err != nil {
					logger.Error("Failed to delete validation record", zap.Error(err))
					continue
				}
				logger.Debug("Deleted validation record")
			}
		}
		pending[status.ExpirationDate.Time.Sub(x)] = append(pending[status.ExpirationDate.Time.Sub(x)], d)
	}

	for duration, domains := range pending {
		go v.validateDomains(duration, domains)
	}

}
