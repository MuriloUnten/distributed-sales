package common

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func LoadPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bytes)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	outputKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("given key is now rsa key")
	}

	return outputKey, nil
}

func LoadPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	outputKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("given key is now rsa key")
	}

	return outputKey, nil
}

/*
 Load all public keys with .pub extensions at given directory path (not recursive)
 */
func LoadPublicKeysFromDirectory(path string) ([]*rsa.PublicKey, error) {
	directory, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	keys := make([]*rsa.PublicKey, 0, len(directory))
	for _, entry := range directory {
		if entry.IsDir() {
			continue
		}
		isKey := strings.HasSuffix(entry.Name(), ".pem")
		if !isKey {
			continue
		}

		filePath := getRelativeFilePath(path, entry.Name())
		key, err := LoadPublicKeyFromFile(filePath)
		if err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}

	return keys, nil
}

/*
 Validates a signature based on a slice of possible public keys.
 SHA256 only
 */
func ValidateSignature(signature string, payload []byte, keys []*rsa.PublicKey) bool {
	for _, key := range keys {
		err := rsa.VerifyPKCS1v15(key, crypto.SHA256, payload, []byte(signature))
		if err != nil {
			continue
		}
		return true
	}

	return false
}

func getRelativeFilePath(directory string, fileName string) string {
	dirPath := directory
	if !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}
	
	return dirPath + fileName
}

func ClearTerminal() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}
