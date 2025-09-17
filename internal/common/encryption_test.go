package common_test

import (
	"testing"

	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

func init() {
	system_testing.BuildSystem()
}

func TestEncryption(t *testing.T) {
	if tools.Empty(environment.GetConfig().Encryption) || tools.Empty(environment.GetConfig().Encryption.CurrentCipher) {
		t.Skip("Skipping test because encryption key is not set")
	}

	encryptedString := common.EncryptedString("im some test key")
	{
		decrypted, err := encryptedString.Decrypt()
		if err != nil {
			t.Fatal(err)
		}
		if decrypted != "im some test key" {
			t.Fatal("decrypted string does not match")
		}
	}

	encrypted, err := encryptedString.Encrypt()
	if err != nil {
		t.Fatal(err)
	}

	{
		encrypted2, err := common.EncryptedString(encrypted).Encrypt()
		if err != nil {
			t.Fatal(err)
		}
		if encrypted != encrypted2 {
			t.Fatal("encrypted strings do not match")
		}
	}

	decrypted, err := common.EncryptedString(encrypted).Decrypt()
	if err != nil {
		t.Fatal(err)
	}

	if decrypted != "im some test key" {
		t.Fatal("decrypted string does not match")
	}
}

func TestGenerateAESKey(_ *testing.T) {
	// fmt.Println(common.GenerateAESKey())
}
