package cert

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"time"
)

const (
	// PrivateKeyBlockType PEM 私钥块类型
	PrivateKeyBlockType = "PRIVATE KEY"
	// PublicKeyBlockType PEM 公钥块类型
	PublicKeyBlockType = "PUBLIC KEY"
	// CertificateBlockType PEM 证书块类型
	CertificateBlockType = "CERTIFICATE"
	// RSAPrivateKeyBlockType PEM RSA 私钥块类型
	RSAPrivateKeyBlockType = "RSA PRIVATE KEY"
	// ECPrivateKeyBlockType PEM ECDSA 私钥块类型
	ECPrivateKeyBlockType = "EC PRIVATE KEY"
	// CertificateRequestBlockType PEM 证书请求块类型
	CertificateRequestBlockType = "CERTIFICATE REQUEST"

	// 默认配置
	defaultRSAKeySize = 2048
	defaultValidYears = 10
)

var (
	// ErrInvalidCertificate 无效的证书
	ErrInvalidCertificate = errors.New("invalid certificate")
	// ErrInvalidPrivateKey 无效的私钥
	ErrInvalidPrivateKey = errors.New("invalid private key")
	// ErrInvalidPublicKey 无效的公钥
	ErrInvalidPublicKey = errors.New("invalid public key")
	// ErrNoCertificateFound 未找到证书
	ErrNoCertificateFound = errors.New("no certificate found in PEM data")
	// ErrNoPrivateKeyFound 未找到私钥
	ErrNoPrivateKeyFound = errors.New("no private key found in PEM data")
)

// KeyType 密钥类型
type KeyType string

const (
	// KeyTypeRSA RSA 密钥
	KeyTypeRSA KeyType = "RSA"
	// KeyTypeECDSA ECDSA 密钥
	KeyTypeECDSA KeyType = "ECDSA"
)

// AltNames 证书的备用名称（SAN - Subject Alternative Names）
type AltNames struct {
	DNSNames []string `json:"dnsNames,omitempty" yaml:"dnsNames"`
	IPs      []net.IP `json:"ips,omitempty" yaml:"ips"`
}

// Config 证书配置
type Config struct {
	// CommonName 证书通用名称
	CommonName string `json:"commonName" yaml:"commonName"`
	// Organization 组织名称列表
	Organization []string `json:"organization,omitempty" yaml:"organization"`
	// ValidYears 证书有效期（年）
	ValidYears int `json:"validYears,omitempty" yaml:"validYears"`
	// AltNames 备用名称
	AltNames AltNames `json:"altNames,omitempty" yaml:"altNames"`
	// Usages 密钥用途
	Usages []x509.ExtKeyUsage `json:"usages,omitempty" yaml:"usages"`
	// KeyType 密钥类型
	KeyType KeyType `json:"keyType,omitempty" yaml:"keyType"`
}

// CA 表示一个证书颁发机构
type CA struct {
	Certificate *x509.Certificate
	PrivateKey  crypto.Signer
}

// CertKeyPair 表示证书和私钥对
type CertKeyPair struct {
	Certificate *x509.Certificate
	PrivateKey  crypto.Signer
}

