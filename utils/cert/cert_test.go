package cert

import (
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewPrivateKey(t *testing.T) {
	tests := []struct {
		name    string
		keyType KeyType
		wantErr bool
	}{
		{
			name:    "RSA key",
			keyType: KeyTypeRSA,
			wantErr: false,
		},
		{
			name:    "ECDSA key",
			keyType: KeyTypeECDSA,
			wantErr: false,
		},
		{
			name:    "Default (empty) key type",
			keyType: "",
			wantErr: false,
		},
		{
			name:    "Invalid key type",
			keyType: "INVALID",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := NewPrivateKey(tt.keyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && key == nil {
				t.Error("NewPrivateKey() returned nil key")
			}
		})
	}
}

func TestNewCA(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "Valid CA with RSA",
			cfg: Config{
				CommonName:   "Test CA",
				Organization: []string{"Test Org"},
				ValidYears:   10,
				KeyType:      KeyTypeRSA,
			},
			wantErr: false,
		},
		{
			name: "Valid CA with ECDSA",
			cfg: Config{
				CommonName:   "Test CA",
				Organization: []string{"Test Org"},
				ValidYears:   10,
				KeyType:      KeyTypeECDSA,
			},
			wantErr: false,
		},
		{
			name: "Valid CA with default values",
			cfg: Config{
				CommonName: "Test CA",
			},
			wantErr: false,
		},
		{
			name:    "Missing common name",
			cfg:     Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca, err := NewCA(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if ca == nil {
					t.Error("NewCA() returned nil CA")
					return
				}
				if ca.Certificate == nil {
					t.Error("NewCA() returned nil Certificate")
				}
				if ca.PrivateKey == nil {
					t.Error("NewCA() returned nil PrivateKey")
				}

				// 验证证书属性
				if ca.Certificate.Subject.CommonName != tt.cfg.CommonName {
					t.Errorf("CA CommonName = %v, want %v", ca.Certificate.Subject.CommonName, tt.cfg.CommonName)
				}
				if !ca.Certificate.IsCA {
					t.Error("CA certificate IsCA should be true")
				}

				// 验证有效期
				now := time.Now()
				if ca.Certificate.NotBefore.After(now) {
					t.Error("CA certificate NotBefore is in the future")
				}
				if ca.Certificate.NotAfter.Before(now) {
					t.Error("CA certificate NotAfter is in the past")
				}
			}
		})
	}
}

func TestCA_NewSignedCert(t *testing.T) {
	// 先创建一个 CA
	ca, err := NewCA(Config{
		CommonName:   "Test CA",
		Organization: []string{"Test Org"},
		ValidYears:   10,
		KeyType:      KeyTypeRSA,
	})
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "Valid server certificate",
			cfg: Config{
				CommonName:   "test.example.com",
				Organization: []string{"Test Org"},
				ValidYears:   1,
				KeyType:      KeyTypeRSA,
				AltNames: AltNames{
					DNSNames: []string{"test.example.com", "*.test.example.com"},
					IPs:      []net.IP{net.ParseIP("127.0.0.1")},
				},
				Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			},
			wantErr: false,
		},
		{
			name: "Valid client certificate",
			cfg: Config{
				CommonName:   "client",
				Organization: []string{"Test Org"},
				ValidYears:   1,
				KeyType:      KeyTypeECDSA,
				Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
			wantErr: false,
		},
		{
			name: "Missing common name",
			cfg: Config{
				Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			},
			wantErr: true,
		},
		{
			name: "Missing usages",
			cfg: Config{
				CommonName: "test.example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			certPair, err := ca.NewSignedCert(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("CA.NewSignedCert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if certPair == nil {
					t.Error("CA.NewSignedCert() returned nil CertKeyPair")
					return
				}
				if certPair.Certificate == nil {
					t.Error("CA.NewSignedCert() returned nil Certificate")
				}
				if certPair.PrivateKey == nil {
					t.Error("CA.NewSignedCert() returned nil PrivateKey")
				}

				// 验证证书属性
				if certPair.Certificate.Subject.CommonName != tt.cfg.CommonName {
					t.Errorf("Certificate CommonName = %v, want %v",
						certPair.Certificate.Subject.CommonName, tt.cfg.CommonName)
				}
				if certPair.Certificate.IsCA {
					t.Error("Signed certificate IsCA should be false")
				}

				// 验证证书是由 CA 签发的
				if err := certPair.Certificate.CheckSignatureFrom(ca.Certificate); err != nil {
					t.Errorf("Certificate signature verification failed: %v", err)
				}

				// 验证 SAN
				if len(tt.cfg.AltNames.DNSNames) > 0 {
					if len(certPair.Certificate.DNSNames) != len(tt.cfg.AltNames.DNSNames) {
						t.Errorf("Certificate DNSNames count = %d, want %d",
							len(certPair.Certificate.DNSNames), len(tt.cfg.AltNames.DNSNames))
					}
				}
			}
		})
	}
}

