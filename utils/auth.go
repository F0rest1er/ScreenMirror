package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"sync"
	"time"
)

type AuthManager struct {
	Password string
	Nonces   map[string]time.Time
	Sessions map[string]bool
	Bans     map[string]time.Time
	mu       sync.Mutex
}

func NewAuthManager() *AuthManager {
	return &AuthManager{
		Password: generatePassword(),
		Nonces:   make(map[string]time.Time),
		Sessions: make(map[string]bool),
		Bans:     make(map[string]time.Time),
	}
}

func generatePassword() string {
	const chars = "abcdefghjkmnopqrstuvwxyzABCDEFGHJKMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[num.Int64()]
	}
	return string(b)
}

func (a *AuthManager) GetChallenge() string {
	b := make([]byte, 16)
	rand.Read(b)
	nonce := hex.EncodeToString(b)

	a.mu.Lock()
	a.Nonces[nonce] = time.Now().Add(5 * time.Minute)
	a.mu.Unlock()
	
	return nonce
}

func (a *AuthManager) IsBanned(ip string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if expiry, ok := a.Bans[ip]; ok {
		if time.Now().Before(expiry) {
			return true
		}
		delete(a.Bans, ip)
	}
	return false
}

func (a *AuthManager) Verify(ip, nonce, clientHash string) bool {
	if a.IsBanned(ip) {
		return false
	}

	a.mu.Lock()
	_, validNonce := a.Nonces[nonce]
	if validNonce {
		delete(a.Nonces, nonce)
	}
	a.mu.Unlock()

	if !validNonce {
		a.mu.Lock()
		a.Bans[ip] = time.Now().Add(5 * time.Second)
		a.mu.Unlock()
		return false
	}

	hash := sha256.New()
	hash.Write([]byte(a.Password + nonce))
	expectedHash := hex.EncodeToString(hash.Sum(nil))

	if expectedHash == clientHash {
		return true
	}

	a.mu.Lock()
	a.Bans[ip] = time.Now().Add(5 * time.Second)
	a.mu.Unlock()

	return false
}

func (a *AuthManager) CreateSession() string {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)

	a.mu.Lock()
	a.Sessions[token] = true
	a.mu.Unlock()
	return token
}

func (a *AuthManager) IsValidSession(token string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Sessions[token]
}
