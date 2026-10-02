package types

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGalaxyLocalListLatest(t *testing.T) {
	dir := t.TempDir()
	versions := []string{"1.2.0", "1.2.10", "1.2.11", "1.2.2", "1.2.9"}
	base := time.Now().Add(-time.Hour)
	for i, v := range versions {
		path := filepath.Join(dir, fmt.Sprintf("ns-coll-%s.tar.gz", v))
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		// 1.2.2 gets the newest mtime, 1.2.11 an older one.
		mtime := base.Add(time.Duration(i) * time.Minute)
		if v == "1.2.2" {
			mtime = base.Add(30 * time.Minute)
		}
		if err := os.Chtimes(path, mtime, mtime); err != nil {
			t.Fatal(err)
		}
	}

	var g GalaxyLocal
	if err := g.List(dir, "ns", "coll"); err != nil {
		t.Fatal(err)
	}
	if len(g.Versions) != len(versions) {
		t.Fatalf("got %d versions, want %d", len(g.Versions), len(versions))
	}
	if g.Latest.Version != "1.2.11" {
		t.Errorf("got latest %s, want 1.2.11", g.Latest.Version)
	}
}

func TestGalaxyLocalListEmpty(t *testing.T) {
	var g GalaxyLocal
	if err := g.List(t.TempDir(), "ns", "coll"); err == nil {
		t.Error("expected error for directory without versions")
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.2.11", "1.2.2", 1},
		{"1.2.2", "1.2.11", -1},
		{"1.2.10", "1.2.9", 1},
		{"1.0.0", "0.99.99", 1},
		{"2.10.0", "2.9.5", 1},
		{"1.2.3", "1.2.3", 0},
		{"1.02.3", "1.2.3", 0},
		{"0.0.0", "0.0.1", -1},
	}
	for _, tt := range tests {
		if got := compareVersions(tt.a, tt.b); got != tt.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
