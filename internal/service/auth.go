package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/realPointer/banners/internal/entity"
	"github.com/realPointer/banners/internal/repository"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type TokenClaims struct {
	jwt.StandardClaims
	Username string
	Role     string
}

type AuthService struct {
	userRepo repository.User
	signKey  string
	tokenTTL time.Duration
	salt     string
	l        *zerolog.Logger
}

func NewAuthService(l *zerolog.Logger, userRepo repository.User, signKey string, tokenTTL time.Duration, salt string) *AuthService {
	return &AuthService{
		l:        l,
		userRepo: userRepo,
		signKey:  signKey,
		tokenTTL: tokenTTL,
		salt:     salt,
	}
}

func (s *AuthService) Register(ctx context.Context, username, password, role string) error {
	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		return err
	}

	user := &entity.User{
		Username: username,
		Password: hashedPassword,
		Role:     role,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password+s.salt))
	if err != nil {
		return "", errors.New("invalid password")
	}

	claims := &TokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.tokenTTL).Unix(),
		},
		Username: user.Username,
		Role:     user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.signKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+s.salt), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (s *AuthService) ParseToken(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.signKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