func TestEncodeCertPEM(t *testing.T) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	pem := EncodeCertPEM(ca.Certificate)
	if pem == nil {
		t.Error("EncodeCertPEM() returned nil")
	}

	// 尝试解析回证书
	certs, err := ParseCertsPEM(pem)
	if err != nil {
		t.Errorf("ParseCertsPEM() error = %v", err)
	}
	if len(certs) != 1 {
		t.Errorf("ParseCertsPEM() returned %d certificates, want 1", len(certs))
	}
}

func TestEncodePrivateKeyPEM(t *testing.T) {
	tests := []struct {
		name    string
		keyType KeyType
		wantErr bool
	}{
		{
			name:    "RSA key",
			keyType: KeyTypeRSA,
			wantErr: false,
		},
		{
			name:    "ECDSA key",
			keyType: KeyTypeECDSA,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := NewPrivateKey(tt.keyType)
			if err != nil {
				t.Fatalf("NewPrivateKey() error = %v", err)
			}

			pem, err := EncodePrivateKeyPEM(key)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodePrivateKeyPEM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if pem == nil {
					t.Error("EncodePrivateKeyPEM() returned nil")
				}

				// 尝试解析回私钥
				parsedKey, err := ParsePrivateKeyPEM(pem)
				if err != nil {
					t.Errorf("ParsePrivateKeyPEM() error = %v", err)
				}
				if parsedKey == nil {
					t.Error("ParsePrivateKeyPEM() returned nil")
				}
			}
		})
	}
}

func TestFileOperations(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "cert-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	certPath := filepath.Join(tmpDir, "test.crt")
	keyPath := filepath.Join(tmpDir, "test.key")

	// 创建 CA
	ca, err := NewCA(Config{
		CommonName:   "Test CA",
		Organization: []string{"Test Org"},
		ValidYears:   10,
	})
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// 测试保存 CA
	t.Run("SaveCA", func(t *testing.T) {
		err := ca.SaveToFile(certPath, keyPath)
		if err != nil {
			t.Errorf("CA.SaveToFile() error = %v", err)
		}

		// 验证文件存在
		exists, err := CertAndKeyExist(certPath, keyPath)
		if err != nil {
			t.Errorf("CertAndKeyExist() error = %v", err)
		}
		if !exists {
			t.Error("Certificate and key files should exist")
		}
	})

	// 测试加载 CA
	t.Run("LoadCA", func(t *testing.T) {
		loadedCA, err := LoadCA(certPath, keyPath)
		if err != nil {
			t.Errorf("LoadCA() error = %v", err)
			return
		}

		if loadedCA.Certificate.Subject.CommonName != ca.Certificate.Subject.CommonName {
			t.Error("Loaded CA CommonName doesn't match original")
		}
	})

	// 测试签发证书并保存
	t.Run("SignedCertAndSave", func(t *testing.T) {
		certPair, err := ca.NewSignedCert(Config{
			CommonName: "test.example.com",
			ValidYears: 1,
			AltNames: AltNames{
				DNSNames: []string{"test.example.com"},
			},
			Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		})
		if err != nil {
			t.Fatalf("CA.NewSignedCert() error = %v", err)
		}

		serverCertPath := filepath.Join(tmpDir, "server.crt")
		serverKeyPath := filepath.Join(tmpDir, "server.key")

		err = certPair.SaveToFile(serverCertPath, serverKeyPath)
		if err != nil {
			t.Errorf("CertKeyPair.SaveToFile() error = %v", err)
		}

		// 验证文件存在
		exists, err := CertAndKeyExist(serverCertPath, serverKeyPath)
		if err != nil {
			t.Errorf("CertAndKeyExist() error = %v", err)
		}
		if !exists {
			t.Error("Server certificate and key files should exist")
		}

		// 读取并验证
		loadedCert, loadedKey, err := ReadCertAndKeyFromFile(serverCertPath, serverKeyPath)
		if err != nil {
			t.Errorf("ReadCertAndKeyFromFile() error = %v", err)
			return
		}

		if loadedCert.Subject.CommonName != "test.example.com" {
			t.Error("Loaded certificate CommonName doesn't match")
		}
		if loadedKey == nil {
			t.Error("Loaded key is nil")
		}
	})
}

