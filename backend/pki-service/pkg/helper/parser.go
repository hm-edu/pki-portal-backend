package helper

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// LoadDER parses a certificate from DER encoded data.
func LoadDER(der [][]byte) ([]*x509.Certificate, error) {

	certs := make([]*x509.Certificate, len(der))
	for i, data := range der {
		item, err := x509.ParseCertificate(data)
		if err != nil {
			return nil, err
		}
		certs[i] = item
	}

	return certs, nil
}

// ParseCertificates parses a certificate from PEM encoded data.
func ParseCertificates(cert []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	for block, rest := pem.Decode(cert); block != nil; block, rest = pem.Decode(rest) {
		switch block.Type {
		case "CERTIFICATE":
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}
			certs = append(certs, cert)
		default:
			return nil, errors.New("Unknown entry in cert chain")
		}
	}
	return certs, nil
}

const (
	pemTypeKey  = "EC PRIVATE KEY"
	pemTypeCert = "CERTIFICATE"
)

// EncodePem encodes certificates and key in PEM format
func EncodePem(key *ecdsa.PrivateKey, certs [][]byte) ([]byte, error) {
	var buf bytes.Buffer

	for _, cert := range certs {
		err := pem.Encode(&buf, &pem.Block{
			Type:  pemTypeCert,
			Bytes: cert,
		})
		if err != nil {
			return nil, err
		}
	}

	if key != nil {
		keyBytes, err := x509.MarshalECPrivateKey(key)
		if err != nil {
			return nil, err
		}
		err = pem.Encode(&buf, &pem.Block{
			Type:  pemTypeKey,
			Bytes: keyBytes,
		})
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// FlattenStringSlice joins strings from a slice with commas for printing
func FlattenStringSlice(stringSlice []string) string {
	if len(stringSlice) == 0 {
		return ""
	}
	flattened := ""
	for _, element := range stringSlice {
		flattened = flattened + element + ","
	}
	flattened = flattened[:len(flattened)-1] // Remove trailing comma
	return flattened
}
