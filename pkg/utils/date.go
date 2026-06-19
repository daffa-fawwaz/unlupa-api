package utils

import "time"

// NormalizeDate strips the time component and returns midnight in UTC.
// Using UTC ensures the date value stored in a PostgreSQL `date` column
// is always consistent regardless of the server's local timezone, because
// the Go Postgres driver sends time.Time values in UTC to the DB.
// Passing a timezone-aware time (e.g. +07:00) would cause GORM/pgx to send
// "2026-06-16T00:00:00+07:00", which PostgreSQL stores as the date 2026-06-15
// (UTC), creating a mismatch on every subsequent WHERE task_date = ? query.
func NormalizeDate(t time.Time) time.Time {
	// Convert to local app time first so we get the correct calendar day,
	// then zero out the time component in UTC for consistent DB storage.
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func startOfDay(t time.Time) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		0, 0, 0, 0,
		t.Location(),
	)
}