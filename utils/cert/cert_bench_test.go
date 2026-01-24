package cert

import (
	"crypto/x509"
	"net"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkNewPrivateKey_RSA RSA 私钥生成性能
func BenchmarkNewPrivateKey_RSA(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewPrivateKey(KeyTypeRSA)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNewPrivateKey_ECDSA ECDSA 私钥生成性能
func BenchmarkNewPrivateKey_ECDSA(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewPrivateKey(KeyTypeECDSA)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNewCA_RSA 创建 RSA CA 的性能
func BenchmarkNewCA_RSA(b *testing.B) {
	cfg := Config{
		CommonName:   "Test CA",
		Organization: []string{"Test Org"},
		ValidYears:   10,
		KeyType:      KeyTypeRSA,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewCA(cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNewCA_ECDSA 创建 ECDSA CA 的性能
func BenchmarkNewCA_ECDSA(b *testing.B) {
	cfg := Config{
		CommonName:   "Test CA",
		Organization: []string{"Test Org"},
		ValidYears:   10,
		KeyType:      KeyTypeECDSA,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewCA(cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCA_NewSignedCert 签发证书的性能
func BenchmarkCA_NewSignedCert(b *testing.B) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
		KeyType:    KeyTypeRSA,
	})
	if err != nil {
		b.Fatal(err)
	}

	cfg := Config{
		CommonName: "server.example.com",
		ValidYears: 1,
		KeyType:    KeyTypeRSA,
		AltNames: AltNames{
			DNSNames: []string{"server.example.com"},
		},
		Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ca.NewSignedCert(cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCA_NewSignedCert_ECDSA 签发 ECDSA 证书的性能
func BenchmarkCA_NewSignedCert_ECDSA(b *testing.B) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
		KeyType:    KeyTypeECDSA,
	})
	if err != nil {
		b.Fatal(err)
	}

	cfg := Config{
		CommonName: "server.example.com",
		ValidYears: 1,
		KeyType:    KeyTypeECDSA,
		AltNames: AltNames{
			DNSNames: []string{"server.example.com"},
		},
		Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ca.NewSignedCert(cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEncodeCertPEM 证书编码性能
func BenchmarkEncodeCertPEM(b *testing.B) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeCertPEM(ca.Certificate)
	}
}

// BenchmarkEncodePrivateKeyPEM_RSA RSA 私钥编码性能
func BenchmarkEncodePrivateKeyPEM_RSA(b *testing.B) {
	key, err := NewPrivateKey(KeyTypeRSA)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EncodePrivateKeyPEM(key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEncodePrivateKeyPEM_ECDSA ECDSA 私钥编码性能
func BenchmarkEncodePrivateKeyPEM_ECDSA(b *testing.B) {
	key, err := NewPrivateKey(KeyTypeECDSA)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EncodePrivateKeyPEM(key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseCertsPEM 证书解析性能
func BenchmarkParseCertsPEM(b *testing.B) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	pemData := EncodeCertPEM(ca.Certificate)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseCertsPEM(pemData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParsePrivateKeyPEM_RSA RSA 私钥解析性能
func BenchmarkParsePrivateKeyPEM_RSA(b *testing.B) {
	key, err := NewPrivateKey(KeyTypeRSA)
	if err != nil {
		b.Fatal(err)
	}

	pemData, err := EncodePrivateKeyPEM(key)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParsePrivateKeyPEM(pemData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParsePrivateKeyPEM_ECDSA ECDSA 私钥解析性能
func BenchmarkParsePrivateKeyPEM_ECDSA(b *testing.B) {
	key, err := NewPrivateKey(KeyTypeECDSA)
	if err != nil {
		b.Fatal(err)
	}

	pemData, err := EncodePrivateKeyPEM(key)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParsePrivateKeyPEM(pemData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWriteCertToFile 证书写入文件性能
func BenchmarkWriteCertToFile(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "cert-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		certPath := filepath.Join(tmpDir, "cert.crt")
		err := WriteCertToFile(certPath, ca.Certificate)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkReadCertFromFile 证书读取文件性能
func BenchmarkReadCertFromFile(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "cert-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	certPath := filepath.Join(tmpDir, "cert.crt")
	err = WriteCertToFile(certPath, ca.Certificate)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ReadCertFromFile(certPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCAWorkflow 完整的 CA 工作流性能
func BenchmarkCAWorkflow(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 创建 CA
		ca, err := NewCA(Config{
			CommonName: "Test CA",
			ValidYears: 10,
			KeyType:    KeyTypeRSA,
		})
		if err != nil {
			b.Fatal(err)
		}

		// 签发证书
		_, err = ca.NewSignedCert(Config{
			CommonName: "server.example.com",
			ValidYears: 1,
			KeyType:    KeyTypeRSA,
			AltNames: AltNames{
				DNSNames: []string{"server.example.com"},
				IPs:      []net.IP{net.ParseIP("192.168.1.1")},
			},
			Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParallel_NewCA 并发创建 CA 性能
func BenchmarkParallel_NewCA(b *testing.B) {
	cfg := Config{
		CommonName: "Test CA",
		ValidYears: 10,
		KeyType:    KeyTypeRSA,
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := NewCA(cfg)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkParallel_NewSignedCert 并发签发证书性能
func BenchmarkParallel_NewSignedCert(b *testing.B) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
		KeyType:    KeyTypeRSA,
	})
	if err != nil {
		b.Fatal(err)
	}

	cfg := Config{
		CommonName: "server.example.com",
		ValidYears: 1,
		KeyType:    KeyTypeRSA,
		AltNames: AltNames{
			DNSNames: []string{"server.example.com"},
		},
		Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := ca.NewSignedCert(cfg)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkEncodeCertToBase64 证书 base64 编码性能
func BenchmarkEncodeCertToBase64(b *testing.B) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EncodeCertToBase64(ca.Certificate)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEncodePrivateKeyToBase64 私钥 base64 编码性能
func BenchmarkEncodePrivateKeyToBase64(b *testing.B) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EncodePrivateKeyToBase64(ca.PrivateKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseCertFromBase64 证书 base64 解码性能
func BenchmarkParseCertFromBase64(b *testing.B) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	certBase64, err := EncodeCertToBase64(ca.Certificate)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseCertFromBase64(certBase64)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParsePrivateKeyFromBase64 私钥 base64 解码性能
func BenchmarkParsePrivateKeyFromBase64(b *testing.B) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	keyBase64, err := EncodePrivateKeyToBase64(ca.PrivateKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParsePrivateKeyFromBase64(keyBase64)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCA_ToBase64 CA 转 base64 性能
func BenchmarkCA_ToBase64(b *testing.B) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := ca.ToBase64()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLoadCAFromBase64 从 base64 加载 CA 性能
func BenchmarkLoadCAFromBase64(b *testing.B) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	certBase64, keyBase64, err := ca.ToBase64()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadCAFromBase64(certBase64, keyBase64)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCertKeyPair_ToBase64 证书对转 base64 性能
func BenchmarkCertKeyPair_ToBase64(b *testing.B) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	certPair, err := ca.NewSignedCert(Config{
		CommonName: "server",
		ValidYears: 1,
		Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := certPair.ToBase64()
		if err != nil {
			b.Fatal(err)
		}
	}
}
