package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

type EncryptedString string

func (e *EncryptedString) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	avStr, ok := av.(*types.AttributeValueMemberS)
	if !ok || avStr.Value == "" {
		return nil
	}
	*e = EncryptedString(avStr.Value)
	return nil
}

func (e *EncryptedString) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	if e == nil {
		return &types.AttributeValueMemberNULL{Value: true}, nil
	}
	str, err := e.Encrypt()
	if err != nil {
		return nil, err
	}
	return &types.AttributeValueMemberS{Value: str}, nil
}

// MarshalJSON implements the json.Marshaler interface
func (e EncryptedString) MarshalJSON() ([]byte, error) {
	if e.IsEncrypted() {
		return []byte(fmt.Sprintf("\"%s\"", e)), nil
	}
	str, err := e.Encrypt()
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("\"%s\"", str)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (e *EncryptedString) UnmarshalJSON(data []byte) error {
	if len(data) < 2 {
		return nil
	}
	str := string(data[1 : len(data)-1])
	*e = EncryptedString(str)
	return nil
}

// String implements the fmt.Stringer interface
func (e EncryptedString) String() string {
	if e.IsEncrypted() {
		return string(e)
	}
	str, err := e.Encrypt()
	if err != nil {
		return string(e)
	}
	return str
}

func (e EncryptedString) IsEncrypted() bool {
	return len(e) > 0 && e[2] == ':'
}

func (e EncryptedString) Encrypt() (string, error) {
	if e.IsEncrypted() {
		return string(e), nil
	}
	cipherKey := environment.GetCurrentCipher()
	ciphers := environment.GetCiphers()
	key, err := base64.StdEncoding.DecodeString(ciphers[cipherKey])
	if err != nil {
		return string(e), err
	}
	str, err := encryptAES(key, string(e))
	if err != nil {
		return string(e), err
	}
	return fmt.Sprintf("%s:%s", cipherKey, str), nil
}

func (e EncryptedString) Decrypt() (string, error) {
	if !e.IsEncrypted() {
		return string(e), nil
	}
	encryptedCipher := e[:2]
	ciphers := environment.GetCiphers()
	cipher, ok := ciphers[string(encryptedCipher)]
	if !ok {
		return "", fmt.Errorf("invalid cipher")
	}
	key, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		return string(e), err
	}

	return decryptAES(key, string(e[3:]))
}

// encryptAES encrypts plaintext using AES with the provided key
func encryptAES(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.WithStack(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", errors.WithStack(err)
	}
	// TODO update this
	// nolint: staticcheck
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptAES decrypts ciphertext using AES with the provided key
func decryptAES(key []byte, encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", errors.WithStack(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.WithStack(err)
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	// TODO update this
	// nolint: staticcheck
	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext), nil
}

func GenerateAESKey() (string, error) {
	key := make([]byte, 32) // 32 bytes for AES-256
	_, err := rand.Read(key)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
