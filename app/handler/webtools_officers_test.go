package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
)

func addAuth(t *testing.T, memberId int, sub string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO auth (member_id, sub, idp, idp_hash) VALUES (?, ?, 'GOOGLE', ?)`, memberId, sub, sub)
	if err != nil {
		t.Fatal(err)
	}
}

func addOfficer(t *testing.T, memberId int, position string, security int) {
	t.Helper()
	_, err := db.Exec(`
		INSERT INTO officer (member_id, create_datetime, expire_datetime, position, security)
		VALUES (?, datetime('now'), datetime('now', '+10 years'), ?, ?)`, memberId, position, security)
	if err != nil {
		t.Fatal(err)
	}
}

func deleteOfficerAs(sub string, officerId string) *httptest.ResponseRecorder {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("memberId", officerId)
	req := httptest.NewRequest("DELETE", "/", nil)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, "sub", sub)
	w := httptest.NewRecorder()
	DeleteWebtoolsOfficers(w, req.WithContext(ctx))
	return w
}

func isOfficerRow(t *testing.T, memberId int) bool {
	t.Helper()
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM officer WHERE member_id = ?)`, memberId).Scan(&exists)
	if err != nil {
		t.Fatal(err)
	}
	return exists
}

/*
 * The founding admin (8000001) can remove an officer holding equal security;
 * this is the case that was stuck — an officer accidentally added at security
 * 100, the admin's own level, that the security guard refused to let anyone
 * touch.
 */
func TestDeleteOfficerSuperAdminBypass(t *testing.T) {
	openTestDB(t)
	addAuth(t, SUPER_ADMIN_MEMBER_ID, "superadmin")
	addOfficer(t, 8000002, "Stuck At 100", 100)

	w := deleteOfficerAs("superadmin", "8000002")
	if w.Code != http.StatusNoContent {
		t.Fatalf("super admin delete: got %d %s, want 204", w.Code, w.Body.String())
	}
	if isOfficerRow(t, 8000002) {
		t.Error("officer row still present after super admin delete")
	}
}

/* A non-super-admin officer is still capped by security. */
func TestDeleteOfficerRegularStillCapped(t *testing.T) {
	openTestDB(t)
	addAuth(t, 8000002, "regular")
	addOfficer(t, 8000002, "Regular Officer", 90)
	addOfficer(t, 8000001, "Higher Up", 100)

	w := deleteOfficerAs("regular", "8000001")
	if w.Code != http.StatusForbidden {
		t.Fatalf("regular deleting higher security: got %d, want 403", w.Code)
	}
	if !isOfficerRow(t, 8000001) {
		t.Error("higher-security officer row was wrongly removed")
	}
}

/* Nobody, super admin included, removes themselves and orphans the club. */
func TestDeleteOfficerSelfBlocked(t *testing.T) {
	openTestDB(t)
	addAuth(t, SUPER_ADMIN_MEMBER_ID, "superadmin")
	addOfficer(t, SUPER_ADMIN_MEMBER_ID, "Super Admin", 100)

	w := deleteOfficerAs("superadmin", "8000001")
	if w.Code != http.StatusForbidden {
		t.Fatalf("self-removal: got %d, want 403", w.Code)
	}
	if !isOfficerRow(t, SUPER_ADMIN_MEMBER_ID) {
		t.Error("super admin removed their own officer row")
	}
}
