package auth

import (
	"bytes"
	"encoding/json"
	"fmt"

	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/gomail.v2"
)

type KeycloakConfig struct {
	URL          string
	Realm        string
	ClientID     string
	ClientSecret string
}

type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	From     string
}

type JWTConfig struct {
	Secret            string
	ExpirationMinutes int
}

type AuthService struct {
	Keycloak KeycloakConfig
	Email    EmailConfig
	JWT      JWTConfig
	Frontend string // base URL for links
}

func NewAuthService(kc KeycloakConfig, email EmailConfig, jwtCfg JWTConfig, frontend string) *AuthService {
	return &AuthService{
		Keycloak: kc,
		Email:    email,
		JWT:      jwtCfg,
		Frontend: frontend,
	}
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// ====================== REGISTER ======================
func (s *AuthService) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	adminToken, err := s.getAdminToken()
	if err != nil {
		http.Error(w, "Failed to get admin token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Создаём пользователя disabled + emailVerified=false
	payload := map[string]interface{}{
		"username":      user.Username,
		"email":         user.Email,
		"enabled":       false,
		"emailVerified": false,
		"credentials": []map[string]interface{}{
			{"type": "password", "value": user.Password, "temporary": false},
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST",
		fmt.Sprintf("%s/admin/realms/%s/users", s.Keycloak.URL, s.Keycloak.Realm),
		bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		http.Error(w, "Failed to create user: "+string(body), resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	// Получаем userID
	location := resp.Header.Get("Location")
	userID := location[strings.LastIndex(location, "/")+1:]

	// Генерируем JWT для подтверждения email
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(time.Minute * time.Duration(s.JWT.ExpirationMinutes)).Unix(),
	})
	tokenStr, _ := token.SignedString([]byte(s.JWT.Secret))

	verifyLink := fmt.Sprintf("%s/verify-email?token=%s", s.Frontend, tokenStr)
	if err := sendVerificationEmail(s.Email, user.Email, verifyLink); err != nil {
		http.Error(w, "Failed to send verification email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered. Please check your email to verify the account.",
	})
}

// ====================== VERIFY EMAIL ======================
func (s *AuthService) VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "Token required", http.StatusBadRequest)
		return
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusBadRequest)
		return
	}

	userID := fmt.Sprintf("%v", claims["userID"])
	adminToken, _ := s.getAdminToken()

	// Включаем пользователя и отмечаем emailVerified=true
	payload := map[string]interface{}{"enabled": true, "emailVerified": true}
	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT",
		fmt.Sprintf("%s/admin/realms/%s/users/%s", s.Keycloak.URL, s.Keycloak.Realm, userID),
		bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 204 {
		body, _ := ioutil.ReadAll(resp.Body)
		http.Error(w, "Failed to verify user: "+string(body), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Email verified successfully! You can now log in."))
}

// ====================== RESEND VERIFICATION ======================
func (s *AuthService) ResendVerificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || reqBody.Username == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	adminToken, err := s.getAdminToken()
	if err != nil {
		http.Error(w, "Failed to get admin token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем пользователя по username
	getUserURL := fmt.Sprintf("%s/admin/realms/%s/users?username=%s", s.Keycloak.URL, s.Keycloak.Realm, reqBody.Username)
	req, _ := http.NewRequest("GET", getUserURL, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to query user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	var users []map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&users); err != nil || len(users) == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	userID := fmt.Sprintf("%v", users[0]["id"])
	email := fmt.Sprintf("%v", users[0]["email"])

	// Генерируем новый JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(time.Minute * time.Duration(s.JWT.ExpirationMinutes)).Unix(),
	})
	tokenStr, _ := token.SignedString([]byte(s.JWT.Secret))
	verifyLink := fmt.Sprintf("%s/verify-email?token=%s", s.Frontend, tokenStr)

	if err := sendVerificationEmail(s.Email, email, verifyLink); err != nil {
		http.Error(w, "Failed to send verification email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Verification email has been sent",
	})
}

// ====================== LOGIN ======================
func (s *AuthService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", s.Keycloak.ClientID)
	data.Set("client_secret", s.Keycloak.ClientSecret)
	data.Set("username", creds.Username)
	data.Set("password", creds.Password)
	data.Set("scope", "openid")

	resp, err := http.PostForm(fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", s.Keycloak.URL, s.Keycloak.Realm), data)
	if err != nil {
		http.Error(w, "Keycloak error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		http.Error(w, string(body), resp.StatusCode)
		return
	}

	var tokens TokenResponse
	if err := json.Unmarshal(body, &tokens); err != nil {
		http.Error(w, "Failed to parse token response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// ====================== HELPERS ======================
func (s *AuthService) getAdminToken() (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", s.Keycloak.ClientID)
	data.Set("client_secret", s.Keycloak.ClientSecret)

	resp, err := http.PostForm(fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", s.Keycloak.URL, s.Keycloak.Realm), data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("admin token error: %s", string(body))
	}

	var t TokenResponse
	if err := json.Unmarshal(body, &t); err != nil {
		return "", err
	}
	return t.AccessToken, nil
}

func sendVerificationEmail(cfg EmailConfig, toEmail, link string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.From)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Verify your email address")
	m.SetBody("text/plain", fmt.Sprintf(
		"Someone has created an account with this email address. If this was you, click the link below to verify your email address:\n\n%s\n\nThis link will expire within 5 minutes.\n\nIf you didn't create this account, just ignore this message.",
		link))

	d := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	return d.DialAndSend(m)
}

// ====================== REFRESH TOKEN ======================
func (s *AuthService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", s.Keycloak.ClientID)
	data.Set("client_secret", s.Keycloak.ClientSecret)
	data.Set("refresh_token", refreshToken)

	resp, err := http.PostForm(fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", s.Keycloak.URL, s.Keycloak.Realm), data)
	if err != nil {
		return nil, fmt.Errorf("keycloak error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to refresh token: %s", string(body))
	}

	var tokens TokenResponse
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokens, nil
}
