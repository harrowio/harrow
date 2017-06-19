package http

import (
	"encoding/base32"
	"testing"
)

func Test_PasswordResetHandler_Status422_ForEmptyPassword(t *testing.T) {

	h := NewHandlerTest(MountPasswordResetHandler, t)
	defer h.Cleanup()

	user := h.World().User("without_password")

	hmac := base32.StdEncoding.EncodeToString(user.HMAC([]byte(c.HttpConfig().UserHmacSecret)))
	params := &struct {
		Email    string
		Mac      string
		Password string
	}{user.Email, hmac, ""}

	result := new(ErrorJSON)
	h.ResultTo(result)
	h.Do("POST", h.Url("/reset-password"), &params)
	if have, want := h.Response().StatusCode, 422; have != want {
		t.Errorf("h.Response().StatusCode have=%d; want=%d", have, want)
	}
	if have, want := result.Reason, "invalid"; have != want {
		t.Errorf("result.Reason have=%q; want=%q", have, want)
	}
	_, ok := result.Errors["password"]
	if have, want := ok, true; have != want {
		t.Fatalf("result.Errors[\"password\"] should exist", have, want)
	}
	if have, want := len(result.Errors["password"]), 1; have != want {
		t.Fatalf("len(result.Errors[\"password\"]) have=%d; want=%d", have, want)
	}
	if have, want := result.Errors["password"][0], "invalid"; have != want {
		t.Errorf("len(result.Errors[\"password\"][0]) have=%s; want=%s", have, want)
	}

}
