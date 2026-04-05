package integration_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"
)

func pickBookablePackageAndSlot(t *testing.T) (int64, time.Time) {
	t.Helper()
	res, body := call(http.MethodGet, "/catalog", "", nil, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 catalog, got %d body=%+v", res.StatusCode, body)
	}
	items, ok := body["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected non-empty catalog items")
	}
	for _, raw := range items {
		it, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		inv, _ := it["inventoryRemaining"].(float64)
		if inv <= 0 {
			continue
		}
		idNum, ok := it["id"].(float64)
		if !ok || idNum <= 0 {
			continue
		}
		serviceDateRaw, ok := it["serviceDate"].(string)
		if !ok || serviceDateRaw == "" {
			continue
		}
		serviceDate := serviceDateRaw
		if len(serviceDate) >= 10 {
			serviceDate = serviceDate[:10]
		}
		day, err := time.Parse("2006-01-02", serviceDate)
		if err != nil {
			continue
		}
		// Keep deterministic slot time; hold placement validates package calendar date.
		slot := time.Date(day.Year(), day.Month(), day.Day(), 11, 0, 0, 0, time.UTC)
		return int64(idNum), slot
	}
	t.Fatalf("no bookable package found in catalog")
	return 0, time.Time{}
}

func pickBookablePackageAndAvailableSchedulingSlot(t *testing.T, token string, duration int) (int64, int64, int64, time.Time) {
	t.Helper()
	res, body := call(http.MethodGet, "/catalog", "", nil, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 catalog, got %d body=%+v", res.StatusCode, body)
	}
	items, ok := body["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected non-empty catalog items")
	}

	for _, raw := range items {
		it, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		inv, _ := it["inventoryRemaining"].(float64)
		if inv <= 0 {
			continue
		}
		idNum, ok := it["id"].(float64)
		if !ok || idNum <= 0 {
			continue
		}
		serviceDateRaw, ok := it["serviceDate"].(string)
		if !ok || serviceDateRaw == "" {
			continue
		}
		serviceDate := serviceDateRaw
		if len(serviceDate) >= 10 {
			serviceDate = serviceDate[:10]
		}
		day, err := time.Parse("2006-01-02", serviceDate)
		if err != nil {
			continue
		}
		hostID, roomID, slotStart, found := findAvailableSchedulingSlot(t, token, day, duration)
		if found {
			return int64(idNum), hostID, roomID, slotStart
		}
	}

	t.Fatalf("no bookable package with available scheduling slot found for duration=%d", duration)
	return 0, 0, 0, time.Time{}
}

func findAvailableSchedulingSlot(t *testing.T, token string, day time.Time, duration int) (int64, int64, time.Time, bool) {
	t.Helper()
	dayStr := day.UTC().Format("2006-01-02")

	resHosts, bodyHosts := call(http.MethodGet, "/scheduling/hosts", token, nil, t)
	if resHosts.StatusCode != http.StatusOK {
		return 0, 0, time.Time{}, false
	}
	resRooms, bodyRooms := call(http.MethodGet, "/scheduling/rooms", token, nil, t)
	if resRooms.StatusCode != http.StatusOK {
		return 0, 0, time.Time{}, false
	}

	hosts, ok := bodyHosts["items"].([]any)
	if !ok || len(hosts) == 0 {
		return 0, 0, time.Time{}, false
	}
	rooms, ok := bodyRooms["items"].([]any)
	if !ok || len(rooms) == 0 {
		return 0, 0, time.Time{}, false
	}

	for _, hr := range hosts {
		h, ok := hr.(map[string]any)
		if !ok {
			continue
		}
		hostIDf, ok := h["id"].(float64)
		if !ok || hostIDf <= 0 {
			continue
		}
		hostID := int64(hostIDf)

		for _, rr := range rooms {
			r, ok := rr.(map[string]any)
			if !ok {
				continue
			}
			roomIDf, ok := r["id"].(float64)
			if !ok || roomIDf <= 0 {
				continue
			}
			roomID := int64(roomIDf)

			q := "/scheduling/slots?hostId=" + strconv.FormatInt(hostID, 10) +
				"&roomId=" + strconv.FormatInt(roomID, 10) +
				"&day=" + dayStr +
				"&duration=" + strconv.Itoa(duration)
			resSlots, bodySlots := call(http.MethodGet, q, token, nil, t)
			if resSlots.StatusCode != http.StatusOK {
				continue
			}
			slots, ok := bodySlots["items"].([]any)
			if !ok || len(slots) == 0 {
				continue
			}
			first, ok := slots[0].(map[string]any)
			if !ok {
				continue
			}
			rawStart, ok := first["slotStart"].(string)
			if !ok || rawStart == "" {
				continue
			}
			slotStart, err := time.Parse(time.RFC3339, rawStart)
			if err != nil {
				continue
			}
			return hostID, roomID, slotStart, true
		}
	}

	return 0, 0, time.Time{}, false
}

