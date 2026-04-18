package integration_tests

import (
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

func skipIfNotIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("BASE_URL") == "" && os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("integration tests skipped; set RUN_INTEGRATION_TESTS=true or BASE_URL to run")
	}
}

// userMeID returns the user ID for a given token via /auth/me.
func userMeID(t *testing.T, token string) int {
	t.Helper()
	res, body := call(http.MethodGet, "/auth/me", token, nil, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("auth/me failed: %d %+v", res.StatusCode, body)
	}
	idF, _ := body["id"].(float64)
	return int(idF)
}

// ─── Public / health ──────────────────────────────────────────────────────────

func TestHealthEndpoint(t *testing.T) {
	skipIfNotIntegration(t)
	res, err := testClient().Get(baseURL() + "/health")
	if err != nil {
		t.Fatalf("request /health: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /health, got %d", res.StatusCode)
	}
}

func TestPublicCatalogExtended(t *testing.T) {
	for _, path := range []string{
		"/catalog/routes",
		"/catalog/hotels",
		"/catalog/attractions",
		"/config/coverage",
	} {
		res, body := call(http.MethodGet, path, "", nil, t)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d body=%+v", path, res.StatusCode, body)
		}
		if _, ok := body["items"]; path != "/config/coverage" && !ok {
			t.Fatalf("expected items field in response for %s", path)
		}
	}
}

func TestPublicCatalogRoutesMeta(t *testing.T) {
	res, body := call(http.MethodGet, "/catalog/routes", "", nil, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 routes, got %d", res.StatusCode)
	}
	if _, ok := body["items"]; !ok {
		t.Fatalf("expected items in routes response")
	}
}

// ─── Auth / me ────────────────────────────────────────────────────────────────

func TestAuthMeSuccess(t *testing.T) {
	token := makeUserToken(t)
	res, body := call(http.MethodGet, "/auth/me", token, nil, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /auth/me, got %d body=%+v", res.StatusCode, body)
	}
	if _, ok := body["id"]; !ok {
		t.Fatalf("expected id field in /auth/me response")
	}
	if _, ok := body["username"]; !ok {
		t.Fatalf("expected username field in /auth/me response")
	}
}

func TestAuthMeRequiresToken(t *testing.T) {
	res, _ := call(http.MethodGet, "/auth/me", "", nil, t)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for /auth/me without token, got %d", res.StatusCode)
	}
}

// ─── Booking holds list & history ─────────────────────────────────────────────

func TestBookingHoldsListAndHistory(t *testing.T) {
	token := makeUserToken(t)

	resHolds, bodyHolds := call(http.MethodGet, "/bookings/holds", token, nil, t)
	if resHolds.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /bookings/holds, got %d body=%+v", resHolds.StatusCode, bodyHolds)
	}
	if _, ok := bodyHolds["items"]; !ok {
		t.Fatalf("expected items field in /bookings/holds response")
	}

	resHist, bodyHist := call(http.MethodGet, "/bookings/history", token, nil, t)
	if resHist.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /bookings/history, got %d body=%+v", resHist.StatusCode, bodyHist)
	}
	if _, ok := bodyHist["items"]; !ok {
		t.Fatalf("expected items field in /bookings/history response")
	}
}

func TestBookingHoldsRequiresAuth(t *testing.T) {
	res, _ := call(http.MethodGet, "/bookings/holds", "", nil, t)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for /bookings/holds without token, got %d", res.StatusCode)
	}
}

// ─── Hold confirm & cancel ────────────────────────────────────────────────────

func TestBookingConfirmHold(t *testing.T) {
	token := makeUserToken(t)

	res, body := call(http.MethodPost, "/profile/addresses", token, map[string]any{
		"line1": "99 Confirm Lane", "line2": "", "city": "Testville", "state": "TV", "postalCode": "10001",
	}, t)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("address setup failed: %d %+v", res.StatusCode, body)
	}

	packageID, hostID, roomID, slot := pickBookablePackageAndAvailableSchedulingSlot(t, token, 45)
	resHold, bodyHold := call(http.MethodPost, "/bookings/holds", token, map[string]any{
		"packageId": packageID,
		"hostId":    hostID,
		"roomId":    roomID,
		"slotStart": slot.Format(time.RFC3339),
		"duration":  45,
	}, t)
	if resHold.StatusCode != http.StatusCreated {
		t.Fatalf("hold placement failed: %d %+v", resHold.StatusCode, bodyHold)
	}
	holdID := int(bodyHold["holdId"].(float64))
	version := int(bodyHold["version"].(float64))

	resConfirm, bodyConfirm := call(http.MethodPost, "/bookings/confirm", token, map[string]any{
		"holdId":  holdID,
		"version": version,
	}, t)
	if resConfirm.StatusCode != http.StatusOK && resConfirm.StatusCode != http.StatusCreated {
		t.Fatalf("expected 200/201 confirm, got %d body=%+v", resConfirm.StatusCode, bodyConfirm)
	}
}

