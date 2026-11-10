package server

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoggedURI_HidesShareAndLoginTokens(t *testing.T) {
	for target, want := range map[string]string{
		"/d/AbC-_123":                        "/d/REDACTED",
		"/server/d/AbC-_123":                 "/server/d/REDACTED",
		"http://nas.example/d/AbC-_123":      "/d/REDACTED",
		"/D/AbC-_123":                        "/d/REDACTED",
		"/img?loc=%2Fphotos&token=jwt.value": "/img?loc=%2Fphotos&token=REDACTED",
		"/img?%74oken=jwt.value&name=a.png":  "/img?name=a.png&token=REDACTED",
		"/searchInAllFiles?query=d%2Fsecret": "/searchInAllFiles?query=d%2Fsecret",
		"/getActivityLog":                    "/getActivityLog",
	} {
		require.Equal(t, want, loggedURI(httptest.NewRequest("GET", target, nil)), target)
	}
}
