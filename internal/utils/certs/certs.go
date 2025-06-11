package certs

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func CheckCertsFolder(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path, 0700); err != nil {
				return fmt.Errorf("failed to create certs folder %q: %w", path, err)
			}
			return nil
		}
		return fmt.Errorf("failed to get stats for certs folder %q: %w", path, err)
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to get stats for certs folder %q: %w", path, err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("path %q exists but is not a directory", path)
	}

	return nil
}

func CreateCertsFiles(certsDir string) error {
	keyPath := filepath.Join(certsDir, "key.pem")
	certPath := filepath.Join(certsDir, "cert.pem")

	cmd := exec.Command(
		"openssl",
		"req",
		"-x509",
		"-newkey", "rsa:4096",
		"-keyout", keyPath,
		"-out", certPath,
		"-days", "365",
		"-nodes",
		"-subj", "/CN=localhost",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create certificate files: %w, output: %s", err, string(output))
	}

	return nil
}

func CheckCertificate(certsDir string) (bool, error) {
	certPath := filepath.Join(certsDir, "cert.pem")

	_, err := os.Stat(certPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, err

		}
		if err := CreateCertsFiles(certsDir); err != nil {
			return false, err
		}
	}

	certData, err := os.ReadFile(certPath)
	if err != nil {
		return false, fmt.Errorf("failed to read certificate file %q: %w", certPath, err)
	}

	block, _ := pem.Decode(certData)
	if block == nil || block.Type != "CERTIFICATE" {
		return false, fmt.Errorf("failed to decode PEM block from %q", certPath)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %w", err)
	}

	now := time.Now()
	if now.After(cert.NotAfter) {
		if err := os.RemoveAll(certsDir); err != nil {
			return true, fmt.Errorf("failed to remove certs folder %q: %w", certsDir, err)
		}
		return true, nil
	}

	nearExpiration := now.Add(7 * 24 * time.Hour)
	if nearExpiration.After(cert.NotAfter) {
		return true, nil
	}

	return false, nil
}
