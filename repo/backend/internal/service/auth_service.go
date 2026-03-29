package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"meridian/backend/internal/config"
	"meridian/backend/internal/repository"
	"meridian/backend/internal/security"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	repo      *repository.Repository
	cfg       config.Config
	encryptor *security.Encryptor
}

type LoginResult struct {
	Token string   `json:"token"`
	Roles []string `json:"roles"`
}

func NewAuthService(repo *repository.Repository, cfg config.Config, encryptor *security.Encryptor) *AuthService {
	return &AuthService{repo: repo, cfg: cfg, encryptor: encryptor}
}

func (s *AuthService) Register(ctx context.Context, username, password, phone, address string) (int64, error) {
	if err := security.ValidatePassword(password); err != nil {
		return 0, err
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return 0, err
	}
	encryptedPhone, err := s.encryptor.Encrypt(phone)
	if err != nil {
		return 0, err
	}
	encryptedAddress, err := s.encryptor.Encrypt(address)
	if err != nil {
		return 0, err
	}
	id, err := s.repo.CreateUser(ctx, username, hash, encryptedPhone, encryptedAddress)
	if err != nil {
		return 0, err
	}
	return id, s.repo.AssignRole(ctx, id, id, "traveler")
}

func (s *AuthService) Login(ctx context.Context, username, password string) (LoginResult, error) {
	u, roles, err := s.repo.FindUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return LoginResult{}, errors.New("invalid credentials")
		}
		return LoginResult{}, err
	}
	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		return LoginResult{}, errors.New("account locked; try later")
	}
	if err := security.ComparePassword(u.PasswordHash, password); err != nil {
		attempts := u.FailedAttempts + 1
		var lock *time.Time
		if attempts >= s.cfg.LockoutThreshold {
			t := time.Now().Add(s.cfg.LockoutDuration)
			lock = &t
		}
		_ = s.repo.SetUserFailedAttempts(ctx, u.ID, attempts, lock)
		return LoginResult{}, errors.New("invalid credentials")
	}
	if err := s.repo.ResetFailedAttempts(ctx, u.ID); err != nil {
		return LoginResult{}, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   u.ID,
		"roles": roles,
		"exp":   time.Now().Add(s.cfg.TokenTTL).Unix(),
	})
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return LoginResult{}, fmt.Errorf("sign token: %w", err)
	}
	return LoginResult{Token: signed, Roles: roles}, nil
}

func (s *AuthService) Me(ctx context.Context, userID int64) (map[string]any, error) {
	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	roles, err := s.repo.GetRolesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	phone, err := s.encryptor.Decrypt(u.EncryptedPhone)
	if err != nil {
		return nil, err
	}
	address, err := s.encryptor.Decrypt(u.EncryptedAddress)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id":       u.ID,
		"username": u.Username,
		"roles":    roles,
		"phone":    security.MaskPhone(phone),
		"address":  security.MaskAddress(address),
	}, nil
}
