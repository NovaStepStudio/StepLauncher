package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

func VerifySHA1(path, expected string) (bool, error) {
	if expected == "" {
		return true, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	return hex.EncodeToString(h.Sum(nil)) == expected, nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}