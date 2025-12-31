//nolint:testpackage // test unexported helpers in cmd.
package cmd

import "testing"

const (
	userTestBaseNoV2  = "https://wbsapi.withings.net"
	userTestBaseV2    = "https://wbsapi.withings.net/v2"
	userTestBaseV2Sl  = "https://wbsapi.withings.net/v2/"
	userTestID        = "123"
	userTestFirstName = "Jane"
	userTestLastName  = "Doe"
	userTestEmail     = "jane@example.com"
	userTestBirthdate = "1990-01-01"
	userTestGender    = "1"
	userTestTimezone  = "Europe/Berlin"
	userTestServiceFm = "service got %q want %q"
	userTestRowCount  = 1
)

// TestUserServiceForBase handles base URLs with and without /v2.
func TestUserServiceForBase(t *testing.T) {
	t.Parallel()

	got := userServiceForBase(userTestBaseNoV2)
	if got != userServiceName {
		t.Fatalf(userTestServiceFm, got, userServiceName)
	}

	got = userServiceForBase(userTestBaseV2)
	if got != userServiceShort {
		t.Fatalf(userTestServiceFm, got, userServiceShort)
	}

	got = userServiceForBase(userTestBaseV2Sl)
	if got != userServiceShort {
		t.Fatalf(userTestServiceFm, got, userServiceShort)
	}
}

// TestBuildUserRows maps user profile fields to rows.
func TestBuildUserRows(t *testing.T) {
	t.Parallel()

	profiles := []map[string]any{
		{
			"id":        userTestID,
			"firstname": userTestFirstName,
			"lastname":  userTestLastName,
			"email":     userTestEmail,
			"birthdate": userTestBirthdate,
			"gender":    userTestGender,
			"timezone":  userTestTimezone,
		},
	}

	rows := buildUserRows(profiles)
	if len(rows) != userTestRowCount {
		t.Fatalf("rows got %d want %d", len(rows), userTestRowCount)
	}

	row := rows[defaultInt]
	assertUserField(t, "id", row.ID, userTestID)
	assertUserField(t, "first name", row.FirstName, userTestFirstName)
	assertUserField(t, "last name", row.LastName, userTestLastName)
	assertUserField(t, "email", row.Email, userTestEmail)
	assertUserField(t, "birthdate", row.Birthdate, userTestBirthdate)
	assertUserField(t, "gender", row.Gender, userTestGender)
	assertUserField(t, "timezone", row.Timezone, userTestTimezone)
}

func assertUserField(t *testing.T, label, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s got %q want %q", label, got, want)
	}
}