func pickAvailableSchedulingSlot(t *testing.T, token string, day time.Time, duration int) (int64, int64, time.Time) {
	t.Helper()
	hostID, roomID, slotStart, found := findAvailableSchedulingSlot(t, token, day, duration)
	if found {
		return hostID, roomID, slotStart
	}
	t.Fatalf("no available scheduling slot found for day=%s duration=%d", day.UTC().Format("2006-01-02"), duration)
	return 0, 0, time.Time{}
}

func TestSuccessfulBooking(t *testing.T) {
	token := makeUserToken(t)

	// Add required address
	res, body := call(http.MethodPost, "/profile/addresses", token, map[string]any{
		"line1":      "123 Test St",
		"line2":      "Apt 4B",
		"city":       "Test City",
		"state":      "TS",
		"postalCode": "10001",
	}, t)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("address creation failed: %d %+v", res.StatusCode, body)
	}

	packageID, slot := pickBookablePackageAndSlot(t)
	res, body = call(http.MethodPost, "/bookings/holds", token, map[string]any{
		"packageId": packageID,
		"hostId":    3,
		"roomId":    1,
		"slotStart": slot.Format(time.RFC3339),
		"duration":  45,
	}, t)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected booking created, got %d: %+v", res.StatusCode, body)
	}

	holdId, ok := body["holdId"].(float64)
	if !ok || holdId <= 0 {
		t.Fatal("expected valid holdId in response")
	}
	version, ok := body["version"].(float64)
	if !ok || version != 1 {
		t.Fatal("expected version=1 in response")
	}
	status, ok := body["status"].(string)
	if !ok || status != "active" {
		t.Fatal("expected status=active in response")
	}
}

func TestBookingConflict(t *testing.T) {
	token := makeUserToken(t)
	res, body := call(http.MethodPost, "/profile/addresses", token, map[string]any{
		"line1":      "123 Test St",
		"line2":      "Apt 4B",
		"city":       "Test City",
		"state":      "TS",
		"postalCode": "10001",
	}, t)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("address creation failed: %d %+v", res.StatusCode, body)
	}
	packageID, hostID, roomID, slot := pickBookablePackageAndAvailableSchedulingSlot(t, token, 60)

	res1, _ := call(http.MethodPost, "/bookings/holds", token, map[string]any{
		"packageId": packageID,
		"hostId":    hostID,
		"roomId":    roomID,
		"slotStart": slot.Format(time.RFC3339),
		"duration":  60,
	}, t)
	if res1.StatusCode != http.StatusCreated {
		t.Fatalf("expected first booking created, got %d", res1.StatusCode)
	}

	res2, body2 := call(http.MethodPost, "/bookings/holds", token, map[string]any{
		"packageId": packageID,
		"hostId":    hostID,
		"roomId":    roomID,
		"slotStart": slot.Add(15 * time.Minute).Format(time.RFC3339),
		"duration":  45,
	}, t)
	if res2.StatusCode != http.StatusConflict {
		t.Fatalf("expected conflict (409), got %d: %+v", res2.StatusCode, body2)
	}
}

func TestConcurrentBookingConflict(t *testing.T) {
	tokenA := makeUserToken(t)
	tokenB := makeUserToken(t)
	packageID, hostID, roomID, slot := pickBookablePackageAndAvailableSchedulingSlot(t, tokenA, 45)

	addAddress := func(token string) {
		res, body := call(http.MethodPost, "/profile/addresses", token, map[string]any{
			"line1":      "456 Test Ave",
			"line2":      "",
			"city":       "Test City",
			"state":      "TS",
			"postalCode": "10001",
		}, t)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("address creation failed: %d %+v", res.StatusCode, body)
		}
	}
	addAddress(tokenA)
	addAddress(tokenB)
	payload := map[string]any{
		"packageId": packageID,
		"hostId":    hostID,
		"roomId":    roomID,
		"slotStart": slot.Format(time.RFC3339),
		"duration":  45,
	}

	results := make(chan int, 2)
	var wg sync.WaitGroup
	fire := func(token string) {
		defer wg.Done()
		body, _ := json.Marshal(payload)
		req, err := http.NewRequest(http.MethodPost, apiBase()+"/bookings/holds", bytes.NewReader(body))
		if err != nil {
			results <- 0
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		res, err := testClient().Do(req)
		if err != nil {
			results <- 0
			return
		}
		defer res.Body.Close()
		results <- res.StatusCode
	}

	wg.Add(2)
	go fire(tokenA)
	go fire(tokenB)
	wg.Wait()
	close(results)

	created := 0
	conflicts := 0
	for code := range results {
		if code == http.StatusCreated {
			created++
		}
		if code == http.StatusConflict {
			conflicts++
		}
	}
	if created != 1 || conflicts != 1 {
		t.Fatalf("expected one create and one conflict, got created=%d conflict=%d", created, conflicts)
	}
}

