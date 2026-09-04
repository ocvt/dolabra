package handler

import (
	"database/sql"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ocvt/dolabra/utils"
	"github.com/stripe/stripe-go/v71"
)

/*
 * Real schema in a shared-cache memory database. DBMigrate reads the schema
 * relative to the repo root, so the test runs from there.
 */
func openTestDB(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(wd) })

	db, err = sql.Open("sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_foreign_keys=1")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	utils.DBMigrate(db)

	for _, m := range []struct {
		id     int
		expire string
	}{{8000000, "2126-01-01 00:00:00"}, {8000001, "2025-09-01 19:51:21"}, {8000002, "2027-03-01 00:00:00"}} {
		_, err = db.Exec(`
			INSERT INTO member (id, email, name, create_datetime, cell_number, pronouns,
				birth_year, active, medical_cond, medical_cond_desc, paid_expire_datetime,
				ec_name, ec_number, ec_relationship, notification_preference)
			VALUES (?, 'm@example.com', 'Member', datetime('now'), '', 'they/them',
				2000, true, false, '', ?, '', '', '', '{}')`, m.id, m.expire)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func expiry(t *testing.T, memberId int) time.Time {
	t.Helper()
	// go-sqlite3 returns DATETIME columns as time.Time
	var exp time.Time
	err := db.QueryRow(`SELECT paid_expire_datetime FROM member WHERE id = ?`, memberId).Scan(&exp)
	if err != nil {
		t.Fatal(err)
	}
	return exp.UTC()
}

func completed(t *testing.T, sessionId string, storeItemId string) bool {
	t.Helper()
	var c bool
	err := db.QueryRow(`
		SELECT completed FROM payment
		WHERE payment_id = ? AND store_item_id = ?`, sessionId, storeItemId).Scan(&c)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

/*
 * A freshman special for a lapsed member: four years from today, granted once
 * no matter how many times Stripe (or the reconcile sweep) reports the session.
 * The shirt row stays open for pickup.
 */
func TestCompleteStripeSessionGrantsOnce(t *testing.T) {
	openTestDB(t)
	w := httptest.NewRecorder()
	const sid = "cs_test_freshman"
	dbInsertPayment(w, 8000000, "", 8000001, "MEMBERSHIP", 4, 6500, "STRIPE", sid, false)
	dbInsertPayment(w, 8000000, "", 8000001, "SHIRT", 1, 6500, "STRIPE", sid, false)

	applied, ok := completeStripeSession(w, sid)
	if !ok || !applied {
		t.Fatalf("first completion: applied=%v ok=%v body=%s", applied, ok, w.Body.String())
	}
	exp := expiry(t, 8000001)
	want := time.Now().UTC().AddDate(4, 0, 0)
	if d := exp.Sub(want); d < -time.Minute || d > time.Minute {
		t.Errorf("expiry %v, want ~%v (four years from now, not from the lapsed date)", exp, want)
	}
	if !completed(t, sid, "MEMBERSHIP") {
		t.Error("MEMBERSHIP row not marked completed")
	}
	if completed(t, sid, "SHIRT") {
		t.Error("SHIRT row must stay open until pickup")
	}

	applied, ok = completeStripeSession(w, sid)
	if !ok || applied {
		t.Fatalf("second completion: applied=%v ok=%v, want a no-op", applied, ok)
	}
	if again := expiry(t, 8000001); !again.Equal(exp) {
		t.Errorf("retry stacked years: %v -> %v", exp, again)
	}
}

/* A custom-amount donation is not a membership. */
func TestCompleteStripeSessionCustomAmount(t *testing.T) {
	openTestDB(t)
	w := httptest.NewRecorder()
	const sid = "cs_test_custom"
	before := expiry(t, 8000001)
	dbInsertPayment(w, 8000000, "", 8000001, "CUSTOM", 1, 2500, "STRIPE", sid, false)

	applied, ok := completeStripeSession(w, sid)
	if !ok || !applied {
		t.Fatalf("applied=%v ok=%v body=%s", applied, ok, w.Body.String())
	}
	if !completed(t, sid, "CUSTOM") {
		t.Error("CUSTOM row not marked completed")
	}
	if after := expiry(t, 8000001); !after.Equal(before) {
		t.Errorf("custom amount extended membership: %v -> %v", before, after)
	}
}

/* Renewing early adds to the existing expiry rather than restarting from today. */
func TestExtendMembershipFromCurrentExpiry(t *testing.T) {
	openTestDB(t)
	w := httptest.NewRecorder()
	if !dbExtendMembership(w, 8000002, 1) {
		t.Fatal(w.Body.String())
	}
	exp := expiry(t, 8000002)
	want := time.Date(2028, 3, 1, 0, 0, 0, 0, time.UTC)
	if !exp.Equal(want) {
		t.Errorf("expiry %v, want %v", exp, want)
	}
}

/* Refunding a duplicate checkout in Stripe must be enough to keep it from granting. */
func TestStripeSessionRefunded(t *testing.T) {
	paid := &stripe.CheckoutSession{PaymentIntent: &stripe.PaymentIntent{
		Status:  stripe.PaymentIntentStatusSucceeded,
		Charges: &stripe.ChargeList{Data: []*stripe.Charge{{Refunded: false}}},
	}}
	refunded := &stripe.CheckoutSession{PaymentIntent: &stripe.PaymentIntent{
		Status:  stripe.PaymentIntentStatusSucceeded,
		Charges: &stripe.ChargeList{Data: []*stripe.Charge{{Refunded: true}}},
	}}
	noCharges := &stripe.CheckoutSession{PaymentIntent: &stripe.PaymentIntent{Status: stripe.PaymentIntentStatusSucceeded}}

	if stripeSessionRefunded(paid) {
		t.Error("an unrefunded charge reported as refunded")
	}
	if !stripeSessionRefunded(refunded) {
		t.Error("a refunded charge not detected")
	}
	if stripeSessionRefunded(noCharges) || stripeSessionRefunded(&stripe.CheckoutSession{}) {
		t.Error("missing charge data must not read as refunded")
	}
}

/*
 * Rows ticked "completed" in webtools during the outage, where the member
 * never got the years, must be re-opened so the sweep can settle them.
 */
func TestResetUnappliedStripePayments(t *testing.T) {
	openTestDB(t)
	w := httptest.NewRecorder()
	// Lapsed member: ticked complete by hand, never granted
	dbInsertPayment(w, 8000000, "", 8000001, "MEMBERSHIP", 4, 6500, "STRIPE", "cs_test_ticked", true)
	dbInsertPayment(w, 8000000, "", 8000001, "SHIRT", 1, 6500, "STRIPE", "cs_test_ticked", true)
	// Active member: completed properly, years applied
	dbInsertPayment(w, 8000000, "", 8000002, "MEMBERSHIP", 1, 2000, "STRIPE", "cs_test_granted", false)
	if _, ok := completeStripeSession(w, "cs_test_granted"); !ok {
		t.Fatal(w.Body.String())
	}

	utils.DBMigrate(db)

	if completed(t, "cs_test_ticked", "MEMBERSHIP") {
		t.Error("hand-ticked row with no membership effect was not re-opened")
	}
	if !completed(t, "cs_test_ticked", "SHIRT") {
		t.Error("shirt pickup state must not be touched")
	}
	if !completed(t, "cs_test_granted", "MEMBERSHIP") {
		t.Error("a payment whose years were applied was wrongly re-opened")
	}
}