func TestCancelHold(t *testing.T) {
	token := makeUserToken(t)

	res, body := call(http.MethodPost, "/profile/addresses", token, map[string]any{
		"line1": "88 Cancel St", "line2": "", "city": "Testville", "state": "TV", "postalCode": "10001",
	}, t)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("address setup failed: %d %+v", res.StatusCode, body)
	}

	packageID, hostID, roomID, slot := pickBookablePackageAndAvailableSchedulingSlot(t, token, 30)
	resHold, bodyHold := call(http.MethodPost, "/bookings/holds", token, map[string]any{
		"packageId": packageID,
		"hostId":    hostID,
		"roomId":    roomID,
		"slotStart": slot.Format(time.RFC3339),
		"duration":  30,
	}, t)
	if resHold.StatusCode != http.StatusCreated {
		t.Fatalf("hold placement failed: %d %+v", resHold.StatusCode, bodyHold)
	}
	holdID := int(bodyHold["holdId"].(float64))

	resCancel, bodyCancel := call(http.MethodDelete, "/bookings/holds/"+strconv.Itoa(holdID), token, nil, t)
	if resCancel.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 cancel, got %d body=%+v", resCancel.StatusCode, bodyCancel)
	}
}

// ─── Profile address delete ────────────────────────────────────────────────────

func TestDeleteAddress(t *testing.T) {
	token := makeUserToken(t)

	resAdd, bodyAdd := call(http.MethodPost, "/profile/addresses", token, map[string]any{
		"line1": "77 Delete Ave", "line2": "", "city": "Testville", "state": "TV", "postalCode": "10001",
	}, t)
	if resAdd.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 address, got %d body=%+v", resAdd.StatusCode, bodyAdd)
	}
	addrID, ok := bodyAdd["id"].(float64)
	if !ok {
		t.Fatalf("expected id in address response")
	}

	resDel, bodyDel := call(http.MethodDelete, "/profile/addresses/"+strconv.Itoa(int(addrID)), token, nil, t)
	if resDel.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 delete address, got %d body=%+v", resDel.StatusCode, bodyDel)
	}
}

// ─── Notification mark-read ────────────────────────────────────────────────────

func TestNotificationMarkRead(t *testing.T) {
	token := makeUserToken(t)

	// Create a post so the user has at least one notification path
	resPost, bodyPost := call(http.MethodPost, "/community/posts", token, map[string]any{
		"title": "Notify test post",
		"body":  "Testing notification mark-read.",
	}, t)
	if resPost.StatusCode != http.StatusCreated {
		t.Fatalf("post creation failed: %d %+v", resPost.StatusCode, bodyPost)
	}

	resNotif, bodyNotif := call(http.MethodGet, "/notifications", token, nil, t)
	if resNotif.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 notifications, got %d", resNotif.StatusCode)
	}
	items, _ := bodyNotif["items"].([]any)
	if len(items) == 0 {
		t.Skip("no notifications present to mark read")
	}
	first, _ := items[0].(map[string]any)
	notifID := int(first["id"].(float64))

	resMark, bodyMark := call(http.MethodPost, "/notifications/"+strconv.Itoa(notifID)+"/read", token, nil, t)
	if resMark.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 mark-read, got %d body=%+v", resMark.StatusCode, bodyMark)
	}
}

// ─── Scheduling agendas & chairs ──────────────────────────────────────────────

