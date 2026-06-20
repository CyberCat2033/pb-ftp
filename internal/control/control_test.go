package control

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleUpdateIgnoresUnknownFields(t *testing.T) {
	var received UpdateRequest
	server := &Server{
		updateHandler: func(request UpdateRequest) error {
			received = request
			return nil
		},
	}

	body := `{
		"sourcePath": "/mnt/ext1/applications/.pb-ftp-update/pb-ftp.app",
		"versionName": "1.0.4",
		"versionCode": 23,
		"releasedAt": "2026-06-20T00:00:00Z",
		"sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"futureField": "ignored"
	}`
	request := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(body))
	recorder := httptest.NewRecorder()

	server.handleUpdate(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d; body=%q", recorder.Code, http.StatusAccepted, recorder.Body.String())
	}
	if received.VersionName != "1.0.4" || received.VersionCode != 23 {
		t.Fatalf("received update = %+v", received)
	}
}
