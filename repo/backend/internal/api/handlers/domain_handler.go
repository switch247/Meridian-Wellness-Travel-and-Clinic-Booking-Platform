package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"meridian/backend/internal/api/middleware"
	"meridian/backend/internal/api/response"
	"meridian/backend/internal/repository"
	"meridian/backend/internal/security"
	"meridian/backend/internal/service"

	"github.com/labstack/echo/v4"
)

type DomainHandler struct {
	profiles        *service.ProfileService
	booking         *service.BookingService
	repo            *repository.Repository
	encryptor       *security.Encryptor
	slotGranularity int
}

func optionalInt64(v string) (*int64, error) {
	if v == "" {
		return nil, nil
	}
	out, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func NewDomainHandler(
	profiles *service.ProfileService,
	booking *service.BookingService,
	repo *repository.Repository,
	encryptor *security.Encryptor,
	slotGranularity int,
) *DomainHandler {
	return &DomainHandler{
		profiles:        profiles,
		booking:         booking,
		repo:            repo,
		encryptor:       encryptor,
		slotGranularity: slotGranularity,
	}
}

type addressRequest struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
}

func (h *DomainHandler) AddAddress(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	var req addressRequest
	if err := response.BindAndValidate(c, &req, func(r *addressRequest) error {
		if r.Line1 == "" || r.City == "" || r.State == "" || r.PostalCode == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "line1, city, state, postalCode required")
		}
		return nil
	}); err != nil {
		return err
	}
	result, err := h.profiles.AddAddress(c.Request().Context(), uid, req.Line1, req.Line2, req.City, req.State, req.PostalCode)
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *DomainHandler) ListAddresses(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	items, err := h.profiles.ListAddresses(c.Request().Context(), uid)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

type contactRequest struct {
	Name         string `json:"name"`
	Relationship string `json:"relationship"`
	Phone        string `json:"phone"`
}

func (h *DomainHandler) AddContact(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	var req contactRequest
	if err := response.BindAndValidate(c, &req, func(r *contactRequest) error {
		if r.Name == "" || r.Phone == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "name and phone required")
		}
		return nil
	}); err != nil {
		return err
	}
	id, err := h.profiles.AddContact(c.Request().Context(), uid, req.Name, req.Relationship, req.Phone)
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]any{"id": id})
}

func (h *DomainHandler) ListContacts(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	items, err := h.profiles.ListContacts(c.Request().Context(), uid)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) DeleteContact(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "invalid contact id")
	}
	if err := h.profiles.DeleteContact(c.Request().Context(), uid, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return response.JSONError(c, http.StatusNotFound, "contact not found")
		}
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

func (h *DomainHandler) Catalog(c echo.Context) error {
	items, err := h.booking.Catalog(c.Request().Context())
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

type holdRequest struct {
	PackageID int64  `json:"packageId"`
	HostID    int64  `json:"hostId"`
	RoomID    int64  `json:"roomId"`
	SlotStart string `json:"slotStart"`
	Duration  int    `json:"duration"`
}

func (h *DomainHandler) PlaceHold(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	var req holdRequest
	if err := response.BindAndValidate(c, &req, func(r *holdRequest) error {
		if r.PackageID <= 0 || r.HostID <= 0 || r.RoomID <= 0 || r.SlotStart == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "packageId, hostId, roomId, slotStart required")
		}
		return nil
	}); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339, req.SlotStart)
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, "slotStart must be RFC3339")
	}
	out, err := h.booking.PlaceHold(c.Request().Context(), uid, req.PackageID, req.HostID, req.RoomID, t, req.Duration)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAddressRequired):
			return response.JSONError(c, http.StatusBadRequest, "add a profile address before booking")
		case errors.Is(err, repository.ErrPostalBlocked):
			return response.JSONError(c, http.StatusForbidden, err.Error())
		case errors.Is(err, repository.ErrServiceWindowViolation):
			return response.JSONError(c, http.StatusConflict, err.Error())
		default:
			return response.JSONError(c, http.StatusConflict, err.Error())
		}
	}
	return c.JSON(http.StatusCreated, out)
}

func (h *DomainHandler) ListHolds(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	items, err := h.repo.ListHoldsByUser(c.Request().Context(), uid)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) ListBookingHistory(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	items, err := h.repo.ListBookingHistoryByUser(c.Request().Context(), uid)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

type roleRequest struct {
	TargetUserID int64  `json:"targetUserId"`
	Role         string `json:"role"`
}

type regionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    *int64 `json:"parentId"`
}

