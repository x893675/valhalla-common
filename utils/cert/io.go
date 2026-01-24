package cert

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// 文件权限
	certFileMode = 0644
	keyFileMode  = 0600
	dirFileMode  = 0755
)

// WriteCertToFile 将证书写入文件
func WriteCertToFile(certPath string, cert *x509.Certificate) error {
	if cert == nil {
		return ErrInvalidCertificate
	}

	pemData := EncodeCertPEM(cert)
	if pemData == nil {
		return ErrInvalidCertificate
	}

	return writeFile(certPath, pemData, certFileMode)
}

// WritePrivateKeyToFile 将私钥写入文件
func WritePrivateKeyToFile(keyPath string, key crypto.Signer) error {
	if key == nil {
		return ErrInvalidPrivateKey
	}

	pemData, err := EncodePrivateKeyPEM(key)
	if err != nil {
		return err
	}

	return writeFile(keyPath, pemData, keyFileMode)
}

// WritePublicKeyToFile 将公钥写入文件
func WritePublicKeyToFile(keyPath string, key crypto.PublicKey) error {
	if key == nil {
		return ErrInvalidPublicKey
	}

	pemData, err := EncodePublicKeyPEM(key)
	if err != nil {
		return err
	}

	return writeFile(keyPath, pemData, certFileMode)
}

// WriteCertAndKeyToFile 将证书和私钥写入文件
// certPath: 证书文件路径
// keyPath: 私钥文件路径
func WriteCertAndKeyToFile(certPath, keyPath string, cert *x509.Certificate, key crypto.Signer) error {
	if err := WritePrivateKeyToFile(keyPath, key); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	if err := WriteCertToFile(certPath, cert); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	return nil
}

// ReadCertFromFile 从文件读取证书
func ReadCertFromFile(certPath string) (*x509.Certificate, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	certs, err := ParseCertsPEM(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	if len(certs) == 0 {
		return nil, ErrNoCertificateFound
	}

	return certs[0], nil
}

// ReadCertsFromFile 从文件读取多个证书
func ReadCertsFromFile(certPath string) ([]*x509.Certificate, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	certs, err := ParseCertsPEM(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificates: %w", err)
	}

	return certs, nil
}

// ReadPrivateKeyFromFile 从文件读取私钥
func ReadPrivateKeyFromFile(keyPath string) (crypto.Signer, error) {
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	key, err := ParsePrivateKeyPEM(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
}

// ReadPublicKeyFromFile 从文件读取公钥
func ReadPublicKeyFromFile(keyPath string) (crypto.PublicKey, error) {
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	key, err := ParsePublicKeyPEM(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return key, nil
}

// ReadCertAndKeyFromFile 从文件读取证书和私钥
func ReadCertAndKeyFromFile(certPath, keyPath string) (*x509.Certificate, crypto.Signer, error) {
	cert, err := ReadCertFromFile(certPath)
	if err != nil {
		return nil, nil, err
	}

	key, err := ReadPrivateKeyFromFile(keyPath)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

// LoadCA 从文件加载 CA
func LoadCA(certPath, keyPath string) (*CA, error) {
	cert, key, err := ReadCertAndKeyFromFile(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	return &CA{
		Certificate: cert,
		PrivateKey:  key,
	}, nil
}

// SaveCA 保存 CA 到文件
func (ca *CA) SaveToFile(certPath, keyPath string) error {
	return WriteCertAndKeyToFile(certPath, keyPath, ca.Certificate, ca.PrivateKey)
}

// SaveCertKeyPair 保存证书和私钥对到文件
func (ckp *CertKeyPair) SaveToFile(certPath, keyPath string) error {
	return WriteCertAndKeyToFile(certPath, keyPath, ckp.Certificate, ckp.PrivateKey)
}

// CertAndKeyExist 检查证书和私钥文件是否都存在
func CertAndKeyExist(certPath, keyPath string) (bool, error) {
	certExists := fileExists(certPath)
	keyExists := fileExists(keyPath)

	if !certExists && !keyExists {
		return false, nil
	}

	if !certExists {
		return false, fmt.Errorf("certificate file not found: %s", certPath)
	}

	if !keyExists {
		return false, fmt.Errorf("private key file not found: %s", keyPath)
	}

	return true, nil
}

// writeFile 写入文件（自动创建目录）
func writeFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirFileMode); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// fileExists 检查文件是否存在且可读
func fileExists(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// Base64 编码相关函数

// EncodeCertToBase64 将证书编码为 base64 字符串
func EncodeCertToBase64(cert *x509.Certificate) (string, error) {
	if cert == nil {
		return "", ErrInvalidCertificate
	}

	pemData := EncodeCertPEM(cert)
	if pemData == nil {
		return "", ErrInvalidCertificate
	}

	return base64.StdEncoding.EncodeToString(pemData), nil
}

// EncodePrivateKeyToBase64 将私钥编码为 base64 字符串
func EncodePrivateKeyToBase64(key crypto.Signer) (string, error) {
	if key == nil {
		return "", ErrInvalidPrivateKey
	}

	pemData, err := EncodePrivateKeyPEM(key)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(pemData), nil
}

// ParseCertFromBase64 从 base64 字符串解析证书
func ParseCertFromBase64(base64Str string) (*x509.Certificate, error) {
	pemData, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	certs, err := ParseCertsPEM(pemData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	if len(certs) == 0 {
		return nil, ErrNoCertificateFound
	}

	return certs[0], nil
}

// ParsePrivateKeyFromBase64 从 base64 字符串解析私钥
func ParsePrivateKeyFromBase64(base64Str string) (crypto.Signer, error) {
	pemData, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	key, err := ParsePrivateKeyPEM(pemData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
}

// ToBase64 将 CA 编码为 base64 字符串（返回证书和私钥的 base64）
func (ca *CA) ToBase64() (certBase64, keyBase64 string, err error) {
	certBase64, err = EncodeCertToBase64(ca.Certificate)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode certificate: %w", err)
	}

	keyBase64, err = EncodePrivateKeyToBase64(ca.PrivateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode private key: %w", err)
	}

	return certBase64, keyBase64, nil
}

// LoadCAFromBase64 从 base64 字符串加载 CA
func LoadCAFromBase64(certBase64, keyBase64 string) (*CA, error) {
	cert, err := ParseCertFromBase64(certBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	key, err := ParsePrivateKeyFromBase64(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &CA{
		Certificate: cert,
		PrivateKey:  key,
	}, nil
}

// ToBase64 将证书和私钥对编码为 base64 字符串
func (ckp *CertKeyPair) ToBase64() (certBase64, keyBase64 string, err error) {
	certBase64, err = EncodeCertToBase64(ckp.Certificate)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode certificate: %w", err)
	}

	keyBase64, err = EncodePrivateKeyToBase64(ckp.PrivateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode private key: %w", err)
	}

	return certBase64, keyBase64, nil
}

// LoadCertKeyPairFromBase64 从 base64 字符串加载证书和私钥对
func LoadCertKeyPairFromBase64(certBase64, keyBase64 string) (*CertKeyPair, error) {
	cert, err := ParseCertFromBase64(certBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	key, err := ParsePrivateKeyFromBase64(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &CertKeyPair{
		Certificate: cert,
		PrivateKey:  key,
	}, nil
}
