//nolint:testpackage // test unexported helpers in cmd.
package cmd

import "testing"

// TestDecodeTokenResponseUserIDNumber accepts numeric user IDs.
func TestDecodeTokenResponseUserIDNumber(t *testing.T) {
	t.Parallel()

	payload := []byte(
		`{"status":0,"body":` +
			`{"access_token":"acc","refresh_token":"ref",` +
			`"expires_in":3600,"token_type":"Bearer",` +
			`"scope":"user.metrics","userid":12345}}`,
	)

	token, err := decodeTokenResponse(payload)
	if err != nil {
		t.Fatalf("decodeTokenResponse: %v", err)
	}

	if string(token.UserID) != "12345" {
		t.Fatalf("userid got %q want %q", token.UserID, "12345")
	}
}

// TestDecodeTokenResponseUserIDString accepts string user IDs.
func TestDecodeTokenResponseUserIDString(t *testing.T) {
	t.Parallel()

	payload := []byte(
		`{"status":0,"body":` +
			`{"access_token":"acc","refresh_token":"ref",` +
			`"expires_in":3600,"token_type":"Bearer",` +
			`"scope":"user.metrics","userid":"abc"}}`,
	)

	token, err := decodeTokenResponse(payload)
	if err != nil {
		t.Fatalf("decodeTokenResponse: %v", err)
	}

	if string(token.UserID) != "abc" {
		t.Fatalf("userid got %q want %q", token.UserID, "abc")
	}
}