type serviceRuleRequest struct {
	AllowHomePickup bool   `json:"allowHomePickup"`
	AllowMailDocs   bool   `json:"allowMailDocuments"`
	Blocked         bool   `json:"blocked"`
	StartTime       string `json:"startTime"`
	EndTime         string `json:"endTime"`
}

type blockedPostalRequest struct {
	ServiceRuleID int64  `json:"serviceRuleId"`
	PostalCode    string `json:"postalCode"`
}

func (h *DomainHandler) AssignRole(c echo.Context) error {
	actorID, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	var req roleRequest
	if err := response.BindAndValidate(c, &req, func(r *roleRequest) error {
		if r.TargetUserID <= 0 || r.Role == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "targetUserId and role required")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := h.repo.AssignRole(c.Request().Context(), actorID, req.TargetUserID, req.Role); err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

func (h *DomainHandler) ListUsers(c echo.Context) error {
	roleFilter := c.QueryParam("role")
	items, err := h.repo.ListUsers(c.Request().Context(), roleFilter)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) ListHosts(c echo.Context) error {
	items, err := h.repo.ListUsers(c.Request().Context(), "")
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	// Filter to only coach and clinician
	filtered := []map[string]any{}
	for _, item := range items {
		roles, ok := item["roles"].([]string)
		if !ok {
			continue
		}
		for _, role := range roles {
			if role == "coach" || role == "clinician" {
				filtered = append(filtered, item)
				break
			}
		}
	}
	return c.JSON(http.StatusOK, map[string]any{"items": filtered})
}

func (h *DomainHandler) ListRoleAudits(c echo.Context) error {
	items, err := h.repo.ListPermissionAudits(c.Request().Context())
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) ListRegions(c echo.Context) error {
	items, err := h.repo.ListRegions(c.Request().Context())
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) CreateRegion(c echo.Context) error {
	var req regionRequest
	if err := c.Bind(&req); err != nil || req.Name == "" {
		return response.JSONError(c, http.StatusBadRequest, "name required")
	}
	id, err := h.repo.CreateRegion(c.Request().Context(), req.Name, req.Description, req.ParentID)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]any{"id": id})
}

func (h *DomainHandler) UpsertServiceRule(c echo.Context) error {
	regionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || regionID <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "invalid region id")
	}
	var req serviceRuleRequest
	if err := c.Bind(&req); err != nil {
		return response.JSONError(c, http.StatusBadRequest, "invalid payload")
	}
	var startTime, endTime *time.Time
	if req.StartTime != "" {
		parsed, err := time.Parse("15:04", req.StartTime)
		if err != nil {
			return response.JSONError(c, http.StatusBadRequest, "startTime must be HH:MM")
		}
		startTime = &parsed
	}
	if req.EndTime != "" {
		parsed, err := time.Parse("15:04", req.EndTime)
		if err != nil {
			return response.JSONError(c, http.StatusBadRequest, "endTime must be HH:MM")
		}
		endTime = &parsed
	}
	id, err := h.repo.UpsertServiceRule(c.Request().Context(), regionID, req.AllowHomePickup, req.AllowMailDocs, req.Blocked, startTime, endTime)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"serviceRuleId": id})

}

func (h *DomainHandler) ListBlockedPostalCodes(c echo.Context) error {
	items, err := h.repo.ListBlockedPostalCodes(c.Request().Context())
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) AddBlockedPostalCode(c echo.Context) error {
	var req blockedPostalRequest
	if err := c.Bind(&req); err != nil || req.ServiceRuleID <= 0 || req.PostalCode == "" {
		return response.JSONError(c, http.StatusBadRequest, "serviceRuleId and postalCode required")
	}
	id, err := h.repo.AddBlockedPostalCode(c.Request().Context(), req.ServiceRuleID, req.PostalCode)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]any{"id": id})
}

func (h *DomainHandler) HostAgenda(c echo.Context) error {
	requester, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	hostID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, "invalid host id")
	}
	if !middleware.HasAnyRole(c, "operations", "admin") && hostID != requester {
		return response.JSONError(c, http.StatusForbidden, "ownership check failed")
	}
	items, err := h.repo.ListHostAgenda(c.Request().Context(), hostID)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	h.enrichSessionNotes(items)
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) RoomAgenda(c echo.Context) error {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, "invalid room id")
	}
	items, err := h.repo.ListRoomAgenda(c.Request().Context(), roomID)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	h.enrichSessionNotes(items)
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

type bookingStatusRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

