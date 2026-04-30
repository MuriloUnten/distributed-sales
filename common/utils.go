package common

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
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

/**
 * Validates a signature based on a slice of possible public keys.
 * SHA256 only
 */
func ValidateSignature(signature string, payload []byte, keys []*rsa.PublicKey) bool {
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}

	hashed := sha256.Sum256(payload)
	for _, key := range keys {
		err := rsa.VerifyPKCS1v15(key, crypto.SHA256, hashed[:], sig)
		if err != nil {
			continue
		}
		return true
	}

	return false
}

func Sign(priv *rsa.PrivateKey, hash crypto.Hash, payload []byte) (string, error) {
	hashed := sha256.Sum256(payload)
	signature, err := rsa.SignPKCS1v15(nil, priv, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(signature)
	return encoded, err
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
