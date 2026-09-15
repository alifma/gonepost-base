package config

import "testing"

func TestLoad(t *testing.T) {
	cases := []struct {
		name     string
		envValue string
		want     string
	}{
		{"PORT not set", "", ":8080"},
		{"PORT set to 9090", "9090", ":9090"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("PORT", tc.envValue)
			t.Setenv("DATABASE_URL", "postgres://basecode:basecode@localhost:5432/basecode?sslmode=disable")
			got, err := Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Port != tc.want {
				t.Errorf("got %q, want %q", got.Port, tc.want)
			}
		})
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Error("expected error when DATABASE_URL is missing, got nil")
	}
}
