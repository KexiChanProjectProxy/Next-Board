package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/config"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/models"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(email, password string, role string) (*models.User, error)
	Login(email, password string) (string, string, error)
	RefreshToken(refreshToken string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
	GenerateTelegramLinkToken() (string, error)
}

type authService struct {
	cfg      *config.AuthConfig
	userRepo repository.UserRepository
	db       *gorm.DB
}

type Claims struct {
	UserID uint64 `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(cfg *config.AuthConfig, userRepo repository.UserRepository, db *gorm.DB) AuthService {
	return &authService{
		cfg:      cfg,
		userRepo: userRepo,
		db:       db,
	}
}

func (s *authService) Register(email, password, role string) (*models.User, error) {
	// Check if user exists
	_, err := s.userRepo.FindByEmail(email)
	if err == nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(email, password string) (string, string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if user.Banned {
		return "", "", errors.New("user is banned")
	}

	if err := s.ComparePassword(user.PasswordHash, password); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	// Generate access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *authService) RefreshToken(refreshToken string) (string, error) {
	var rt models.RefreshToken
	if err := s.db.Where("token = ? AND expires_at > ?", refreshToken, time.Now()).First(&rt).Error; err != nil {
		return "", errors.New("invalid refresh token")
	}

	user, err := s.userRepo.FindByID(rt.UserID)
	if err != nil {
		return "", err
	}

	if user.Banned {
		return "", errors.New("user is banned")
	}

	return s.generateAccessToken(user)
}

func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *authService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *authService) ComparePassword(hashedPassword, password string) error {
	// PHP's password_hash() uses $2y$ prefix for bcrypt, but Go's bcrypt library
	// only recognizes $2a$ and $2b$ prefixes. Since $2y$ and $2a$ are algorithmically
	// identical (the difference is only in the version identifier), we can safely
	// convert $2y$ to $2a$ for verification.
	if len(hashedPassword) > 3 && hashedPassword[:4] == "$2y$" {
		hashedPassword = "$2a$" + hashedPassword[4:]
	}

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *authService) generateAccessToken(user *models.User) (string, error) {
	expiresAt := time.Now().Add(s.cfg.GetAccessTokenDuration())

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *authService) generateRefreshToken(user *models.User) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	tokenString := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(s.cfg.GetRefreshTokenDuration())

	rt := &models.RefreshToken{
		UserID:    user.ID,
		Token:     tokenString,
		ExpiresAt: expiresAt,
	}

	if err := s.db.Create(rt).Error; err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *authService) GenerateTelegramLinkToken() (string, error) {
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}
