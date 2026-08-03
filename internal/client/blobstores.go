package client

import (
	"encoding/json"
	"fmt"
)

// BlobStore is a physical storage backend that repositories write their blobs to.
// The server's config map is deliberately not decoded here: it carries backend
// credentials that have no business being printed by a CLI.
type BlobStore struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	QuotaBytes *int64 `json:"quotaBytes,omitempty"`
	UsedBytes  int64  `json:"usedBytes"`
}

// BlobStoreList returns every blob store the server knows about.
func (c *Client) BlobStoreList() ([]BlobStore, error) {
	resp, err := c.r.R().Get("/service/rest/v1/blobstores")
	if err != nil {
		return nil, err
	}
	if err := checkErr(resp); err != nil {
		return nil, err
	}
	var stores []BlobStore
	return stores, json.Unmarshal(resp.Body(), &stores)
}

// BlobStoreInfo returns a single blob store by name.
func (c *Client) BlobStoreInfo(name string) (*BlobStore, error) {
	resp, err := c.r.R().Get("/service/rest/v1/blobstores/" + name)
	if err != nil {
		return nil, err
	}
	if err := checkErr(resp); err != nil {
		return nil, err
	}
	var bs BlobStore
	return &bs, json.Unmarshal(resp.Body(), &bs)
}

// GCResult reports what a blob store compaction found and removed.
type GCResult struct {
	Store        string   `json:"store"`
	ScannedBlobs int      `json:"scannedBlobs"`
	Orphans      int      `json:"orphans"`
	FreedBytes   int64    `json:"freedBytes"`
	DryRun       bool     `json:"dryRun"`
	Errors       []string `json:"errors,omitempty"`
}

// BlobStoreCompact runs garbage collection on a blob store, removing blobs no
// longer referenced by any asset. When dryRun is true, orphans are reported but
// not deleted. minAge (e.g. "24h") overrides the server's grace period; an empty
// string uses the server default.
func (c *Client) BlobStoreCompact(name string, dryRun bool, minAge string) (*GCResult, error) {
	req := c.r.R().SetHeader("Content-Type", "application/json")
	if dryRun {
		req.SetQueryParam("dry_run", "true")
	}
	if minAge != "" {
		req.SetQueryParam("min_age", minAge)
	}
	resp, err := req.Post(fmt.Sprintf("/api/v1/blobstores/%s/compact", name))
	if err != nil {
		return nil, err
	}
	if err := checkErr(resp); err != nil {
		return nil, err
	}
	var res GCResult
	return &res, json.Unmarshal(resp.Body(), &res)
}