func TestParseCertsPEM(t *testing.T) {
	ca, err := NewCA(Config{
		CommonName: "Test CA",
	})
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	pem := EncodeCertPEM(ca.Certificate)

	tests := []struct {
		name    string
		pemData []byte
		wantErr bool
		wantLen int
	}{
		{
			name:    "Valid PEM",
			pemData: pem,
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "Empty PEM",
			pemData: []byte{},
			wantErr: true,
			wantLen: 0,
		},
		{
			name:    "Invalid PEM",
			pemData: []byte("invalid data"),
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			certs, err := ParseCertsPEM(tt.pemData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCertsPEM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(certs) != tt.wantLen {
				t.Errorf("ParseCertsPEM() returned %d certificates, want %d", len(certs), tt.wantLen)
			}
		})
	}
}

func TestNewCertPool(t *testing.T) {
	ca1, err := NewCA(Config{CommonName: "CA1"})
	if err != nil {
		t.Fatalf("Failed to create CA1: %v", err)
	}

	ca2, err := NewCA(Config{CommonName: "CA2"})
	if err != nil {
		t.Fatalf("Failed to create CA2: %v", err)
	}

	pool := NewCertPool(ca1.Certificate, ca2.Certificate)
	if pool == nil {
		t.Error("NewCertPool() returned nil")
	}
}

func TestNewCertPoolFromPEM(t *testing.T) {
	ca, err := NewCA(Config{CommonName: "Test CA"})
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	pemData := EncodeCertPEM(ca.Certificate)

	pool, err := NewCertPoolFromPEM(pemData)
	if err != nil {
		t.Errorf("NewCertPoolFromPEM() error = %v", err)
	}
	if pool == nil {
		t.Error("NewCertPoolFromPEM() returned nil")
	}
}

// 示例：创建 CA
func ExampleNewCA() {
	ca, err := NewCA(Config{
		CommonName:   "My Root CA",
		Organization: []string{"My Organization"},
		ValidYears:   10,
		KeyType:      KeyTypeRSA,
	})
	if err != nil {
		panic(err)
	}

	// 保存 CA 到文件
	_ = ca.SaveToFile("ca.crt", "ca.key")
}

