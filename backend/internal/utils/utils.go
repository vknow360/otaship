package utils

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type contextKey string

const projectIDKey contextKey = "project_id"

func SetProjectId(ctx context.Context, id pgtype.UUID) context.Context {
	ctx = context.WithValue(ctx, projectIDKey, id)
	return ctx
}

func GetProjectId(ctx context.Context) pgtype.UUID {
	val := ctx.Value(projectIDKey)
	if val == nil {
		return pgtype.UUID{}
	}
	id, ok := val.(pgtype.UUID)
	if !ok {
		return pgtype.UUID{}
	}
	return id
}

func ParseUUID(uuidStr string) (pgtype.UUID, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return pgtype.UUID{}, err
	}

	id := pgtype.UUID{
		Bytes: parsedUUID,
		Valid: true,
	}
	return id, nil
}

func ParseInt32(s string) (int32, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}

func CalculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func getClientIP(r *http.Request) string {
	ip := strings.TrimSpace(r.RemoteAddr)
	if ip == "" {
		return "unknown"
	}

	host, _, err := net.SplitHostPort(ip)
	if err == nil && host != "" {
		return host
	}

	return ip
}

func BuildDeviceHash(r *http.Request, platform string) string {
	fingerprint := getClientIP(r) + "|" + platform
	return CalculateSHA256([]byte(fingerprint))
}

func GenerateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func SignManifest(manifest []byte, privateKeyString string) (string, error) {
	privateKey, _ := pem.Decode([]byte(privateKeyString))
	if privateKey == nil {
		return "", errors.New("invalid private key")
	}

	var rsaKey *rsa.PrivateKey
	key, err := x509.ParsePKCS8PrivateKey(privateKey.Bytes)
	if err != nil {
		rsaKey, err = x509.ParsePKCS1PrivateKey(privateKey.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse key (PKCS8/PKCS1): %v", err)
		}
	} else {
		var ok bool
		rsaKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return "", errors.New("parsed key is not an RSA private key")
		}
	}

	hashed := sha256.Sum256(manifest)

	signatureBytes, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signatureBytes), nil
}
