package client_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nexspence/nxs/internal/client"
)

func TestClient_BlobStoreCompact(t *testing.T) {
	var gotPath, gotQuery, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotQuery = r.Method, r.URL.Path, r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"store":"default","scannedBlobs":5,"orphans":2,"freedBytes":2048,"dryRun":true}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "token")
	res, err := c.BlobStoreCompact("default", true, "24h")
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost || gotPath != "/api/v1/blobstores/default/compact" {
		t.Errorf("unexpected %s %s", gotMethod, gotPath)
	}
	if !strings.Contains(gotQuery, "dry_run=true") || !strings.Contains(gotQuery, "min_age=24h") {
		t.Errorf("unexpected query %q", gotQuery)
	}
	if res.Orphans != 2 || res.FreedBytes != 2048 || !res.DryRun || res.Store != "default" {
		t.Errorf("unexpected result %+v", res)
	}
}

func TestClient_BlobStoreList(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{"id":"bs-1","name":"default","type":"local","usedBytes":2048},
			{"id":"bs-2","name":"archive","type":"s3","quotaBytes":10737418240,"usedBytes":1024}
		]`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "token")
	stores, err := c.BlobStoreList()
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodGet || gotPath != "/service/rest/v1/blobstores" {
		t.Errorf("unexpected %s %s", gotMethod, gotPath)
	}
	if len(stores) != 2 {
		t.Fatalf("want 2 stores, got %d", len(stores))
	}
	if stores[0].Name != "default" || stores[0].Type != "local" || stores[0].UsedBytes != 2048 {
		t.Errorf("unexpected first store %+v", stores[0])
	}
	if stores[0].QuotaBytes != nil {
		t.Errorf("store without a quota must decode to nil, got %v", *stores[0].QuotaBytes)
	}
	if stores[1].QuotaBytes == nil || *stores[1].QuotaBytes != 10737418240 {
		t.Errorf("unexpected quota on second store %+v", stores[1])
	}
}

func TestClient_BlobStoreList_PropagatesHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"forbidden"}`))
	}))
	defer srv.Close()

	if _, err := client.New(srv.URL, "token").BlobStoreList(); err == nil {
		t.Fatal("want an error for a 403 response")
	}
}

func TestClient_BlobStoreInfo(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"bs-2","name":"archive","type":"s3","quotaBytes":10737418240,"usedBytes":1024}`))
	}))
	defer srv.Close()

	bs, err := client.New(srv.URL, "token").BlobStoreInfo("archive")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/service/rest/v1/blobstores/archive" {
		t.Errorf("unexpected path %q", gotPath)
	}
	if bs.Name != "archive" || bs.Type != "s3" || bs.UsedBytes != 1024 {
		t.Errorf("unexpected store %+v", bs)
	}
}

func TestClient_BlobStoreInfo_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"blob store not found"}`))
	}))
	defer srv.Close()

	if _, err := client.New(srv.URL, "token").BlobStoreInfo("ghost"); err == nil {
		t.Fatal("want an error for a 404 response")
	}
}
