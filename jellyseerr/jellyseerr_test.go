package jellyseerr

import (
	"os"
	"testing"

	"github.com/hrfee/jfa-go/common"
)

const (
	PERM = 2097184
)

func client(t *testing.T) *Jellyseerr {
	t.Helper()
	uri := os.Getenv("JELLYSEERR_TEST_URI")
	key := os.Getenv("JELLYSEERR_TEST_API_KEY")
	if uri == "" || key == "" {
		t.Skip("set JELLYSEERR_TEST_URI and JELLYSEERR_TEST_API_KEY to run Jellyseerr integration tests")
	}
	return NewJellyseerr(uri, key, common.NewTimeoutHandler("Jellyseerr", uri, false))
}

func testJellyfinID(t *testing.T) string {
	t.Helper()
	id := os.Getenv("JELLYSEERR_TEST_JF_ID")
	if id == "" {
		t.Skip("set JELLYSEERR_TEST_JF_ID to run this Jellyseerr integration test")
	}
	return id
}

func TestMe(t *testing.T) {
	js := client(t)
	u, err := js.Me()
	if err != nil {
		t.Fatalf("returned error %+v", err)
	}
	if u.ID < 0 {
		t.Fatalf("returned no user %+v\n", u)
	}
}

/* func TestImportFromJellyfin(t *testing.T) {
	js := client()
	list, err := js.ImportFromJellyfin("6b75e189efb744f583aa2e8e9cee41d3")
	if err != nil {
		t.Fatalf("returned error %+v", err)
	}
	if len(list) == 0 {
		t.Fatalf("returned no users")
	}
} */

func TestMustGetUser(t *testing.T) {
	js := client(t)
	u, err := js.MustGetUser(testJellyfinID(t))
	if err != nil {
		t.Fatalf("returned error %+v", err)
	}
	if u.ID < 0 {
		t.Fatalf("returned no users")
	}
}

func TestSetPermissions(t *testing.T) {
	js := client(t)
	err := js.SetPermissions(testJellyfinID(t), PERM)
	if err != nil {
		t.Fatalf("returned error %+v", err)
	}
}

func TestGetPermissions(t *testing.T) {
	js := client(t)
	perm, err := js.GetPermissions(testJellyfinID(t))
	if err != nil {
		t.Fatalf("returned error %+v", err)
	}
	if perm != PERM {
		t.Fatalf("got unexpected perm code %d", perm)
	}
}
