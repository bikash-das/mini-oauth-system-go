package utils

import (
	"crypto/rand"
	"math/big"
)

func GenerateRandomCode(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	max := big.NewInt(int64(len(charset)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		idx := int(n.Int64())
		b[i] = charset[idx]
	}
	return string(b), nil
}
