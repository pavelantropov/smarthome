package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/argon2"
)

type PasswordService struct {
	saltLen uint32
	keyLen  uint32
	time    uint32
	memory  uint32
	threads uint8
}

func NewPasswordService() *PasswordService {
	return &PasswordService{
		saltLen: 16,
		keyLen:  32,
		time:    4,
		memory:  64 * 1024,
		threads: 4,
	}
}

func (p *PasswordService) HashPassword(password string) (string, error) {
	salt := make([]byte, p.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		p.time,
		p.memory,
		p.threads,
		p.keyLen,
	)

	// Format: $argon2id$v=19$m=65536,t=4,p=4$salt$hash
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		p.memory,
		p.time,
		p.threads,
		hex.EncodeToString(salt),
		hex.EncodeToString(hash),
	), nil
}

func (p *PasswordService) VerifyPassword(password, hash string) (bool, error) {
	// Parse hash format
	// Simplified for brevity - in production use proper parsing
	var saltHex, hashHex string
	var memory, time, threads int

	_, err := fmt.Sscanf(hash, "$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		&memory, &time, &threads, &saltHex, &hashHex)
	if err != nil {
		return false, err
	}

	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false, err
	}

	hashBytes, err := hex.DecodeString(hashHex)
	if err != nil {
		return false, err
	}

	newHash := argon2.IDKey(
		[]byte(password),
		salt,
		uint32(time),
		uint32(memory),
		uint8(threads),
		uint32(len(hashBytes)),
	)

	return subtle.ConstantTimeCompare(hashBytes, newHash) == 1, nil
}
