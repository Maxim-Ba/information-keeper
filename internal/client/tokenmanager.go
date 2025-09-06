package client

import "sync"

type TokenManager struct {
	accessToken  string
	refreshToken string
	mu           sync.RWMutex
}

func (tm *TokenManager) SetTokens(access, refresh string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.accessToken = access
	tm.refreshToken = refresh
}

func (tm *TokenManager) GetAccessToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.accessToken
}

func (tm *TokenManager) GetRefreshToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.refreshToken
}