func (h *DomainHandler) UpdateBookingStatus(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	bookingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || bookingID <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "invalid booking id")
	}
	var req bookingStatusRequest
	if err := response.BindAndValidate(c, &req, func(r *bookingStatusRequest) error {
		if strings.TrimSpace(r.Status) == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "status required")
		}
		return nil
	}); err != nil {
		return err
	}
	trimmedStatus := strings.TrimSpace(req.Status)
	if !middleware.HasAnyRole(c, "operations", "admin") {
		hostID, hostErr := h.repo.GetBookingHost(c.Request().Context(), bookingID)
		if hostErr != nil {
			if errors.Is(hostErr, repository.ErrNotFound) {
				return response.JSONError(c, http.StatusNotFound, "booking not found")
			}
			return response.JSONError(c, http.StatusInternalServerError, hostErr.Error())
		}
		if hostID != uid {
			return response.JSONError(c, http.StatusForbidden, "ownership check failed")
		}
	}
	var encryptedNotes *string
	if trimmed := strings.TrimSpace(req.Notes); trimmed != "" {
		if h.encryptor == nil {
			return response.JSONError(c, http.StatusInternalServerError, "encryption unavailable")
		}
		payload, encErr := h.encryptor.Encrypt(trimmed)
		if encErr != nil {
			return response.JSONError(c, http.StatusInternalServerError, "notes encryption failed")
		}
		encryptedNotes = &payload
	}
	if err := h.repo.UpdateBookingStatus(c.Request().Context(), bookingID, trimmedStatus, encryptedNotes); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return response.JSONError(c, http.StatusNotFound, "booking not found")
		}
		if errors.Is(err, repository.ErrInvalidBookingStatus) {
			return response.JSONError(c, http.StatusBadRequest, "invalid status")
		}
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

func (h *DomainHandler) GetUser(c echo.Context) error {
	requester, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, "invalid user id")
	}
	if !middleware.OwnsResource(requester, id) && !middleware.HasAnyRole(c, "operations", "admin") {
		return response.JSONError(c, http.StatusForbidden, "ownership check failed")
	}
	u, err := h.repo.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return response.JSONError(c, http.StatusNotFound, "user not found")
	}
	return c.JSON(http.StatusOK, map[string]any{"id": u.ID, "username": u.Username})
}

func (h *DomainHandler) DeleteAddress(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "invalid address id")
	}
	if err := h.repo.DeleteAddress(c.Request().Context(), uid, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return response.JSONError(c, http.StatusNotFound, "address not found")
		}
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

func (h *DomainHandler) CancelHold(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "invalid hold id")
	}
	if err := h.repo.CancelHoldByUser(c.Request().Context(), uid, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return response.JSONError(c, http.StatusNotFound, "hold not found")
		}
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

func (h *DomainHandler) Routes(c echo.Context) error {
	items, err := h.repo.ListRoutes(c.Request().Context())
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) Hotels(c echo.Context) error {
	items, err := h.repo.ListHotels(c.Request().Context())
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) Attractions(c echo.Context) error {
	items, err := h.repo.ListAttractions(c.Request().Context())
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) AvailableSlots(c echo.Context) error {
	hostID, err := strconv.ParseInt(c.QueryParam("hostId"), 10, 64)
	if err != nil || hostID <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "invalid hostId")
	}
	roomID, err := strconv.ParseInt(c.QueryParam("roomId"), 10, 64)
	if err != nil || roomID <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "invalid roomId")
	}
	day, err := time.Parse("2006-01-02", c.QueryParam("day"))
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, "invalid day")
	}
	duration, err := strconv.Atoi(c.QueryParam("duration"))
	if err != nil || (duration != 30 && duration != 45 && duration != 60) {
		return response.JSONError(c, http.StatusBadRequest, "duration must be 30/45/60")
	}
	items, err := h.repo.ListAvailableSlots(c.Request().Context(), hostID, roomID, day, duration, h.slotGranularity)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

type confirmHoldRequest struct {
	HoldID  int64 `json:"holdId"`
	Version int   `json:"version"`
}

func (h *DomainHandler) ConfirmHold(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	var req confirmHoldRequest
	if err := response.BindAndValidate(c, &req, func(r *confirmHoldRequest) error {
		if r.HoldID <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "holdId required")
		}
		return nil
	}); err != nil {
		return err
	}
	bookingID, err := h.repo.ConfirmHold(c.Request().Context(), uid, req.HoldID, req.Version)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return response.JSONError(c, http.StatusNotFound, "hold not found")
		}
		if errors.Is(err, repository.ErrHoldExpired) {
			return response.JSONError(c, http.StatusConflict, "hold expired")
		}
		return response.JSONError(c, http.StatusConflict, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"bookingId": bookingID, "status": "confirmed"})
}

