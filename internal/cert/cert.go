package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	CACommonName   = "Trae Switch Root CA"
	CAOrganization = "Trae Switch"
	CertValidity   = 365 * 24 * time.Hour
)

type CertificateManager struct {
	dataDir       string
	caCert        *x509.Certificate
	caKey         *rsa.PrivateKey
	caCertPEM     []byte
	serverCert    *x509.Certificate
	serverKey     *rsa.PrivateKey
	serverCertPEM []byte
	serverKeyPEM  []byte
}

func NewCertificateManager(dataDir string) *CertificateManager {
	return &CertificateManager{
		dataDir: dataDir,
	}
}

func (cm *CertificateManager) LoadOrGenerateCA() error {
	caCertPath := filepath.Join(cm.dataDir, "ca.crt")
	caKeyPath := filepath.Join(cm.dataDir, "ca.key")

	if _, err := os.Stat(caCertPath); err == nil {
		if _, err := os.Stat(caKeyPath); err == nil {
			return cm.loadCA(caCertPath, caKeyPath)
		}
	}

	return cm.generateAndSaveCA()
}

func (cm *CertificateManager) loadCA(certPath, keyPath string) error {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return errors.New("failed to decode CA certificate PEM")
	}

	caCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return err
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return errors.New("failed to decode CA key PEM")
	}

	caKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return err
	}

	cm.caCert = caCert
	cm.caKey = caKey
	cm.caCertPEM = certPEM
	return nil
}

func (cm *CertificateManager) generateAndSaveCA() error {
	if err := os.MkdirAll(cm.dataDir, 0755); err != nil {
		return err
	}

	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   CACommonName,
			Organization: []string{CAOrganization},
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &caKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	caCert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return err
	}

	cm.caCert = caCert
	cm.caKey = caKey
	cm.caCertPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	caKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caKey),
	})

	caCertPath := filepath.Join(cm.dataDir, "ca.crt")
	caKeyPath := filepath.Join(cm.dataDir, "ca.key")

	if err := os.WriteFile(caCertPath, cm.caCertPEM, 0644); err != nil {
		return err
	}

	if err := os.WriteFile(caKeyPath, caKeyPEM, 0600); err != nil {
		return err
	}

	return nil
}

func (cm *CertificateManager) GenerateServerCert(domain string) error {
	if cm.caCert == nil || cm.caKey == nil {
		return errors.New("CA certificate not loaded")
	}

	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   domain,
			Organization: []string{CAOrganization},
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(CertValidity),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    []string{domain},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, cm.caCert, &serverKey.PublicKey, cm.caKey)
	if err != nil {
		return err
	}

	serverCert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return err
	}

	cm.serverCert = serverCert
	cm.serverKey = serverKey
	cm.serverCertPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})
	cm.serverKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverKey),
	})

	return nil
}

func (cm *CertificateManager) GetCACertPEM() []byte {
	return cm.caCertPEM
}

func (cm *CertificateManager) GetCACertPath() string {
	return filepath.Join(cm.dataDir, "ca.crt")
}

func (cm *CertificateManager) GetServerCertPEM() []byte {
	return cm.serverCertPEM
}

func (cm *CertificateManager) GetServerKeyPEM() []byte {
	return cm.serverKeyPEM
}

func (cm *CertificateManager) GetServerTLSCertificate() (*x509.Certificate, *rsa.PrivateKey) {
	return cm.serverCert, cm.serverKey
}

func (cm *CertificateManager) CACertificateExists() bool {
	caCertPath := filepath.Join(cm.dataDir, "ca.crt")
	_, err := os.Stat(caCertPath)
	return err == nil
}
