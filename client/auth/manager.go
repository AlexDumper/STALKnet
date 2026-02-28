package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AuthManager управляет авторизацией через HTTP-клиент к Auth Service
type AuthManager struct {
	authServiceURL string
	httpClient     *http.Client
	sessionID      string
	accessToken    string
	refreshToken   string
	username       string
	userID         int
	expiresAt      time.Time
}

// AuthState состояние авторизации
type AuthState int

const (
	StateGuest AuthState = iota
	StateEnteringName
	StateConfirmCreate
	StateEnteringPassword
	StateAuthorized
)

func (s AuthState) String() string {
	switch s {
	case StateGuest:
		return "Guest"
	case StateEnteringName:
		return "EnteringName"
	case StateConfirmCreate:
		return "ConfirmCreate"
	case StateEnteringPassword:
		return "EnteringPassword"
	case StateAuthorized:
		return "Authorized"
	default:
		return "Unknown"
	}
}

// RegisterRequest запрос регистрации
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
}

// LoginRequest запрос входа
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// TokenResponse ответ с токенами
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	SessionID    string `json:"session_id"`
}

// ValidateResponse ответ валидации
type ValidateResponse struct {
	Valid    bool   `json:"valid"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

// NewAuthManager создаёт новый менеджер авторизации
func NewAuthManager(authServiceURL string) *AuthManager {
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8081"
	}
	return &AuthManager{
		authServiceURL: authServiceURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Register регистрирует пользователя
func (am *AuthManager) Register(ctx context.Context, username, password string) error {
	req := RegisterRequest{
		Username: username,
		Password: password,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := am.httpClient.Post(
		fmt.Sprintf("%s/api/auth/register", am.authServiceURL),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: %s", string(respBody))
	}

	return nil
}

// Login авторизует пользователя
func (am *AuthManager) Login(ctx context.Context, username, password string) (*TokenResponse, error) {
	req := LoginRequest{
		Username: username,
		Password: password,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := am.httpClient.Post(
		fmt.Sprintf("%s/api/auth/login", am.authServiceURL),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed: %s", string(respBody))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	// Сохраняем сессию
	am.accessToken = tokenResp.AccessToken
	am.refreshToken = tokenResp.RefreshToken
	am.sessionID = tokenResp.SessionID
	am.username = tokenResp.Username
	am.userID = tokenResp.UserID
	am.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &tokenResp, nil
}

// Logout завершает сессию
func (am *AuthManager) Logout(ctx context.Context) error {
	if am.accessToken == "" {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/api/auth/logout", am.authServiceURL), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+am.accessToken)

	resp, err := am.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Очищаем сессию локально
	am.accessToken = ""
	am.refreshToken = ""
	am.sessionID = ""
	am.username = ""
	am.userID = 0

	return nil
}

// Validate проверяет валидность токена
func (am *AuthManager) Validate(ctx context.Context) (*ValidateResponse, error) {
	if am.accessToken == "" {
		return &ValidateResponse{Valid: false}, nil
	}

	req := ValidateRequest{Token: am.accessToken}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := am.httpClient.Post(
		fmt.Sprintf("%s/api/auth/validate", am.authServiceURL),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var validateResp ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
		return nil, err
	}

	return &validateResp, nil
}

// Refresh обновляет токен
func (am *AuthManager) Refresh(ctx context.Context) (*TokenResponse, error) {
	if am.refreshToken == "" {
		return nil, fmt.Errorf("no refresh token")
	}

	reqBody := map[string]string{"refresh_token": am.refreshToken}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := am.httpClient.Post(
		fmt.Sprintf("%s/api/auth/refresh", am.authServiceURL),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("refresh failed: %s", string(respBody))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	// Обновляем сессию
	am.accessToken = tokenResp.AccessToken
	am.refreshToken = tokenResp.RefreshToken
	am.sessionID = tokenResp.SessionID
	am.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &tokenResp, nil
}

// GetSessionID возвращает ID текущей сессии
func (am *AuthManager) GetSessionID() string {
	return am.sessionID
}

// GetUsername возвращает имя пользователя
func (am *AuthManager) GetUsername() string {
	return am.username
}

// GetUserID возвращает ID пользователя
func (am *AuthManager) GetUserID() int {
	return am.userID
}

// GetAccessToken возвращает access токен
func (am *AuthManager) GetAccessToken() string {
	return am.accessToken
}

// IsAuthorized проверяет, авторизован ли пользователь
func (am *AuthManager) IsAuthorized() bool {
	return am.accessToken != "" && time.Now().Before(am.expiresAt)
}

// ValidateRequest для валидации токена
type ValidateRequest struct {
	Token string `json:"token"`
}
