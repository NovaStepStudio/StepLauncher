package mods

import (
	"crypto/sha1"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// VerifyFile comprueba el hash de un archivo ya descargado. Acepta sha1 o
// sha512 en hexadecimal (se usa el primero que venga informado). Si ambos
// vienen vacíos, el archivo se da por válido.
func VerifyFile(path, wantSHA1, wantSHA512 string) (bool, error) {
	wantSHA1 = strings.ToLower(strings.TrimSpace(wantSHA1))
	wantSHA512 = strings.ToLower(strings.TrimSpace(wantSHA512))
	if wantSHA1 == "" && wantSHA512 == "" {
		return true, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("abrir para verificar: %w", err)
	}
	defer f.Close()
	if wantSHA512 != "" {
		h := sha512.New()
		if _, err := io.Copy(h, f); err != nil {
			return false, fmt.Errorf("leer para sha512: %w", err)
		}
		return hex.EncodeToString(h.Sum(nil)) == wantSHA512, nil
	}
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, fmt.Errorf("leer para sha1: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)) == wantSHA1, nil
}
