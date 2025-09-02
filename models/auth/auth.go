package auth

import (
	"birdsfoot/app/models/account"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// OAuthRequest represents OAuth callback data
type OAuthRequest struct {
	Code     string `json:"code"`
	Provider string `json:"provider"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	User         account.User `json:"user"`
	ExpiresIn    time.Time    `json:"expiresIn"`
	AuthLevel    AuthLevel    `json:"authLevel"`
}
type AuthLevel int

const (
	NoAuthLevel AuthLevel = iota
	LoginLevel
	StepUpLevel
	ClubHouseLevel
	AdminLevel
)

func (a *AuthResponse) GetCookie() http.Cookie {
	cookie := http.Cookie{
		Name:     "bftapc",
		Value:    a.RefreshToken,
		HttpOnly: true,
		Secure:   false,                // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode, // Or http.SameSiteStrictMode, or http.SameSiteNoneMode with Secure=true
		Expires:  a.ExpiresIn,          // Example: expires in 7 days
		Path:     "/",                  // Makes the cookie available to all paths
	}
	return cookie
}

// GoogleUserInfo represents Google user information
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// AppleUserInfo represents Apple user information
type AppleUserInfo struct {
	Sub            string `json:"sub"`
	Email          string `json:"email"`
	EmailVerified  string `json:"email_verified"`
	IsPrivateEmail string `json:"is_private_email"`
	Name           struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"name"`
}

// Server represents the login server
type AuthServer struct {
	jwtSecret    []byte
	googleConfig OAuthConfig
	appleConfig  OAuthConfig
}

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	RedirectURL  string `json:"redirectURL"`
	TokenURL     string `json:"tokenURL"`
	UserInfoURL  string `json:"userInfoURL"`
}

type reqKey int

const key reqKey = iota

// JWT Claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Elev   bool   `json:"elev"`
	jwt.RegisteredClaims
}

// Register a new user
func (s *AuthServer) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req account.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	users, err := account.QueryUsers(map[string]interface{}{
		"email": req.Email,
	})
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if len(users) > 0 {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create account.User
	user := account.User{
		Email: req.Email,
		//Username:  req.Username,
		Password:  string(hashedPassword),
		Phone:     req.Phone,
		Provider:  "local",
		FirstName: req.FirstName,
		LastName:  req.LastName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := user.Save(); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate tokens
	response, err := s.generateTokens(user)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}
	_cookie := response.GetCookie()
	http.SetCookie(w, &_cookie)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&response)
}

// Login with email and password
func (s AuthServer) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find user
	user, err := account.QueryUser(map[string]interface{}{"email": req.Email})
	if err != nil || user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate tokens
	response, err := s.generateTokens(*user)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}
	_cookie := response.GetCookie()
	http.SetCookie(w, &_cookie)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&response)
}

// Google OAuth login
func (s AuthServer) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateRandomString(32)
	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid email profile&state=%s",
		s.googleConfig.ClientID,
		url.QueryEscape(s.googleConfig.RedirectURL),
		state,
	)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// Google OAuth callback
func (s AuthServer) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	var req OAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	tokenData, err := s.exchangeCodeForToken(req.Code, s.googleConfig)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	// Get user info from Google
	userInfo, err := s.getGoogleUserInfo(tokenData["access_token"].(string))
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Create or update user
	user, err := s.createOrUpdateOAuthUser(userInfo, "google")
	if err != nil {
		http.Error(w, "Failed to create/update user", http.StatusInternalServerError)
		return
	}

	// Generate tokens
	response, err := s.generateTokens(*user)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}
	_cookie := response.GetCookie()
	http.SetCookie(w, &_cookie)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Apple OAuth login
func (s AuthServer) HandleAppleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateRandomString(32)
	authURL := fmt.Sprintf(
		"https://appleid.apple.com/auth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=name email&response_mode=form_post&state=%s",
		s.appleConfig.ClientID,
		url.QueryEscape(s.appleConfig.RedirectURL),
		state,
	)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// Apple OAuth callback
func (s AuthServer) HandleAppleCallback(w http.ResponseWriter, r *http.Request) {
	var req OAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	tokenData, err := s.exchangeCodeForToken(req.Code, s.appleConfig)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	// Decode Apple ID token (simplified - in production, verify JWT signature)
	idToken := tokenData["id_token"].(string)
	userInfo, err := s.decodeAppleIDToken(idToken)
	if err != nil {
		http.Error(w, "Failed to decode ID token", http.StatusInternalServerError)
		return
	}

	// Create or update user
	user, err := s.createOrUpdateAppleUser(userInfo)
	if err != nil {
		http.Error(w, "Failed to create/update user", http.StatusInternalServerError)
		return
	}

	// Generate tokens
	response, err := s.generateTokens(*user)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}
	_cookie := response.GetCookie()
	http.SetCookie(w, &_cookie)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get current user info