// NewPrivateKey 生成新的私钥
func NewPrivateKey(keyType KeyType) (crypto.Signer, error) {
	switch keyType {
	case KeyTypeECDSA:
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case KeyTypeRSA, "":
		return rsa.GenerateKey(rand.Reader, defaultRSAKeySize)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// NewCA 创建新的 CA 证书和私钥
func NewCA(cfg Config) (*CA, error) {
	if cfg.CommonName == "" {
		return nil, errors.New("common name is required")
	}

	// 设置默认值
	if cfg.ValidYears == 0 {
		cfg.ValidYears = defaultValidYears
	}

	// 生成私钥
	key, err := NewPrivateKey(cfg.KeyType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// 生成 CA 证书
	cert, err := newSelfSignedCACert(key, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CA certificate: %w", err)
	}

	return &CA{
		Certificate: cert,
		PrivateKey:  key,
	}, nil
}

// newSelfSignedCACert 创建自签名 CA 证书
func newSelfSignedCACert(key crypto.Signer, cfg Config) (*x509.Certificate, error) {
	now := time.Now()
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	tmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		NotBefore:             now.UTC(),
		NotAfter:              now.AddDate(cfg.ValidYears, 0, 0).UTC(),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDERBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, key.Public(), key)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	return x509.ParseCertificate(certDERBytes)
}

// NewSignedCert 使用 CA 签发新证书
func (ca *CA) NewSignedCert(cfg Config) (*CertKeyPair, error) {
	if cfg.CommonName == "" {
		return nil, errors.New("common name is required")
	}
	if len(cfg.Usages) == 0 {
		return nil, errors.New("at least one key usage is required")
	}

	// 设置默认值
	if cfg.ValidYears == 0 {
		cfg.ValidYears = defaultValidYears
	}

	// 生成私钥
	key, err := NewPrivateKey(cfg.KeyType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// 生成证书
	cert, err := ca.signCert(key, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign certificate: %w", err)
	}

	return &CertKeyPair{
		Certificate: cert,
		PrivateKey:  key,
	}, nil
}

// signCert 使用 CA 签发证书
func (ca *CA) signCert(key crypto.Signer, cfg Config) (*x509.Certificate, error) {
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	now := time.Now()
	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:     cfg.AltNames.DNSNames,
		IPAddresses:  cfg.AltNames.IPs,
		SerialNumber: serialNumber,
		NotBefore:    now.UTC(),
		NotAfter:     now.AddDate(cfg.ValidYears, 0, 0).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
	}

	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTmpl, ca.Certificate, key.Public(), ca.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	return x509.ParseCertificate(certDERBytes)
}

// EncodeCertPEM 将证书编码为 PEM 格式
func EncodeCertPEM(cert *x509.Certificate) []byte {
	if cert == nil {
		return nil
	}
	block := pem.Block{
		Type:  CertificateBlockType,
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(&block)
}

// EncodePrivateKeyPEM 将私钥编码为 PEM 格式
func EncodePrivateKeyPEM(key crypto.Signer) ([]byte, error) {
	if key == nil {
		return nil, ErrInvalidPrivateKey
	}

	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		derBytes, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ECDSA private key: %w", err)
		}
		block := &pem.Block{
			Type:  ECPrivateKeyBlockType,
			Bytes: derBytes,
		}
		return pem.EncodeToMemory(block), nil
	case *rsa.PrivateKey:
		block := &pem.Block{
			Type:  RSAPrivateKeyBlockType,
			Bytes: x509.MarshalPKCS1PrivateKey(k),
		}
		return pem.EncodeToMemory(block), nil
	default:
		return nil, fmt.Errorf("unsupported private key type: %T", key)
	}
}

// EncodePublicKeyPEM 将公钥编码为 PEM 格式
func EncodePublicKeyPEM(key crypto.PublicKey) ([]byte, error) {
	if key == nil {
		return nil, ErrInvalidPublicKey
	}

	der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	block := pem.Block{
		Type:  PublicKeyBlockType,
		Bytes: der,
	}
	return pem.EncodeToMemory(&block), nil
}

// ParseCertsPEM 从 PEM 数据中解析证书
func ParseCertsPEM(pemData []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	for len(pemData) > 0 {
		var block *pem.Block
		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}

		// 只处理证书块
		if block.Type != CertificateBlockType || len(block.Headers) != 0 {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}

		certs = append(certs, cert)
	}

	if len(certs) == 0 {
		return nil, ErrNoCertificateFound
	}

	return certs, nil
}

// ParsePrivateKeyPEM 从 PEM 数据中解析私钥
func ParsePrivateKeyPEM(pemData []byte) (crypto.Signer, error) {
	for len(pemData) > 0 {
		var block *pem.Block
		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}

		switch block.Type {
		case ECPrivateKeyBlockType:
			key, err := x509.ParseECPrivateKey(block.Bytes)
			if err == nil {
				return key, nil
			}
		case RSAPrivateKeyBlockType:
			key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err == nil {
				return key, nil
			}
		case PrivateKeyBlockType:
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err == nil {
				// 需要类型断言为 crypto.Signer
				if signer, ok := key.(crypto.Signer); ok {
					return signer, nil
				}
			}
		}
	}

	return nil, ErrNoPrivateKeyFound
}

// ParsePublicKeyPEM 从 PEM 数据中解析公钥
func ParsePublicKeyPEM(pemData []byte) (crypto.PublicKey, error) {
	for len(pemData) > 0 {
		var block *pem.Block
		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}

		if block.Type == PublicKeyBlockType {
			key, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err == nil {
				return key, nil
			}
		}
	}

	return nil, ErrInvalidPublicKey
}

// NewCertPool 创建证书池
func NewCertPool(certs ...*x509.Certificate) *x509.CertPool {
	pool := x509.NewCertPool()
	for _, cert := range certs {
		pool.AddCert(cert)
	}
	return pool
}

// NewCertPoolFromPEM 从 PEM 数据创建证书池
func NewCertPoolFromPEM(pemData []byte) (*x509.CertPool, error) {
	certs, err := ParseCertsPEM(pemData)
	if err != nil {
		return nil, err
	}
	return NewCertPool(certs...), nil
}