func TestSchedulingAgendasAndChairs(t *testing.T) {
	admin := makeAdminToken(t)

	resHosts, bodyHosts := call(http.MethodGet, "/scheduling/hosts", admin, nil, t)
	if resHosts.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 hosts, got %d", resHosts.StatusCode)
	}
	hosts, _ := bodyHosts["items"].([]any)
	if len(hosts) == 0 {
		t.Skip("no hosts in seed data")
	}
	hostID := int(hosts[0].(map[string]any)["id"].(float64))

	resRooms, bodyRooms := call(http.MethodGet, "/scheduling/rooms", admin, nil, t)
	if resRooms.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 rooms, got %d", resRooms.StatusCode)
	}
	rooms, _ := bodyRooms["items"].([]any)
	if len(rooms) == 0 {
		t.Skip("no rooms in seed data")
	}
	roomID := int(rooms[0].(map[string]any)["id"].(float64))

	// Host agenda
	resHostAgenda, bodyHostAgenda := call(http.MethodGet, "/scheduling/hosts/"+strconv.Itoa(hostID)+"/agenda", admin, nil, t)
	if resHostAgenda.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 host agenda, got %d body=%+v", resHostAgenda.StatusCode, bodyHostAgenda)
	}
	if _, ok := bodyHostAgenda["items"]; !ok {
		t.Fatalf("expected items in host agenda")
	}

	// Room agenda
	resRoomAgenda, bodyRoomAgenda := call(http.MethodGet, "/scheduling/rooms/"+strconv.Itoa(roomID)+"/agenda", admin, nil, t)
	if resRoomAgenda.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 room agenda, got %d body=%+v", resRoomAgenda.StatusCode, bodyRoomAgenda)
	}
	if _, ok := bodyRoomAgenda["items"]; !ok {
		t.Fatalf("expected items in room agenda")
	}

	// Room chairs
	resChairs, bodyChairs := call(http.MethodGet, "/scheduling/rooms/"+strconv.Itoa(roomID)+"/chairs", admin, nil, t)
	if resChairs.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 room chairs, got %d body=%+v", resChairs.StatusCode, bodyChairs)
	}
	if _, ok := bodyChairs["items"]; !ok {
		t.Fatalf("expected items in room chairs")
	}
}

func TestHostAgendaForbiddenForOtherUser(t *testing.T) {
	token := makeUserToken(t)
	// Requesting another user's agenda as a non-privileged user should be forbidden.
	res, _ := call(http.MethodGet, "/scheduling/hosts/9999/agenda", token, nil, t)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 host agenda cross-user, got %d", res.StatusCode)
	}
}

// ─── Community read endpoints ──────────────────────────────────────────────────

func TestCommunityPostsAndComments(t *testing.T) {
	token := makeUserToken(t)

	// Create a post to ensure there's at least one
	resCreate, bodyCreate := call(http.MethodPost, "/community/posts", token, map[string]any{
		"title": "Extended test post",
		"body":  "Testing community read endpoints.",
	}, t)
	if resCreate.StatusCode != http.StatusCreated {
		t.Fatalf("post creation failed: %d %+v", resCreate.StatusCode, bodyCreate)
	}
	postID := int(bodyCreate["id"].(float64))

	// List posts
	resList, bodyList := call(http.MethodGet, "/community/posts", token, nil, t)
	if resList.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 community posts, got %d body=%+v", resList.StatusCode, bodyList)
	}
	if _, ok := bodyList["items"]; !ok {
		t.Fatalf("expected items in community posts")
	}

	// Add a comment
	resComment, bodyComment := call(http.MethodPost, "/community/posts/"+strconv.Itoa(postID)+"/comments", token, map[string]any{
		"body": "Test comment",
	}, t)
	if resComment.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 comment, got %d body=%+v", resComment.StatusCode, bodyComment)
	}

	// List comments
	resComments, bodyComments := call(http.MethodGet, "/community/posts/"+strconv.Itoa(postID)+"/comments", token, nil, t)
	if resComments.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 comments, got %d body=%+v", resComments.StatusCode, bodyComments)
	}
	items, ok := bodyComments["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected non-empty items in comments")
	}
}

// ─── Community social actions ─────────────────────────────────────────────────

