package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	logger    *slog.Logger
}

type LoginResult struct {
	Token string   `json:"token"`
	Roles []string `json:"roles"`
}

func NewAuthService(repo *repository.Repository, cfg config.Config, encryptor *security.Encryptor, logger *slog.Logger) *AuthService {
	return &AuthService{repo: repo, cfg: cfg, encryptor: encryptor, logger: logger}
}

func (s *AuthService) Register(ctx context.Context, username, password, phone, address string) (int64, error) {
	if err := security.ValidatePassword(password); err != nil {
		s.logger.Info("[auth] register failed: invalid password", "username", username, "error", err)
		return 0, err
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		s.logger.Error("[auth] register failed: hash password", "username", username, "error", err)
		return 0, err
	}
	encryptedPhone, err := s.encryptor.Encrypt(phone)
	if err != nil {
		s.logger.Error("[auth] register failed: encrypt phone", "username", username, "error", err)
		return 0, err
	}
	encryptedAddress, err := s.encryptor.Encrypt(address)
	if err != nil {
		s.logger.Error("[auth] register failed: encrypt address", "username", username, "error", err)
		return 0, err
	}
	id, err := s.repo.CreateUser(ctx, username, hash, encryptedPhone, encryptedAddress)
	if err != nil {
		s.logger.Error("[auth] register failed: create user", "username", username, "error", err)
		return 0, err
	}
	if err := s.repo.AssignRole(ctx, id, id, "traveler"); err != nil {
		s.logger.Error("[auth] register failed: assign role", "user_id", id, "error", err)
		return 0, err
	}
	s.logger.Info("[auth] register success", "user_id", id, "username", username)
	return id, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (LoginResult, error) {
	u, roles, err := s.repo.FindUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.logger.Info("[auth] login failed: user not found", "username", username)
			return LoginResult{}, errors.New("invalid credentials")
		}
		s.logger.Error("[auth] login failed: find user", "username", username, "error", err)
		return LoginResult{}, err
	}
	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		s.logger.Info("[auth] login failed: account locked", "user_id", u.ID, "username", username, "locked_until", u.LockedUntil)
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
		s.logger.Info("[auth] login failed: invalid password", "user_id", u.ID, "username", username, "attempts", attempts)
		return LoginResult{}, errors.New("invalid credentials")
	}
	if err := s.repo.ResetFailedAttempts(ctx, u.ID); err != nil {
		s.logger.Error("[auth] login failed: reset attempts", "user_id", u.ID, "error", err)
		return LoginResult{}, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   u.ID,
		"roles": roles,
		"exp":   time.Now().Add(s.cfg.TokenTTL).Unix(),
	})
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		s.logger.Error("[auth] login failed: sign token", "user_id", u.ID, "error", err)
		return LoginResult{}, fmt.Errorf("sign token: %w", err)
	}
	s.logger.Info("[auth] login success", "user_id", u.ID, "username", username, "roles", roles)
	return LoginResult{Token: signed, Roles: roles}, nil
}

func (s *AuthService) Me(ctx context.Context, userID int64) (map[string]any, error) {
	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("[auth] me failed: user not found", "user_id", userID, "error", fmt.Errorf("user not found: %w", err))
		return nil, fmt.Errorf("user not found: %w", err)
	}
	roles, err := s.repo.GetRolesForUser(ctx, userID)
	if err != nil {
		s.logger.Error("[auth] me failed: get roles", "user_id", userID, "error", err)
		return nil, err
	}
	phone, err := s.encryptor.Decrypt(u.EncryptedPhone)
	if err != nil {
		s.logger.Error("[auth] me failed: decrypt phone", "user_id", userID, "error", err)
		return nil, err
	}
	address, err := s.encryptor.Decrypt(u.EncryptedAddress)
	if err != nil {
		s.logger.Error("[auth] me failed: decrypt address", "user_id", userID, "error", err)
		return nil, err
	}
	s.logger.Info("[auth] me success", "user_id", userID, "username", u.Username)
	return map[string]any{
		"id":       u.ID,
		"username": u.Username,
		"roles":    roles,
		"phone":    security.MaskPhone(phone),
		"address":  security.MaskAddress(address),
	}, nil
}
