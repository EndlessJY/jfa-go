package main

import (
	"testing"
	"time"
)

func TestCalculateRenewalExpiryExtendsFromFutureExistingExpiry(t *testing.T) {
	now := time.Date(2026, time.June, 27, 9, 0, 0, 0, time.UTC)
	existing := &UserExpiry{Expiry: now.AddDate(0, 0, 10)}
	invite := Invite{
		UserExpiry:  true,
		UserMonths:  1,
		UserDays:    2,
		UserHours:   3,
		UserMinutes: 4,
	}

	got, ok := calculateRenewalExpiry(now, existing, invite)
	if !ok {
		t.Fatal("expected invite with user expiry duration to be renewable")
	}

	want := existing.Expiry.AddDate(0, 1, 2).Add((3*time.Hour + 4*time.Minute))
	if !got.Equal(want) {
		t.Fatalf("expiry = %s, want %s", got, want)
	}
}

func TestCalculateRenewalExpiryUsesNowWhenExistingExpiryIsExpiredOrMissing(t *testing.T) {
	now := time.Date(2026, time.June, 27, 9, 0, 0, 0, time.UTC)
	invite := Invite{
		UserExpiry: true,
		UserDays:   30,
	}

	tests := []struct {
		name     string
		existing *UserExpiry
	}{
		{
			name:     "expired existing expiry",
			existing: &UserExpiry{Expiry: now.AddDate(0, 0, -1)},
		},
		{
			name:     "missing existing expiry",
			existing: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := calculateRenewalExpiry(now, tt.existing, invite)
			if !ok {
				t.Fatal("expected invite with user expiry duration to be renewable")
			}

			want := now.AddDate(0, 0, 30)
			if !got.Equal(want) {
				t.Fatalf("expiry = %s, want %s", got, want)
			}
		})
	}
}

func TestCalculateRenewalExpiryRejectsInviteWithoutUserExpiryDuration(t *testing.T) {
	now := time.Date(2026, time.June, 27, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		invite Invite
	}{
		{name: "user expiry disabled", invite: Invite{UserExpiry: false, UserDays: 30}},
		{name: "user expiry enabled but zero duration", invite: Invite{UserExpiry: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := calculateRenewalExpiry(now, nil, tt.invite); ok {
				t.Fatal("expected invite without user expiry duration to be rejected")
			}
		})
	}
}

func TestNormalizeInviteCodeAcceptsRawCodesAndInviteLinks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "raw code", input: "  abc123  ", want: "abc123"},
		{name: "simple invite link", input: "https://example.com/invite/abc123", want: "abc123"},
		{name: "invite link under subfolder", input: "https://example.com/jfa-go/invite/abc123?utm=renew", want: "abc123"},
		{name: "invite link with trailing slash", input: "https://example.com/invite/abc123/", want: "abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeInviteCode(tt.input); got != tt.want {
				t.Fatalf("code = %q, want %q", got, tt.want)
			}
		})
	}
}
