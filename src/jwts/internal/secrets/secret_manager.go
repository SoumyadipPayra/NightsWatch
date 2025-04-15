package secrets

import (
	"os"
)

var secretKey []byte

func Initialize() error {
	//generally kept in secure vaults or buckets
	secretBytes, err := os.ReadFile("resources/secrets.key")
	if err != nil {
		return err
	}
	secretKey = secretBytes
	return nil
}

func GetSecretKey() []byte {
	return secretKey
}
