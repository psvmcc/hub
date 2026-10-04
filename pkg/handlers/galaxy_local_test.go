package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/psvmcc/hub/pkg/types"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func galaxyLocalRequest(t *testing.T, dir string, handler func(string) echo.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	var cfg types.ConfigFile
	if err := yaml.Unmarshal(fmt.Appendf(nil, "server:\n  galaxy:\n    test:\n      dir: %q\n", dir), &cfg); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", http.NoBody), rec)
	c.SetParamNames("namespace", "name")
	c.SetParamValues("ns", "coll")
	c.Set("cfg", cfg)
	c.Set("logger", zap.NewNop().Sugar())

	if err := handler("test")(c); err != nil {
		t.Fatal(err)
	}
	return rec
}

func galaxyLocalDir(t *testing.T, versions ...string) string {
	t.Helper()
	dir := t.TempDir()
	collectionDir := filepath.Join(dir, "ns", "coll")
	if err := os.MkdirAll(collectionDir, 0o750); err != nil {
		t.Fatal(err)
	}
	for _, v := range versions {
		path := filepath.Join(collectionDir, fmt.Sprintf("ns-coll-%s.tar.gz", v))
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestGalaxyLocalCollectionWithoutVersions(t *testing.T) {
	dir := galaxyLocalDir(t)
	handlers := map[string]func(string) echo.HandlerFunc{
		"collection": GalaxyLocalCollection,
		"versions":   GalaxyLocalCollectionVersions,
	}
	for name, handler := range handlers {
		rec := galaxyLocalRequest(t, dir, handler)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: got status %d, want %d", name, rec.Code, http.StatusNotFound)
		}
	}
}

func TestGalaxyLocalCollectionHighestVersion(t *testing.T) {
	dir := galaxyLocalDir(t, "1.2.0", "1.2.10", "1.2.11", "1.2.2", "1.2.9")
	rec := galaxyLocalRequest(t, dir, GalaxyLocalCollection)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
	}
	var collection types.GalaxyCollection
	if err := json.Unmarshal(rec.Body.Bytes(), &collection); err != nil {
		t.Fatal(err)
	}
	if collection.HighestVersion.Version != "1.2.11" {
		t.Errorf("got highest_version %s, want 1.2.11", collection.HighestVersion.Version)
	}
}
