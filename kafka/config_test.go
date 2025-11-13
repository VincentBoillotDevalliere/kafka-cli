package kafka

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLookupEnvBool(t *testing.T) {
	t.Setenv("KAFKA_TEST_BOOL", "true")
	if val, ok := lookupEnvBool("KAFKA_TEST_BOOL"); !ok || !val {
		t.Fatalf("expected true, got %v (ok=%v)", val, ok)
	}

	t.Setenv("KAFKA_TEST_BOOL", "false")
	if val, ok := lookupEnvBool("KAFKA_TEST_BOOL"); !ok || val {
		t.Fatalf("expected false, got %v (ok=%v)", val, ok)
	}

	t.Setenv("KAFKA_TEST_BOOL", "not-a-bool")
	if _, ok := lookupEnvBool("KAFKA_TEST_BOOL"); ok {
		t.Fatalf("expected invalid bool value to be ignored")
	}
}

func TestNewConfigRespectsTLSEnv(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("KAFKA_TLS_ENABLED", "false")

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.TLSEnabled {
		t.Fatalf("expected TLS to be disabled when KAFKA_TLS_ENABLED=false")
	}
	if cfg.tlsConfig != nil {
		t.Fatalf("expected tlsConfig to be nil when TLS disabled")
	}
}

func TestBuildTLSConfigFromEnvLoadsCA(t *testing.T) {
	certPath := writeTestCA(t)
	t.Setenv("KAFKA_TLS_CA_FILE", certPath)
	t.Setenv("KAFKA_TLS_INSECURE_SKIP_VERIFY", "true")

	cfg, err := buildTLSConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error building TLS config, got %v", err)
	}
	if cfg.RootCAs == nil {
		t.Fatalf("expected RootCAs to be populated from CA file")
	}
	if !cfg.InsecureSkipVerify {
		t.Fatalf("expected InsecureSkipVerify to be true")
	}
}

func writeTestCA(t *testing.T) string {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "kafka-cli-test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if len(pemBytes) == 0 {
		t.Fatalf("failed to encode certificate to PEM")
	}

	path := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		t.Fatalf("failed to write CA file: %v", err)
	}
	return path
}
