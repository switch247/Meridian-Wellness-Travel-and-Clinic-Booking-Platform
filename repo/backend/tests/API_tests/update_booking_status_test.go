package integration_tests

import (
	"net/http"
	"strconv"
	"testing"
)

func TestUpdateBookingStatus(t *testing.T) {
	token := makeAdminToken(t)
	// Place a booking hold first (reuse helper or create minimal booking)
	// ...simulate hold placement, get bookingID...
	bookingID := 1 // Replace with actual booking ID from hold
	res, body := call(http.MethodPost, "/bookings/"+strconv.Itoa(bookingID)+"/status", token, map[string]any{
		"status": "confirmed",
	}, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for update booking status, got %d body=%+v", res.StatusCode, body)
	}
}