func TestCommunityFavoritesFollowsBlocks(t *testing.T) {
	actor := makeUserToken(t)
	other := makeUserToken(t)
	otherID := userMeID(t, other)

	// Favorite a package from the catalog
	res, body := call(http.MethodGet, "/catalog", "", nil, t)
	if res.StatusCode != http.StatusOK || len(body) == 0 {
		t.Fatalf("catalog unavailable: %d", res.StatusCode)
	}
	items, _ := body["items"].([]any)
	if len(items) > 0 {
		pkgID := int(items[0].(map[string]any)["id"].(float64))
		resFav, bodyFav := call(http.MethodPost, "/community/favorites", actor, map[string]any{
			"packageId": pkgID,
		}, t)
		if resFav.StatusCode != http.StatusOK && resFav.StatusCode != http.StatusCreated {
			t.Fatalf("expected 200/201 favorite, got %d body=%+v", resFav.StatusCode, bodyFav)
		}
	}

	// Follow the other user
	resFollow, bodyFollow := call(http.MethodPost, "/community/follows", actor, map[string]any{
		"userId": otherID,
	}, t)
	if resFollow.StatusCode != http.StatusOK && resFollow.StatusCode != http.StatusCreated {
		t.Fatalf("expected 200/201 follow, got %d body=%+v", resFollow.StatusCode, bodyFollow)
	}

	// Block the other user
	resBlock, bodyBlock := call(http.MethodPost, "/community/blocks", actor, map[string]any{
		"userId": otherID,
	}, t)
	if resBlock.StatusCode != http.StatusOK && resBlock.StatusCode != http.StatusCreated {
		t.Fatalf("expected 200/201 block, got %d body=%+v", resBlock.StatusCode, bodyBlock)
	}
}

// ─── Admin role operations ────────────────────────────────────────────────────

func TestAdminRoleAssignAndAuditList(t *testing.T) {
	admin := makeAdminToken(t)

	// Create a user to assign a role to
	target := makeUserToken(t)
	targetID := userMeID(t, target)

	resAssign, bodyAssign := call(http.MethodPost, "/admin/roles/assign", admin, map[string]any{
		"targetUserId": targetID,
		"role":         "traveler",
	}, t)
	if resAssign.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 role assign, got %d body=%+v", resAssign.StatusCode, bodyAssign)
	}
	if status, _ := bodyAssign["status"].(string); status != "ok" {
		t.Fatalf("expected status=ok in role assign response, got %+v", bodyAssign)
	}

	// List role audits
	resAudits, bodyAudits := call(http.MethodGet, "/admin/roles/audits", admin, nil, t)
	if resAudits.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 role audits, got %d body=%+v", resAudits.StatusCode, bodyAudits)
	}
	if _, ok := bodyAudits["items"]; !ok {
		t.Fatalf("expected items in role audits response")
	}
}

func TestAdminRoleAssignBlockedForNonAdmin(t *testing.T) {
	token := makeUserToken(t)
	res, _ := call(http.MethodPost, "/admin/roles/assign", token, map[string]any{
		"targetUserId": 1,
		"role":         "admin",
	}, t)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 non-admin role assign, got %d", res.StatusCode)
	}
}

// ─── Admin regions & service rules ────────────────────────────────────────────

func TestAdminRegionsAndServiceRules(t *testing.T) {
	admin := makeAdminToken(t)
	regionName := "TestRegion_" + strconv.FormatInt(int64(time.Now().UnixNano()), 36)

	// Create region
	resCreate, bodyCreate := call(http.MethodPost, "/admin/regions", admin, map[string]any{
		"name":        regionName,
		"description": "Integration test region",
	}, t)
	if resCreate.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 region create, got %d body=%+v", resCreate.StatusCode, bodyCreate)
	}
	regionID, ok := bodyCreate["id"].(float64)
	if !ok || regionID <= 0 {
		t.Fatalf("expected region id in response, got %+v", bodyCreate)
	}

	// List regions
	resList, bodyList := call(http.MethodGet, "/admin/regions", admin, nil, t)
	if resList.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 regions list, got %d body=%+v", resList.StatusCode, bodyList)
	}
	if _, ok := bodyList["items"]; !ok {
		t.Fatalf("expected items in regions list")
	}

	// Create service rule for the region
	resRule, bodyRule := call(http.MethodPost, "/admin/regions/"+strconv.Itoa(int(regionID))+"/service-rule", admin, map[string]any{
		"allowHomePickup":    true,
		"allowMailDocuments": false,
		"blocked":            false,
		"startTime":          "09:00",
		"endTime":            "17:00",
	}, t)
	if resRule.StatusCode != http.StatusOK && resRule.StatusCode != http.StatusCreated {
		t.Fatalf("expected 200/201 service rule, got %d body=%+v", resRule.StatusCode, bodyRule)
	}
	serviceRuleID, ok := bodyRule["serviceRuleId"].(float64)
	if !ok || serviceRuleID <= 0 {
		t.Fatalf("expected serviceRuleId in service rule response, got %+v", bodyRule)
	}

	// List blocked postal codes
	resBPC, bodyBPC := call(http.MethodGet, "/admin/service/blocked-postal-codes", admin, nil, t)
	if resBPC.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 blocked postal codes, got %d body=%+v", resBPC.StatusCode, bodyBPC)
	}

	// Add blocked postal code
	resAdd, bodyAdd := call(http.MethodPost, "/admin/service/blocked-postal-codes", admin, map[string]any{
		"serviceRuleId": int(serviceRuleID),
		"postalCode":    "99999",
	}, t)
	if resAdd.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 blocked postal code, got %d body=%+v", resAdd.StatusCode, bodyAdd)
	}
}