type postRequest struct {
	Title          string `json:"title"`
	Body           string `json:"body"`
	DestinationID  *int64 `json:"destinationId"`
	ProviderUserID *int64 `json:"providerUserId"`
}

func (h *DomainHandler) ListPosts(c echo.Context) error {
	items, err := h.repo.ListCommunityPosts(c.Request().Context())
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) CreatePost(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	var req postRequest
	if err := response.BindAndValidate(c, &req, func(r *postRequest) error {
		if r.Title == "" || r.Body == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "title and body required")
		}
		return nil
	}); err != nil {
		return err
	}
	id, err := h.repo.CreateCommunityPost(c.Request().Context(), uid, req.Title, req.Body, req.DestinationID, req.ProviderUserID)
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	_ = h.repo.CreateNotification(c.Request().Context(), uid, "community", "Post published", req.Title, "post", &id)
	return c.JSON(http.StatusCreated, map[string]any{"id": id})
}

type commentRequest struct {
	Body            string `json:"body"`
	ParentCommentID *int64 `json:"parentCommentId"`
}

func (h *DomainHandler) ListComments(c echo.Context) error {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || postID <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "invalid post id")
	}
	items, err := h.repo.ListCommentsByPost(c.Request().Context(), postID)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) CreateComment(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || postID <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "invalid post id")
	}
	var req commentRequest
	if err := response.BindAndValidate(c, &req, func(r *commentRequest) error {
		if r.Body == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "body required")
		}
		return nil
	}); err != nil {
		return err
	}
	id, err := h.repo.CreateComment(c.Request().Context(), uid, postID, req.ParentCommentID, req.Body)
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]any{"id": id})
}

type packageActionRequest struct {
	PackageID int64 `json:"packageId"`
}

func (h *DomainHandler) FavoritePackage(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	var req packageActionRequest
	if err := c.Bind(&req); err != nil || req.PackageID <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "packageId required")
	}
	if err := h.repo.ToggleFavorite(c.Request().Context(), uid, req.PackageID); err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

type userActionRequest struct {
	UserID int64 `json:"userId"`
}

func (h *DomainHandler) FollowUser(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	var req userActionRequest
	if err := c.Bind(&req); err != nil || req.UserID <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "userId required")
	}
	if err := h.repo.FollowUser(c.Request().Context(), uid, req.UserID); err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

func (h *DomainHandler) BlockUser(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	var req userActionRequest
	if err := c.Bind(&req); err != nil || req.UserID <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "userId required")
	}
	if err := h.repo.BlockUser(c.Request().Context(), uid, req.UserID); err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

type likeRequest struct {
	TargetType string `json:"targetType"`
	TargetID   int64  `json:"targetId"`
}

func (h *DomainHandler) LikeTarget(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	var req likeRequest
	if err := c.Bind(&req); err != nil || req.TargetType == "" || req.TargetID <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "targetType and targetId required")
	}
	if err := h.repo.LikeTarget(c.Request().Context(), uid, req.TargetType, req.TargetID); err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

type reportRequest struct {
	TargetType string `json:"targetType"`
	TargetID   int64  `json:"targetId"`
	Reason     string `json:"reason"`
}

func (h *DomainHandler) ReportContent(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	var req reportRequest
	if err := c.Bind(&req); err != nil || req.TargetType == "" || req.TargetID <= 0 || req.Reason == "" {
		return response.JSONError(c, http.StatusBadRequest, "targetType, targetId, reason required")
	}
	id, err := h.repo.ReportTarget(c.Request().Context(), uid, req.TargetType, req.TargetID, req.Reason)
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]any{"id": id})
}

type resolveReportRequest struct {
	Status  string `json:"status"`
	Outcome string `json:"outcome"`
}

func (h *DomainHandler) ResolveReport(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	reportID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || reportID <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "invalid report id")
	}
	var req resolveReportRequest
	if err := c.Bind(&req); err != nil || req.Status == "" {
		return response.JSONError(c, http.StatusBadRequest, "status required")
	}
	if err := h.repo.ResolveReport(c.Request().Context(), reportID, uid, req.Status, req.Outcome); err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

func (h *DomainHandler) ListNotifications(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	items, err := h.repo.ListNotifications(c.Request().Context(), uid)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) MarkNotificationRead(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return response.JSONError(c, http.StatusBadRequest, "invalid notification id")
	}
	if err := h.repo.MarkNotificationRead(c.Request().Context(), uid, id); err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

type emailQueueRequest struct {
	TemplateKey    string `json:"templateKey"`
	RecipientLabel string `json:"recipientLabel"`
	Subject        string `json:"subject"`
	Body           string `json:"body"`
}

