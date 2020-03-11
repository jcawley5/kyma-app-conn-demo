package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"math/big"
	"strings"
	"time"
)

//KymaCerts - Contains the kyma certificate, key and csr
type KymaCerts struct {
	CRT        []byte
	PrivateKey []byte
	CSR        []byte
}

//GenerateCSR - generate a csr based on the received subject
//subject - csr subject string
func GenerateCSR(subject string, keylength int) (*KymaCerts, error) {

	keyBytes, _ := rsa.GenerateKey(rand.Reader, keylength)

	subjectTrimed := strings.TrimSuffix(subject, ",")
	entries := strings.Split(subjectTrimed, ",")
	subjectMapped := make(map[string]string)

	for _, e := range entries {
		parts := strings.Split(e, "=")
		subjectMapped[parts[0]] = parts[1]
	}

	subj := pkix.Name{
		CommonName:         subjectMapped["CN"],
		Country:            []string{subjectMapped["C"]},
		Province:           []string{subjectMapped["ST"]},
		Locality:           []string{subjectMapped["L"]},
		Organization:       []string{subjectMapped["O"]},
		OrganizationalUnit: []string{subjectMapped["OU"]},
	}

	type basicConstraints struct {
		IsCA       bool `asn1:"optional"`
		MaxPathLen int  `asn1:"optional,default:-1"`
	}

	val, _ := asn1.Marshal(basicConstraints{true, 0})

	var csrTemplate = x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
		ExtraExtensions: []pkix.Extension{
			{
				Id:       asn1.ObjectIdentifier{2, 5, 29, 19},
				Value:    val,
				Critical: true,
			},
		},
	}

	csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keyBytes)

	csr := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrBytes,
	})

	// step: generate a serial number
	serial, _ := rand.Int(rand.Reader, (&big.Int{}).Exp(big.NewInt(2), big.NewInt(159), nil))

	now := time.Now()
	// step: create the request template
	template := x509.Certificate{
		SerialNumber:          serial,
		Subject:               subj,
		NotBefore:             now.Add(-10 * time.Minute).UTC(),
		NotAfter:              now.Add(time.Duration(1200)).UTC(),
		BasicConstraintsValid: true,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	// step: sign the certificate authority
	certificate, _ := x509.CreateCertificate(rand.Reader, &template, &template, &keyBytes.PublicKey, keyBytes)

	clientCrt := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate})

	privateKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(keyBytes)})

	return &KymaCerts{
		PrivateKey: privateKey,
		CRT:        clientCrt,
		CSR:        csr,
	}, nil
}
