package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Issuer     string
}

var AppConfig *Config

// LoadOrGenerateKeys sinh hoặc tải khóa ký ES256 của IdP
func LoadOrGenerateKeys(keyDir string) error {
	keyPath := filepath.Join(keyDir, "idp-private.key")
	AppConfig = &Config{
		Issuer: "https://idp.internal",
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		// Tạo khóa mới nếu chưa tồn tại
		fmt.Println("Khóa ký IdP chưa tồn tại. Đang khởi tạo khóa ECC P-256 mới...")
		if err := os.MkdirAll(keyDir, 0700); err != nil {
			return fmt.Errorf("failed to create key directory: %w", err)
		}

		privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return fmt.Errorf("failed to generate EC key: %w", err)
		}

		// Encode PEM và ghi file
		x509Encoded, err := x509.MarshalECPrivateKey(privKey)
		if err != nil {
			return err
		}
		pemBlock := &pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: x509Encoded,
		}
		pemFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer pemFile.Close()

		if err := pem.Encode(pemFile, pemBlock); err != nil {
			return err
		}

		AppConfig.PrivateKey = privKey
		AppConfig.PublicKey = &privKey.PublicKey
		fmt.Printf("Đã tạo và lưu khóa ký thành công tại: %s\n", keyPath)
	} else {
		// Đọc khóa sẵn có từ file
		fmt.Printf("Đang tải khóa ký IdP từ: %s\n", keyPath)
		pemBytes, err := os.ReadFile(keyPath)
		if err != nil {
			return err
		}
		block, _ := pem.Decode(pemBytes)
		if block == nil || block.Type != "EC PRIVATE KEY" {
			return errors.New("invalid PEM block type, expected EC PRIVATE KEY")
		}

		privKey, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return err
		}

		AppConfig.PrivateKey = privKey
		AppConfig.PublicKey = &privKey.PublicKey
		fmt.Println("Đã tải khóa ký IdP thành công.")
	}

	return nil
}