func (h *DomainHandler) QueueEmail(c echo.Context) error {
	var req emailQueueRequest
	if err := c.Bind(&req); err != nil || req.TemplateKey == "" || req.RecipientLabel == "" || req.Subject == "" || req.Body == "" {
		return response.JSONError(c, http.StatusBadRequest, "templateKey, recipientLabel, subject, body required")
	}
	id, err := h.repo.QueueEmailTemplate(c.Request().Context(), req.TemplateKey, req.RecipientLabel, req.Subject, req.Body)
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]any{"id": id})
}

func (h *DomainHandler) ListEmailQueue(c echo.Context) error {
	items, err := h.repo.ListEmailQueue(c.Request().Context())
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) ExportEmailQueue(c echo.Context) error {
	outDir := os.Getenv("EXPORT_DIR")
	if outDir == "" {
		outDir = "/tmp/exports"
	}
	path, err := h.repo.ExportEmailQueueCSV(c.Request().Context(), outDir)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"path": path})
}

func (h *DomainHandler) AnalyticsKPIs(c echo.Context) error {
	from, err := time.Parse("2006-01-02", c.QueryParam("from"))
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, "invalid from date")
	}
	to, err := time.Parse("2006-01-02", c.QueryParam("to"))
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, "invalid to date")
	}
	providerID, err := optionalInt64(c.QueryParam("providerId"))
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, "invalid providerId")
	}
	packageID, err := optionalInt64(c.QueryParam("packageId"))
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, "invalid packageId")
	}
	kpis, err := h.repo.AnalyticsKPIs(c.Request().Context(), from, to.Add(24*time.Hour), providerID, packageID)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"kpis": kpis})
}

func (h *DomainHandler) ExportAnalytics(c echo.Context) error {
	from, err := time.Parse("2006-01-02", c.QueryParam("from"))
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, "invalid from date")
	}
	to, err := time.Parse("2006-01-02", c.QueryParam("to"))
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, "invalid to date")
	}
	providerID, _ := optionalInt64(c.QueryParam("providerId"))
	packageID, _ := optionalInt64(c.QueryParam("packageId"))
	kpis, err := h.repo.AnalyticsKPIs(c.Request().Context(), from, to.Add(24*time.Hour), providerID, packageID)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	outDir := os.Getenv("EXPORT_DIR")
	if outDir == "" {
		outDir = "/tmp/exports"
	}
	path, err := h.repo.ExportAnalyticsCSV(c.Request().Context(), outDir, 0, "manual", kpis)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"path": path})
}

type scheduleReportRequest struct {
	ReportType   string         `json:"reportType"`
	Parameters   map[string]any `json:"parameters"`
	ScheduledFor string         `json:"scheduledFor"`
}

func (h *DomainHandler) ScheduleReport(c echo.Context) error {
	uid, ok := middleware.UserID(c)
	if !ok {
		return response.JSONError(c, http.StatusUnauthorized, "missing user context")
	}
	var req scheduleReportRequest
	if err := c.Bind(&req); err != nil || req.ReportType == "" || req.ScheduledFor == "" {
		return response.JSONError(c, http.StatusBadRequest, "reportType and scheduledFor required")
	}
	when, err := time.Parse(time.RFC3339, req.ScheduledFor)
	if err != nil {
		return response.JSONError(c, http.StatusBadRequest, "scheduledFor must be RFC3339")
	}
	params, _ := json.Marshal(req.Parameters)
	id, err := h.repo.ScheduleReportJob(c.Request().Context(), req.ReportType, string(params), uid, when)
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]any{"id": id, "status": "scheduled"})
}

func (h *DomainHandler) ListReportJobs(c echo.Context) error {
	items, err := h.repo.ListReportJobs(c.Request().Context())
	if err != nil {
		return response.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *DomainHandler) enrichSessionNotes(items []map[string]any) {
	if h.encryptor == nil {
		return
	}
	for _, item := range items {
		raw, ok := item["sessionNotesEncrypted"].(string)
		if !ok || raw == "" {
			delete(item, "sessionNotesEncrypted")
			continue
		}
		if decoded, err := h.encryptor.Decrypt(raw); err == nil {
			// Do NOT return full decrypted session notes in API responses.
			// Provide only a short summary and a presence flag to avoid leaking sensitive data.
			item["sessionNotesSummary"] = summarizeText(decoded)
			item["hasSessionNotes"] = true
		}
		delete(item, "sessionNotesEncrypted")
	}
}

func summarizeText(value string) string {
	text := strings.TrimSpace(value)
	if len(text) == 0 {
		return ""
	}
	if len(text) <= 120 {
		return text
	}
	return text[:120] + "..."
}
