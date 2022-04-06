package smime

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hm-edu/pki-service/pkg/cfg"
	"github.com/hm-edu/sectigo-client/sectigo"
	"github.com/jarcoal/httpmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSmimeCsrInvalid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"csr":1}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test", "given_name": "Test", "family_name": "Test", "name": "Test"}})
	h := NewHandler(sectigo.NewClient(http.DefaultClient, zap.L(), "", "", ""), &cfg.SectigoConfiguration{})
	assert.Error(t, h.HandleCsr(c))
}

func TestSmimeShortKey(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://cert-manager.com/api/smime/v1/enroll", httpmock.NewStringResponder(200, `{"orderNumber":123,"backendCertId":"123"}`))
	httpmock.RegisterResponder("GET", "https://cert-manager.com/api/smime/v1/collect/123?format=x509", httpmock.NewStringResponder(200, `Test`))

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"csr":"-----BEGIN CERTIFICATE REQUEST-----\r\nMIIERTCCAi0CAQAwADCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAKos\r\nwbYwuT+Ft6czXMZ4OHfBbdoHwcTY+dOpTyxTMXP+up77nAN5yb8dBdJjSDEVj4ZX\r\nQyPC+bVk6pX8IcWwnKfd4IG0250Ydfqwr/SIVUUVgNXu+vGaM8QfmXRi+C3q9kqX\r\n+qlhmJkQ5xpGhDGFtqR0e61Bz28ko0xwtJpXq6tLIRU3LpRtBH3wg28Iv+Pgnh8L\r\ngKgipD3sGEnJLLG1jeOL8TyTK4daGlzbYpdpWyUCFHVAlvUCVqvje/zd6zi34Ba/\r\nqd8MtBYIr7OBPLh2ADuLZT2Xq3ZDowFiPEerOOIZV9w7pgzkyvYmLedQ9Oyz5YfC\r\nGbME71dumae3DrmbdQePvpkPLrMt4X9XFZ46Gmr9q67kI7ip8Tut/ZVF7CAPlpwT\r\nnNAcgs/hL5j14N5H8f14RHQRXmi7N8mw99NpABd+xgWZr9GjjnZhTW+X7nzya9fi\r\n/09WazFqVjhsExM7lptQVofwLBy4HpCO7MFrQwLo1eJz9/j/HaM0JrN87SpF3zFr\r\nnLkAoDwZ0nzarZ4geai3ZpathX8Q2q+Lk0N2MGpocv7dalYDQzavdcITH6OnQNBl\r\n4aXDZg+9RtHoHXG2kv+mdSFj6QU1Jt2Jy2LUBIr/KBk1KGj1uh+Sisdci9LPo1d5\r\nyRTkb4MsuAWOgjgSPI31pbWt5nuPhl1b2SOOlRoBAgMBAAGgADANBgkqhkiG9w0B\r\nAQsFAAOCAgEAHLAkHbXCxbhLY3UeyS5kwW8vXxiZ7K8Q9XJRrCbctJBtRp19DPU6\r\nNgLY+69byosbWmNC3lLfYZGPXjVnSOkim2bKRjGVlC2ANi46yVGtwNqSS+/VfJ6B\r\nLmBdUedRZNoZBMiuSew/8erHCilStZATuM7YpyLJPBH6a4IN9oT6qbF3YhgLOzuZ\r\nc68798BFl5ld07fof+f84ci1PSUmt3GGEYAYCrIvIhJvP4817Q+u3CdkGpUM0nKv\r\nXjP1FEISfsBrIUYZ7hlDx3HXlUDDuH5UWzQ2bNqv6uYt7jOND2kfCLHZxLukpimm\r\nNpkPW0qLYT76wWvKENuEBu6bmYOLTbai7NJthRaUg6ObcqN5oM9FU8Vs/mRf1Deb\r\nCOc56UwyQpwpSFJ1vqNCVGJKXNZO31EemFm2xEmGBI3Cy0JU70X8T9QmPu2ak4ce\r\n/SjQfeDoR2lKeNzpuH5fFBE2oDxU/AdXJCQ1qiQ2ZdpyURbEJAS/EE0LLjPiYlBa\r\nTxDdKk1y/d4x3E9/Y4fwn0ipp4O3exOGiuof99gxQTrAugn8lfO5NihZm24hIMY5\r\n2x7oTIRSRXTNrtPLfUvWpWySrVqyVO0I+lwn9JZja2HbSq1hTbkj1D6TI4d+x17V\r\nuQYbocrA7U7EKOCExOAujpWvdntIIB8egp32hD0hBzExaBvzOrSw9Ys=\r\n-----END CERTIFICATE REQUEST-----\r\n"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test", "given_name": "Test", "family_name": "Test", "name": "Test"}})
	h := NewHandler(sectigo.NewClient(http.DefaultClient, zap.L(), "", "", ""), &cfg.SectigoConfiguration{SmimeKeyType: "RSA", SmimeKeyLength: "8192"})
	assert.Error(t, h.HandleCsr(c))
}

