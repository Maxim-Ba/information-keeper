package client

import "sync"

// TokenManager обеспечивает потокобезопасное управление токенами доступа.
// Использует мьютекс для синхронизации доступа к токенам из нескольких горутин.
type TokenManager struct {
	accessToken  string
	refreshToken string
	mu           sync.RWMutex
}

// SetTokens устанавливает новые значения access и refresh токенов.
// Операция является потокобезопасной.
func (tm *TokenManager) SetTokens(access, refresh string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.accessToken = access
	tm.refreshToken = refresh
}

// GetAccessToken возвращает текущий access токен.
// Операция является потокобезопасной и использует read-lock для эффективности.
func (tm *TokenManager) GetAccessToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.accessToken
}

// GetRefreshToken возвращает текущий refresh токен.
// Операция является потокобезопасной и использует read-lock для эффективности.
func (tm *TokenManager) GetRefreshToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.refreshToken
}