func TestCommunityAndNotificationsFlow(t *testing.T) {
	token := makeUserToken(t)

	resPost, bodyPost := call(http.MethodPost, "/community/posts", token, map[string]any{
		"title": "Trip question",
		"body":  "Is arrival after 6PM possible?",
	}, t)
	if resPost.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 post, got %d body=%+v", resPost.StatusCode, bodyPost)
	}
	postID, ok := bodyPost["id"].(float64)
	if !ok {
		t.Fatalf("missing post id")
	}

	resComment, _ := call(http.MethodPost, "/community/posts/"+strconv.Itoa(int(postID))+"/comments", token, map[string]any{
		"body": "Following this thread.",
	}, t)
	if resComment.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 comment, got %d", resComment.StatusCode)
	}

	resNotif, _ := call(http.MethodGet, "/notifications", token, nil, t)
	if resNotif.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 notifications, got %d", resNotif.StatusCode)
	}

	resLike, _ := call(http.MethodPost, "/community/likes", token, map[string]any{
		"targetType": "post",
		"targetId":   int(postID),
	}, t)
	if resLike.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 like, got %d", resLike.StatusCode)
	}
}

func TestReportNotificationFlow(t *testing.T) {
	// reporter creates a post then files a report; admin resolves it; reporter should get a notification
	reporter := makeUserToken(t)
	// create a post
	resPost, bodyPost := call(http.MethodPost, "/community/posts", reporter, map[string]any{"title": "Reportable", "body": "Please report me"}, t)
	if resPost.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 post, got %d body=%+v", resPost.StatusCode, bodyPost)
	}
	postID := int(bodyPost["id"].(float64))

	// reporter files a report against the post
	resReport, bodyReport := call(http.MethodPost, "/community/reports", reporter, map[string]any{"targetType": "post", "targetId": postID, "reason": "spam"}, t)
	if resReport.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 report, got %d body=%+v", resReport.StatusCode, bodyReport)
	}
	reportID := int(bodyReport["id"].(float64))

	// admin resolves the report
	admin := makeAdminToken(t)
	resResolve, bodyResolve := call(http.MethodPost, "/admin/reports/"+strconv.Itoa(reportID)+"/resolve", admin, map[string]any{"status": "closed", "outcome": "no action"}, t)
	if resResolve.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 resolve, got %d body=%+v", resResolve.StatusCode, bodyResolve)
	}

	// reporter should see notification
	resNotif, bodyNotif := call(http.MethodGet, "/notifications", reporter, nil, t)
	if resNotif.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 notifications, got %d body=%+v", resNotif.StatusCode, bodyNotif)
	}
	items, _ := bodyNotif["items"].([]any)
	if len(items) == 0 {
		t.Fatalf("expected at least one notification for reporter after report resolve")
	}
}

func TestAnalyticsAndEmailOpsEndpoints(t *testing.T) {
	token := makeAdminToken(t)
	today := time.Now().UTC().Format("2006-01-02")

	resKPI, _ := call(http.MethodGet, "/ops/analytics/kpis?from="+today+"&to="+today, token, nil, t)
	if resKPI.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 kpis, got %d", resKPI.StatusCode)
	}

	resQueue, _ := call(http.MethodPost, "/ops/email/queue", token, map[string]any{
		"templateKey":    "booking_confirmation",
		"recipientLabel": "demo@example.com",
		"subject":        "subj",
		"body":           "body",
	}, t)
	if resQueue.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 queue, got %d", resQueue.StatusCode)
	}

	resExport, _ := call(http.MethodPost, "/ops/email/export", token, map[string]any{}, t)
	if resExport.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 export, got %d", resExport.StatusCode)
	}
}

