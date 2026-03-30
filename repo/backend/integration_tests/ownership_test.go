package integration_tests

import "testing"

func TestOwnershipGuardBlocksOtherUser(t *testing.T) {
	token := makeUserToken(t)
	res, _ := call("GET", "/users/999999", token, nil, t)
	if res.StatusCode != 403 {
		t.Fatalf("expected 403 got %d", res.StatusCode)
	}
}
