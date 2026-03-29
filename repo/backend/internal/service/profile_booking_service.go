package service

import (
	"context"
	"fmt"
	"time"

	"meridian/backend/internal/config"
	"meridian/backend/internal/repository"
	"meridian/backend/internal/security"
)

type ProfileService struct {
	repo      *repository.Repository
	cfg       config.Config
	encryptor *security.Encryptor
}

func NewProfileService(repo *repository.Repository, cfg config.Config, encryptor *security.Encryptor) *ProfileService {
	return &ProfileService{repo: repo, cfg: cfg, encryptor: encryptor}
}

func (s *ProfileService) AddAddress(ctx context.Context, userID int64, line1, line2, city, state, postal string) (map[string]any, error) {
	normalized := security.NormalizeUSAddress(line1, city, state, postal)
	duplicate, err := s.repo.AddressExistsByNormalizedKey(ctx, userID, normalized)
	if err != nil {
		return nil, err
	}
	coverage := security.InCoverage(postal, s.cfg.AllowedPostalCode)
	encLine1, err := s.encryptor.Encrypt(line1)
	if err != nil {
		return nil, err
	}
	encLine2 := ""
	if line2 != "" {
		encLine2, err = s.encryptor.Encrypt(line2)
		if err != nil {
			return nil, err
		}
	}
	maskedLine1 := security.MaskAddress(line1)
	maskedLine2 := ""
	if line2 != "" {
		maskedLine2 = security.MaskAddress(line2)
	}
	if err := s.repo.CreateAddress(ctx, userID, maskedLine1, maskedLine2, city, state, postal, normalized, coverage, duplicate, encLine1, encLine2); err != nil {
		return nil, err
	}
	return map[string]any{
		"normalized": normalized,
		"duplicate":  duplicate,
		"inCoverage": coverage,
	}, nil
}

func (s *ProfileService) ListAddresses(ctx context.Context, userID int64) ([]map[string]any, error) {
	items, err := s.repo.ListAddressesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		line1Masked := ""
		if enc, ok := items[i]["line1Encrypted"].(string); ok && enc != "" {
			if plain, decErr := s.encryptor.Decrypt(enc); decErr == nil {
				line1Masked = security.MaskAddress(plain)
			}
		}
		if line1Masked == "" {
			if stored, ok := items[i]["line1"].(string); ok {
				line1Masked = stored
			}
		}
		line2Masked := ""
		if enc, ok := items[i]["line2Encrypted"].(string); ok && enc != "" {
			if plain, decErr := s.encryptor.Decrypt(enc); decErr == nil {
				line2Masked = security.MaskAddress(plain)
			}
		}
		if line2Masked == "" {
			if stored, ok := items[i]["line2"].(string); ok {
				line2Masked = stored
			}
		}
		items[i]["line1Masked"] = line1Masked
		items[i]["line2Masked"] = line2Masked
		delete(items[i], "line1Encrypted")
		delete(items[i], "line2Encrypted")
	}
	return items, nil
}

func (s *ProfileService) AddContact(ctx context.Context, userID int64, name, relationship, phone string) (int64, error) {
	encPhone, err := s.encryptor.Encrypt(phone)
	if err != nil {
		return 0, err
	}
	maskedPhone := security.MaskPhone(phone)
	return s.repo.CreateContact(ctx, userID, name, relationship, maskedPhone, encPhone)
}

func (s *ProfileService) ListContacts(ctx context.Context, userID int64) ([]map[string]any, error) {
	return s.repo.ListContactsByUser(ctx, userID)
}

func (s *ProfileService) DeleteContact(ctx context.Context, userID, contactID int64) error {
	return s.repo.DeleteContact(ctx, userID, contactID)
}

type BookingService struct {
	repo *repository.Repository
	cfg  config.Config
}

func NewBookingService(repo *repository.Repository, cfg config.Config) *BookingService {
	return &BookingService{repo: repo, cfg: cfg}
}

func (s *BookingService) PlaceHold(ctx context.Context, userID, packageID, hostID, roomID int64, slotStart time.Time, duration int) (map[string]any, error) {
	if duration != 30 && duration != 45 && duration != 60 {
		return nil, fmt.Errorf("duration must be one of 30,45,60")
	}
	if err := s.repo.ReleaseExpiredHolds(ctx); err != nil {
		return nil, err
	}
	expires := time.Now().Add(s.cfg.ReservationHold)
	holdID, version, err := s.repo.CreateReservationHold(ctx, userID, packageID, hostID, roomID, slotStart, duration, expires)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"holdId":      holdID,
		"version":     version,
		"expiresAt":   expires,
		"status":      "active",
		"reservation": "temporary",
	}, nil
}

func (s *BookingService) Catalog(ctx context.Context) ([]map[string]any, error) {
	return s.repo.ListCatalog(ctx)
}