func TestScheduledReportLifecycle(t *testing.T) {
	token := makeAdminToken(t)
	now := time.Now().UTC().Add(-1 * time.Second).Format(time.RFC3339)
	resSchedule, bodySchedule := call(http.MethodPost, "/ops/reports/schedule", token, map[string]any{
		"reportType":   "kpi_daily",
		"parameters":   map[string]any{"from": time.Now().UTC().Format("2006-01-02")},
		"scheduledFor": now,
	}, t)
	if resSchedule.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 schedule, got %d body=%+v", resSchedule.StatusCode, bodySchedule)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resJobs, bodyJobs := call(http.MethodGet, "/ops/reports", token, nil, t)
		if resJobs.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 reports list, got %d", resJobs.StatusCode)
		}
		items, _ := bodyJobs["items"].([]any)
		for _, raw := range items {
			job, _ := raw.(map[string]any)
			if job["reportType"] == "kpi_daily" && job["status"] == "completed" && job["outputPath"] != "" {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("scheduled report job did not complete in time")
}

func TestSchedulingExceptionDayBlocksSlots(t *testing.T) {
	token := makeAdminToken(t)
	resUsers, bodyUsers := call(http.MethodGet, "/admin/users?role=coach", token, nil, t)
	if resUsers.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 users, got %d", resUsers.StatusCode)
	}
	items, _ := bodyUsers["items"].([]any)
	if len(items) == 0 {
		t.Fatalf("expected coach user in seed data")
	}
	coach := items[0].(map[string]any)
	coachID := int(coach["id"].(float64))
	day := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")
	resSlots, bodySlots := call(http.MethodGet, "/scheduling/slots?hostId="+strconv.Itoa(coachID)+"&roomId=1&day="+day+"&duration=30", token, nil, t)
	if resSlots.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 slots, got %d body=%+v", resSlots.StatusCode, bodySlots)
	}
	slots, _ := bodySlots["items"].([]any)
	if len(slots) != 0 {
		t.Fatalf("expected no slots on exception day, got %d", len(slots))
	}
}

func TestAddressMaskingPresent(t *testing.T) {
	token := makeUserToken(t)
	original := "123 Main Street"
	_, _ = call(http.MethodPost, "/profile/addresses", token, map[string]any{
		"line1":      original,
		"line2":      "Apt 2",
		"city":       "New York",
		"state":      "NY",
		"postalCode": "10001",
	}, t)
	res, body := call(http.MethodGet, "/profile/addresses", token, nil, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	items, ok := body["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected non-empty items")
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("bad item type")
	}
	if _, ok := first["line1Masked"]; !ok {
		t.Fatalf("expected line1Masked")
	}
	if v, ok := first["line1"].(string); ok && v == original {
		t.Fatalf("expected line1 to be masked in storage")
	}
}

func TestProfileContactsFlow(t *testing.T) {
	token := makeUserToken(t)
	resAdd, bodyAdd := call(http.MethodPost, "/profile/contacts", token, map[string]any{
		"name":         "Alex Parker",
		"relationship": "Emergency",
		"phone":        "+15551234567",
	}, t)
	if resAdd.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 contact, got %d body=%+v", resAdd.StatusCode, bodyAdd)
	}
	contactID, ok := bodyAdd["id"].(float64)
	if !ok {
		t.Fatalf("missing contact id")
	}

	resList, bodyList := call(http.MethodGet, "/profile/contacts", token, nil, t)
	if resList.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 contacts list, got %d", resList.StatusCode)
	}
	items, ok := bodyList["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected non-empty contacts list")
	}
	if first, ok := items[0].(map[string]any); ok {
		if _, ok := first["phoneMasked"]; !ok {
			t.Fatalf("expected phoneMasked in contact list")
		}
	}

	resDel, _ := call(http.MethodDelete, "/profile/contacts/"+strconv.Itoa(int(contactID)), token, nil, t)
	if resDel.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 contact delete, got %d", resDel.StatusCode)
	}
}

func TestContactsOwnershipIsolation(t *testing.T) {
	owner := makeUserToken(t)
	resAdd, bodyAdd := call(http.MethodPost, "/profile/contacts", owner, map[string]any{
		"name":         "Jamie Lee",
		"relationship": "Billing",
		"phone":        "+15550009999",
	}, t)
	if resAdd.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 contact, got %d body=%+v", resAdd.StatusCode, bodyAdd)
	}
	contactID, ok := bodyAdd["id"].(float64)
	if !ok {
		t.Fatalf("missing contact id")
	}

	other := makeUserToken(t)
	resDel, _ := call(http.MethodDelete, "/profile/contacts/"+strconv.Itoa(int(contactID)), other, nil, t)
	if resDel.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for non-owner delete, got %d", resDel.StatusCode)
	}
}