func TestSmimeValid(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://cert-manager.com/api/smime/v1/enroll", httpmock.NewStringResponder(200, `{"orderNumber":123,"backendCertId":"123"}`))
	httpmock.RegisterResponder("GET", "https://cert-manager.com/api/smime/v1/collect/123?format=x509", httpmock.NewStringResponder(200, `Test`))

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"csr":"-----BEGIN CERTIFICATE REQUEST-----\r\nMIIERTCCAi0CAQAwADCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAKos\r\nwbYwuT+Ft6czXMZ4OHfBbdoHwcTY+dOpTyxTMXP+up77nAN5yb8dBdJjSDEVj4ZX\r\nQyPC+bVk6pX8IcWwnKfd4IG0250Ydfqwr/SIVUUVgNXu+vGaM8QfmXRi+C3q9kqX\r\n+qlhmJkQ5xpGhDGFtqR0e61Bz28ko0xwtJpXq6tLIRU3LpRtBH3wg28Iv+Pgnh8L\r\ngKgipD3sGEnJLLG1jeOL8TyTK4daGlzbYpdpWyUCFHVAlvUCVqvje/zd6zi34Ba/\r\nqd8MtBYIr7OBPLh2ADuLZT2Xq3ZDowFiPEerOOIZV9w7pgzkyvYmLedQ9Oyz5YfC\r\nGbME71dumae3DrmbdQePvpkPLrMt4X9XFZ46Gmr9q67kI7ip8Tut/ZVF7CAPlpwT\r\nnNAcgs/hL5j14N5H8f14RHQRXmi7N8mw99NpABd+xgWZr9GjjnZhTW+X7nzya9fi\r\n/09WazFqVjhsExM7lptQVofwLBy4HpCO7MFrQwLo1eJz9/j/HaM0JrN87SpF3zFr\r\nnLkAoDwZ0nzarZ4geai3ZpathX8Q2q+Lk0N2MGpocv7dalYDQzavdcITH6OnQNBl\r\n4aXDZg+9RtHoHXG2kv+mdSFj6QU1Jt2Jy2LUBIr/KBk1KGj1uh+Sisdci9LPo1d5\r\nyRTkb4MsuAWOgjgSPI31pbWt5nuPhl1b2SOOlRoBAgMBAAGgADANBgkqhkiG9w0B\r\nAQsFAAOCAgEAHLAkHbXCxbhLY3UeyS5kwW8vXxiZ7K8Q9XJRrCbctJBtRp19DPU6\r\nNgLY+69byosbWmNC3lLfYZGPXjVnSOkim2bKRjGVlC2ANi46yVGtwNqSS+/VfJ6B\r\nLmBdUedRZNoZBMiuSew/8erHCilStZATuM7YpyLJPBH6a4IN9oT6qbF3YhgLOzuZ\r\nc68798BFl5ld07fof+f84ci1PSUmt3GGEYAYCrIvIhJvP4817Q+u3CdkGpUM0nKv\r\nXjP1FEISfsBrIUYZ7hlDx3HXlUDDuH5UWzQ2bNqv6uYt7jOND2kfCLHZxLukpimm\r\nNpkPW0qLYT76wWvKENuEBu6bmYOLTbai7NJthRaUg6ObcqN5oM9FU8Vs/mRf1Deb\r\nCOc56UwyQpwpSFJ1vqNCVGJKXNZO31EemFm2xEmGBI3Cy0JU70X8T9QmPu2ak4ce\r\n/SjQfeDoR2lKeNzpuH5fFBE2oDxU/AdXJCQ1qiQ2ZdpyURbEJAS/EE0LLjPiYlBa\r\nTxDdKk1y/d4x3E9/Y4fwn0ipp4O3exOGiuof99gxQTrAugn8lfO5NihZm24hIMY5\r\n2x7oTIRSRXTNrtPLfUvWpWySrVqyVO0I+lwn9JZja2HbSq1hTbkj1D6TI4d+x17V\r\nuQYbocrA7U7EKOCExOAujpWvdntIIB8egp32hD0hBzExaBvzOrSw9Ys=\r\n-----END CERTIFICATE REQUEST-----\r\n"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test", "given_name": "Test", "family_name": "Test", "name": "Test"}})
	h := NewHandler(sectigo.NewClient(http.DefaultClient, zap.L(), "", "", ""), &cfg.SectigoConfiguration{SmimeKeyType: "RSA", SmimeKeyLength: "4096"})
	assert.NoError(t, h.HandleCsr(c))
}