// 示例：签发证书
func ExampleCA_NewSignedCert() {
	// 创建 CA
	ca, _ := NewCA(Config{
		CommonName: "My Root CA",
		ValidYears: 10,
	})

	// 签发服务器证书
	certPair, err := ca.NewSignedCert(Config{
		CommonName: "server.example.com",
		ValidYears: 1,
		AltNames: AltNames{
			DNSNames: []string{"server.example.com", "*.server.example.com"},
			IPs:      []net.IP{net.ParseIP("192.168.1.100")},
		},
		Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		panic(err)
	}

	// 保存证书到文件
	_ = certPair.SaveToFile("server.crt", "server.key")
}

// TestBase64Encoding 测试 base64 编码和解码
func TestBase64Encoding(t *testing.T) {
	// 创建测试 CA
	ca, err := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// 测试证书编码
	t.Run("EncodeCertToBase64", func(t *testing.T) {
		base64Str, err := EncodeCertToBase64(ca.Certificate)
		if err != nil {
			t.Errorf("EncodeCertToBase64() error = %v", err)
			return
		}
		if base64Str == "" {
			t.Error("EncodeCertToBase64() returned empty string")
		}

		// 解码并验证
		cert, err := ParseCertFromBase64(base64Str)
		if err != nil {
			t.Errorf("ParseCertFromBase64() error = %v", err)
			return
		}
		if cert.Subject.CommonName != ca.Certificate.Subject.CommonName {
			t.Error("Decoded certificate CommonName doesn't match")
		}
	})

	// 测试私钥编码
	t.Run("EncodePrivateKeyToBase64", func(t *testing.T) {
		base64Str, err := EncodePrivateKeyToBase64(ca.PrivateKey)
		if err != nil {
			t.Errorf("EncodePrivateKeyToBase64() error = %v", err)
			return
		}
		if base64Str == "" {
			t.Error("EncodePrivateKeyToBase64() returned empty string")
		}

		// 解码并验证
		key, err := ParsePrivateKeyFromBase64(base64Str)
		if err != nil {
			t.Errorf("ParsePrivateKeyFromBase64() error = %v", err)
			return
		}
		if key == nil {
			t.Error("ParsePrivateKeyFromBase64() returned nil key")
		}
	})

	// 测试 CA ToBase64
	t.Run("CA_ToBase64", func(t *testing.T) {
		certBase64, keyBase64, err := ca.ToBase64()
		if err != nil {
			t.Errorf("CA.ToBase64() error = %v", err)
			return
		}
		if certBase64 == "" || keyBase64 == "" {
			t.Error("CA.ToBase64() returned empty strings")
		}

		// 从 base64 加载并验证
		loadedCA, err := LoadCAFromBase64(certBase64, keyBase64)
		if err != nil {
			t.Errorf("LoadCAFromBase64() error = %v", err)
			return
		}
		if loadedCA.Certificate.Subject.CommonName != ca.Certificate.Subject.CommonName {
			t.Error("Loaded CA CommonName doesn't match")
		}
	})

	// 测试 CertKeyPair ToBase64
	t.Run("CertKeyPair_ToBase64", func(t *testing.T) {
		certPair, err := ca.NewSignedCert(Config{
			CommonName: "test.example.com",
			ValidYears: 1,
			Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		})
		if err != nil {
			t.Fatalf("Failed to create cert pair: %v", err)
		}

		certBase64, keyBase64, err := certPair.ToBase64()
		if err != nil {
			t.Errorf("CertKeyPair.ToBase64() error = %v", err)
			return
		}
		if certBase64 == "" || keyBase64 == "" {
			t.Error("CertKeyPair.ToBase64() returned empty strings")
		}

		// 从 base64 加载并验证
		loadedPair, err := LoadCertKeyPairFromBase64(certBase64, keyBase64)
		if err != nil {
			t.Errorf("LoadCertKeyPairFromBase64() error = %v", err)
			return
		}
		if loadedPair.Certificate.Subject.CommonName != certPair.Certificate.Subject.CommonName {
			t.Error("Loaded cert CommonName doesn't match")
		}
	})
}

// TestBase64Errors 测试 base64 错误处理
func TestBase64Errors(t *testing.T) {
	// 测试无效的 base64 字符串
	t.Run("InvalidBase64String", func(t *testing.T) {
		_, err := ParseCertFromBase64("invalid-base64!!!")
		if err == nil {
			t.Error("ParseCertFromBase64() should fail with invalid base64")
		}

		_, err = ParsePrivateKeyFromBase64("invalid-base64!!!")
		if err == nil {
			t.Error("ParsePrivateKeyFromBase64() should fail with invalid base64")
		}
	})

	// 测试空字符串
	t.Run("EmptyBase64String", func(t *testing.T) {
		_, err := ParseCertFromBase64("")
		if err == nil {
			t.Error("ParseCertFromBase64() should fail with empty string")
		}
	})

	// 测试 nil 证书编码
	t.Run("NilCertificate", func(t *testing.T) {
		_, err := EncodeCertToBase64(nil)
		if err == nil {
			t.Error("EncodeCertToBase64() should fail with nil certificate")
		}
	})

	// 测试 nil 私钥编码
	t.Run("NilPrivateKey", func(t *testing.T) {
		_, err := EncodePrivateKeyToBase64(nil)
		if err == nil {
			t.Error("EncodePrivateKeyToBase64() should fail with nil private key")
		}
	})
}

// 示例：使用 base64 编码
func ExampleCA_ToBase64() {
	// 创建 CA
	ca, _ := NewCA(Config{
		CommonName: "Test CA",
		ValidYears: 10,
	})

	// 编码为 base64
	certBase64, keyBase64, err := ca.ToBase64()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Certificate encoded: %v\n", len(certBase64) > 0)
	fmt.Printf("Private key encoded: %v\n", len(keyBase64) > 0)

	// 从 base64 加载
	loadedCA, err := LoadCAFromBase64(certBase64, keyBase64)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded CA: %s\n", loadedCA.Certificate.Subject.CommonName)
	// Output:
	// Certificate encoded: true
	// Private key encoded: true
	// Loaded CA: Test CA
}

// 示例：使用 base64 传输证书
func Example_base64Transfer() {
	// 服务端：创建并编码证书
	ca, _ := NewCA(Config{CommonName: "CA"})
	certPair, _ := ca.NewSignedCert(Config{
		CommonName: "server",
		ValidYears: 1,
		Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})

	certBase64, keyBase64, _ := certPair.ToBase64()

	// 传输 certBase64 和 keyBase64（通过网络、数据库等）
	// ...

	// 客户端：从 base64 解码证书
	receivedPair, _ := LoadCertKeyPairFromBase64(certBase64, keyBase64)

	fmt.Printf("Received certificate: %s\n", receivedPair.Certificate.Subject.CommonName)
	// Output:
	// Received certificate: server
}