// ─── Admin catalog publish ────────────────────────────────────────────────────

func TestAdminCatalogPublish(t *testing.T) {
	admin := makeAdminToken(t)

	res, body := call(http.MethodGet, "/catalog", "", nil, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("catalog unavailable: %d", res.StatusCode)
	}
	items, _ := body["items"].([]any)
	if len(items) == 0 {
		t.Skip("no catalog items to publish")
	}
	pkgID := int(items[0].(map[string]any)["id"].(float64))

	resPublish, bodyPublish := call(http.MethodPost, "/admin/catalog/packages/"+strconv.Itoa(pkgID)+"/publish", admin, map[string]any{
		"published": true,
	}, t)
	if resPublish.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 catalog publish, got %d body=%+v", resPublish.StatusCode, bodyPublish)
	}
	if status, _ := bodyPublish["status"].(string); status != "ok" {
		t.Fatalf("expected status=ok in publish response, got %+v", bodyPublish)
	}
}

// ─── Ops: analytics export & email queue list ──────────────────────────────────

func TestOpsAnalyticsExport(t *testing.T) {
	admin := makeAdminToken(t)
	today := time.Now().UTC().Format("2006-01-02")

	res, body := call(http.MethodGet, "/ops/analytics/export?from="+today+"&to="+today, admin, nil, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 analytics export, got %d body=%+v", res.StatusCode, body)
	}
}

func TestOpsEmailQueueList(t *testing.T) {
	admin := makeAdminToken(t)

	res, body := call(http.MethodGet, "/ops/email/queue", admin, nil, t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 email queue list, got %d body=%+v", res.StatusCode, body)
	}
	if _, ok := body["items"]; !ok {
		t.Fatalf("expected items in email queue response")
	}
}

// ─── RBAC: multiple protected endpoints reject unauthenticated requests ────────

func TestUnauthenticatedEndpointsReturn401(t *testing.T) {
	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/auth/me"},
		{http.MethodGet, "/bookings/holds"},
		{http.MethodGet, "/bookings/history"},
		{http.MethodPost, "/bookings/holds"},
		{http.MethodPost, "/bookings/confirm"},
		{http.MethodGet, "/community/posts"},
		{http.MethodGet, "/notifications"},
		{http.MethodGet, "/scheduling/hosts"},
		{http.MethodGet, "/scheduling/rooms"},
		{http.MethodGet, "/ops/analytics/kpis"},
		{http.MethodGet, "/admin/users"},
	}

	for _, ep := range endpoints {
		res, _ := call(ep.method, ep.path, "", nil, t)
		if res.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 for %s %s without token, got %d", ep.method, ep.path, res.StatusCode)
		}
	}
}

func TestAdminEndpointsReturn403ForNonAdmin(t *testing.T) {
	token := makeUserToken(t)
	adminEndpoints := []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodGet, "/admin/users", nil},
		{http.MethodGet, "/admin/roles/audits", nil},
		{http.MethodGet, "/admin/regions", nil},
		{http.MethodGet, "/admin/service/blocked-postal-codes", nil},
		{http.MethodGet, "/ops/analytics/kpis?from=2025-01-01&to=2025-01-01", nil},
		{http.MethodGet, "/ops/email/queue", nil},
	}

	for _, ep := range adminEndpoints {
		res, _ := call(ep.method, ep.path, token, ep.body, t)
		if res.StatusCode != http.StatusForbidden {
			t.Errorf("expected 403 for %s %s with traveler token, got %d", ep.method, ep.path, res.StatusCode)
		}
	}
}