func (s AuthServer) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	user, err := account.QueryUser(map[string]interface{}{"iD": userID})
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Refresh JWT token
func (s AuthServer) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate refresh token (simplified - in production, store and validate properly)
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Generate new tokens
	response, err := s.generateTokens(account.User{ID: claims.UserID, Email: claims.Email, IsAdmin: claims.Elev})
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}
	_cookie := response.GetCookie()
	http.SetCookie(w, &_cookie)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions

func (s AuthServer) generateTokens(user account.User) (*AuthResponse, error) {
	// Access token (15 minutes)
	accessClaims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Elev:   user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Refresh token (7 days)
	refreshClaims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Elev:   user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}
	_resp := AuthResponse{
		Token:        accessTokenString,
		RefreshToken: refreshTokenString,
		User:         user,
		ExpiresIn:    accessClaims.ExpiresAt.Time,
		AuthLevel:    LoginLevel,
	}
	return &_resp, nil
}

func (s AuthServer) exchangeCodeForToken(code string, config OAuthConfig) (map[string]interface{}, error) {
	data := url.Values{}
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", config.RedirectURL)

	resp, err := http.Post(config.TokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		return nil, err
	}

	return tokenData, nil
}

func (s AuthServer) getGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequest("GET", s.googleConfig.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (s AuthServer) decodeAppleIDToken(idToken string) (*AppleUserInfo, error) {
	// This is a simplified version - in production, you should verify the JWT signature
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode the payload (base64)
	payload := parts[1]
	// Add padding if needed
	for len(payload)%4 != 0 {
		payload += "="
	}

	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}

	var userInfo AppleUserInfo
	if err := json.Unmarshal(decoded, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (s AuthServer) createOrUpdateOAuthUser(userInfo *GoogleUserInfo, provider string) (*account.User, error) {
	var user account.User
	//err := s.db.Where("provider = ? AND provider_id = ?", provider, userInfo.ID).First(&user).Error

	_query := make(map[string]interface{})
	_query["provider"] = provider
	_query["provider_id"] = userInfo.ID

	users, err := account.QueryUsers(_query)
	if err != nil {
		return nil, err
	}
	for _, usr := range users {
		if usr.Email == userInfo.Email {
			user.Provider = provider
			user.ProviderID = userInfo.ID
			user.IsVerified = userInfo.VerifiedEmail
			user.Avatar = userInfo.Picture
			user.UpdatedAt = time.Now()
		} else {
			// Create new user
			user = account.User{
				Email:      userInfo.Email,
				Provider:   provider,
				ProviderID: userInfo.ID,
				FirstName:  userInfo.GivenName,
				LastName:   userInfo.FamilyName,
				Avatar:     userInfo.Picture,
				IsVerified: userInfo.VerifiedEmail,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
		}

		if err := user.Save(); err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s AuthServer) createOrUpdateAppleUser(userInfo *AppleUserInfo) (*account.User, error) {
	_query := make(map[string]interface{})
	_query["provider"] = "apple"
	_query["provider_id"] = userInfo.Sub

	users, err := account.QueryUsers(_query)
	if err != nil {
		return nil, err
	}
	var user account.User

	for _, usr := range users {
		// Check if user exists with same email
		if usr.Email == userInfo.Email {
			user = usr
			// Update existing user with Apple info
			user.Provider = "apple"
			user.ProviderID = userInfo.Sub
			user.IsVerified = userInfo.EmailVerified == "true"
			user.UpdatedAt = time.Now()

			user.Save()
			return &user, nil
		}
	}

	// Create new user
	user = account.User{
		Email:      userInfo.Email,
		Provider:   "apple",
		ProviderID: userInfo.Sub,
		FirstName:  userInfo.Name.FirstName,
		LastName:   userInfo.Name.LastName,
		IsVerified: userInfo.EmailVerified == "true",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := user.Save(); err != nil {
		return nil, err
	}

	return &user, nil
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return string(b)
}

// Authentication middleware
func (s AuthServer) AuthenticateMiddleware(admin bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return s.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		//check for admin access
		if admin && !claims.Elev {
			http.Error(w, "NoAdmin Access", http.StatusUnauthorized)
			return
		}

		// Add user ID to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, key, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// CORS middleware
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
