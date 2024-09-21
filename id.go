package ldb

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

func GenerateId() string {
	// MYSQL: CONCAT(UNHEX(CONV(ROUND(UNIX_TIMESTAMP(CURTIME(4))*1000), 10, 16)), RANDOM_BYTES(10))

	timestamp := int64(time.Now().UnixMilli() * 1000)

	entropy := make([]byte, 8)
	rand.Read(entropy)

	return fmt.Sprintf("%x%x", timestamp, entropy)
}

func ValidateId(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid id, expected string value")
	}

	const requiredLen = 31
	if len(str) != requiredLen {
		return fmt.Errorf("invalid id, expected string of length %v", requiredLen)
	}

	str = strings.ToLower(str)
	if len(strings.Trim(str, "0123456789abcdef")) != 0 {
		return fmt.Errorf("invalid id, expcted hex string")
	}

	return nil
}
