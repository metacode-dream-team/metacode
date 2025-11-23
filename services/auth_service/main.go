package main

import (
	"auth_service/auth"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

// Middleware для CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		frontendURL := viper.GetString("server.frontend_url")

		w.Header().Set("Access-Control-Allow-Origin", frontendURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Preflight запрос
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Инициализация конфигурации
	viper.SetConfigFile("config.yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Конфиги
	kcConfig := auth.KeycloakConfig{
		URL:          viper.GetString("keycloak.url"),
		Realm:        viper.GetString("keycloak.realm"),
		ClientID:     viper.GetString("keycloak.client_id"),
		ClientSecret: viper.GetString("keycloak.client_secret"),
	}

	emailConfig := auth.EmailConfig{
		SMTPHost: viper.GetString("email.smtp_host"),
		SMTPPort: viper.GetInt("email.smtp_port"),
		SMTPUser: viper.GetString("email.smtp_user"),
		SMTPPass: viper.GetString("email.smtp_pass"),
		From:     viper.GetString("email.from"),
	}

	jwtConfig := auth.JWTConfig{
		Secret:            viper.GetString("jwt.secret"),
		ExpirationMinutes: viper.GetInt("jwt.expiration_minutes"),
	}

	frontendURL := viper.GetString("server.frontend_url")

	authService := auth.NewAuthService(kcConfig, emailConfig, jwtConfig, frontendURL)

	// Создаём mux
	mux := http.NewServeMux()
	mux.HandleFunc("/login", authService.LoginHandler)
	mux.HandleFunc("/register", authService.RegisterHandler)
	mux.HandleFunc("/verify-email", authService.VerifyEmailHandler)
	mux.HandleFunc("/resend-verification-email", authService.ResendVerificationHandler)
	mux.HandleFunc("/token/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			RefreshToken string `json:"refresh_token"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		newTokens, err := authService.RefreshToken(body.RefreshToken)
		if err != nil {
			http.Error(w, "Failed to refresh token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newTokens)
	})

	port := viper.GetInt("server.port")
	fmt.Printf("Server running on port %d\n", port)

	// Оборачиваем mux в CORS middleware
	handler := corsMiddleware(mux)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
}
