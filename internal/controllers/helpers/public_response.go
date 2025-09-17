package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/sanitize"
	"github.com/CrowdShield/go-core/lib/std_errors"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/tools/maps"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

// StandardPublicRequestWrapper wraps data in simple JSON responses  that filters public data
func StandardPublicRequestWrapper[T any](fn func(res http.ResponseWriter, req *http.Request) (T, int, error)) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		result, code, err := fn(res, req)
		if err != nil {
			var pubError *std_errors.PublicError
			if errors.As(err, &pubError) {
				ErrorWrapper(res, req, pubError.Public(), code)
				return
			}

			ErrorWrapper(res, req, err.Error(), code)
			return
		}

		PublicJSONDataResponseWrapper(res, req, result)
	})
}

// PublicJSONDataResponseWrapper wrapper to convert responses to JSONData api response that filters public data
func PublicJSONDataResponseWrapper[T any](res http.ResponseWriter, req *http.Request, data T) {
	rawSanitizedResponse := sanitize.PublicSanitizeResponse(data)

	sanitizedResponse, err := maps.RecursiveJSON(rawSanitizedResponse)
	if err != nil {
		log.Error(errors.WithStack(err))
		ErrorWrapper(res, req, "Internal Error", http.StatusInternalServerError)
		return
	}

	if !tools.Empty(environment.GetConfig().Server.ResponseEncryptKey) {
		jsonBytes, err := json.Marshal(sanitizedResponse)
		if err != nil {
			log.Error(errors.WithStack(err))
			ErrorWrapper(res, req, "Internal Error", http.StatusInternalServerError)
			return
		}
		key, err := hex.DecodeString(environment.GetConfig().Server.ResponseEncryptKey)
		if err != nil {
			log.Error(errors.WithStack(err))
			ErrorWrapper(res, req, "Internal Error", http.StatusInternalServerError)
			return
		}
		encryptedData, err := encryptPublicResponse(key, jsonBytes)
		if err != nil {
			log.Error(errors.WithStack(err))
			ErrorWrapper(res, req, "Internal Error", http.StatusInternalServerError)
			return
		}
		router.JSONDataResponse(req.Context(), res, encryptedData)
		return
	}
	router.JSONDataResponse(req.Context(), res, sanitizedResponse)
}

// Encrypt data with AES-GCM
func encryptPublicResponse(encryptionKey []byte, data []byte) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", errors.WithStack(err)
	}

	nonce := make([]byte, 12) // 12 bytes nonce for GCM
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.WithStack(err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.WithStack(err)
	}

	ciphertext := aesGCM.Seal(nil, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}

// GenerateKey generates a secure random key of the specified length (in bytes).
// Typical lengths: 16 (AES-128), 24 (AES-192), 32 (AES-256).
func GenerateKey(length int) (string, error) {
	if length != 16 && length != 24 && length != 32 {
		return "", errors.New("invalid key length: must be 16, 24, or 32 bytes")
	}

	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate key: %v", err)
	}

	// Convert the key to a hexadecimal string
	return hex.EncodeToString(key), nil
}
