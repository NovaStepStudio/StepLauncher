package accounts

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

var usernameRe = regexp.MustCompile(`^[A-Za-z0-9_]{3,16}$`)

func ValidUsername(name string) bool {
	return usernameRe.MatchString(name)
}

func OfflineUUID(username string) string {
	sum := md5.Sum([]byte("OfflinePlayer:" + username))
	raw := hex.EncodeToString(sum[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		raw[0:8], raw[8:12], raw[12:16], raw[16:20], raw[20:32])
}

func NormalizeAccountType(t string) AccountType {
	switch AccountType(strings.ToLower(strings.TrimSpace(t))) {
	case TypeOffline:
		return TypeOffline
	case TypeAuthLib:
		return TypeAuthLib
	}
	return ""
}

func UserTypeOf(t AccountType) string {
	return "mojang"
}

func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.Join(strings.Fields(name), " ")
	return name
}

func normalizeAccessToken(token string) string {
	token = strings.TrimSpace(token)
	if len(token) != 32 {
		return token
	}
	if _, err := hex.DecodeString(token); err != nil {
		return token
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		token[0:8], token[8:12], token[12:16], token[16:20], token[20:32])
}

func newClientToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return strings.Repeat("0", 32)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	raw := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		raw[0:8], raw[8:12], raw[12:16], raw[16:20], raw[20:32])
}

func (req CreateAccountReq) errors() []string {
	var errs []string
	typ := NormalizeAccountType(string(req.Type))
	if typ == "" {
		errs = append(errs, fmt.Sprintf("tipo de cuenta invalido: %q", req.Type))
	}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		errs = append(errs, "el nombre de jugador es obligatorio")
	} else if !ValidUsername(username) {
		errs = append(errs, "el nombre de jugador solo puede contener letras, numeros y guion bajo (3-16 caracteres)")
	}
	switch typ {
	case TypeOffline:
	case TypeAuthLib:
		url := strings.TrimSpace(req.AuthServerURL)
		if url == "" {
			errs = append(errs, "el servidor de autenticacion (URL) es obligatorio para cuentas AuthLib")
		} else if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			errs = append(errs, "el servidor de autenticacion debe ser una URL http(s) valida")
		}
	}
	return errs
}

func (req AuthlibLoginReq) errorsLogin() []string {
	var errs []string
	if strings.TrimSpace(req.AuthServerURL) == "" {
		errs = append(errs, "la URL del servidor de autenticacion es obligatoria")
	}
	if strings.TrimSpace(req.Username) == "" {
		errs = append(errs, "el email o nombre de usuario es obligatorio")
	}
	if req.Password == "" {
		errs = append(errs, "la contraseña es obligatoria")
	}
	return errs
}
