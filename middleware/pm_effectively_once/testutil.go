package pm_effectively_once

import (
	"testing"

	"github.com/rs/xid"
)

func randString(t *testing.T, length int) string {
	t.Helper()

	uuid := xid.New().String()
	for len(uuid) < length {
		uuid += xid.New().String()
	}
	return uuid[:length]
}
