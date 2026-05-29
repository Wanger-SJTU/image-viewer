package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"image-viewer/internal/config"
	"image-viewer/internal/repository"
	"image-viewer/internal/service"
	"image-viewer/shared/types"
)

func setupHTTPTest(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "http_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	dbPath := tmpDir + "/test.db"
	db, err := repository.InitDB(dbPath)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}

	cfg := &config.Config{
		Port:              8080,
		DBPath:            dbPath,
		CacheDir:          tmpDir + "/cache",
		SupportedRawExts:  []string{".arw", ".cr3"},
		SupportedJpgExts:  []string{".jpg", ".jpeg"},
		ConcurrencyLimit:  2,
	}
	os.MkdirAll(cfg.CacheDir, 0755)

	repo := repository.NewAssetRepository(db)
	assetSvc := service.NewAssetService(cfg, repo)
	scannerSvc := service.NewScannerService(cfg, repo)
	thumbSvc := service.NewThumbService(cfg, repo)

	// Seed test data
	seedAssets := []*types.Asset{{
		Name: "DSC0001", DirPath: "/photos/test",
		MatchStatus: types.MatchStatusPaired, Rating: 3,
	}}
	seedFiles := []*types.MediaFile{{
		AssetID: 1, FilePath: "/photos/test/DSC0001.JPG", FileName: "DSC0001.JPG",
		MediaType: types.MediaTypeJPG, FileSize: 1024,
	}, {
		AssetID: 1, FilePath: "/photos/test/DSC0001.ARW", FileName: "DSC0001.ARW",
		MediaType: types.MediaTypeRAW, FileSize: 50000,
	}}
	repo.BulkUpsert(context.Background(), seedAssets, seedFiles)

	router := NewRouter(assetSvc, scannerSvc, thumbSvc, nil)
	server := httptest.NewServer(router)

	cleanup := func() {
		server.Close()
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return server, cleanup
}

func TestHealthEndpoint(t *testing.T) {
	server, cleanup := setupHTTPTest(t)
	defer cleanup()

	resp, err := http.Get(server.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["success"] != true {
		t.Errorf("success = %v, want true", body["success"])
	}
}

func TestListAssets(t *testing.T) {
	server, cleanup := setupHTTPTest(t)
	defer cleanup()

	resp, err := http.Get(server.URL + "/api/v1/assets")
	if err != nil {
		t.Fatalf("GET /assets: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["success"] != true {
		t.Errorf("success = %v, want true", body["success"])
	}
}

func TestGetAsset(t *testing.T) {
	server, cleanup := setupHTTPTest(t)
	defer cleanup()

	resp, err := http.Get(server.URL + "/api/v1/assets/1")
	if err != nil {
		t.Fatalf("GET /assets/1: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestGetAsset_NotFound(t *testing.T) {
	server, cleanup := setupHTTPTest(t)
	defer cleanup()

	resp, err := http.Get(server.URL + "/api/v1/assets/99999")
	if err != nil {
		t.Fatalf("GET /assets/99999: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRateAsset(t *testing.T) {
	server, cleanup := setupHTTPTest(t)
	defer cleanup()

	body := strings.NewReader(`{"rating": 5}`)
	resp, err := http.Post(server.URL+"/api/v1/assets/1/rate", "application/json", body)
	if err != nil {
		t.Fatalf("POST /assets/1/rate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRateAsset_InvalidRating(t *testing.T) {
	server, cleanup := setupHTTPTest(t)
	defer cleanup()

	body := strings.NewReader(`{"rating": 10}`)
	resp, err := http.Post(server.URL+"/api/v1/assets/1/rate", "application/json", body)
	if err != nil {
		t.Fatalf("POST /assets/1/rate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestLabelAsset(t *testing.T) {
	server, cleanup := setupHTTPTest(t)
	defer cleanup()

	body := strings.NewReader(`{"color_label": "red"}`)
	resp, err := http.Post(server.URL+"/api/v1/assets/1/label", "application/json", body)
	if err != nil {
		t.Fatalf("POST /assets/1/label: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestDeleteAsset(t *testing.T) {
	server, cleanup := setupHTTPTest(t)
	defer cleanup()

	req, _ := http.NewRequest("DELETE", server.URL+"/api/v1/assets/1", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /assets/1: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	// Verify deleted
	resp2, _ := http.Get(server.URL + "/api/v1/assets/1")
	if resp2.StatusCode != 404 {
		t.Errorf("GET after delete status = %d, want 404", resp2.StatusCode)
	}
}

func TestStartScan_InvalidBody(t *testing.T) {
	server, cleanup := setupHTTPTest(t)
	defer cleanup()

	body := strings.NewReader(`{}`)
	resp, err := http.Post(server.URL+"/api/v1/scan", "application/json", body)
	if err != nil {
		t.Fatalf("POST /scan: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}
