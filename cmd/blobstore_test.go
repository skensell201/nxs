package cmd

import (
	"testing"

	"github.com/nexspence/nxs/internal/client"
)

func int64p(v int64) *int64 { return &v }

func TestBlobStoreRows(t *testing.T) {
	rows := blobStoreRows([]client.BlobStore{
		{Name: "default", Type: "local", UsedBytes: 2048},
		{Name: "archive", Type: "s3", UsedBytes: 1024, QuotaBytes: int64p(10 * 1024 * 1024 * 1024)},
	})

	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	want := [][]string{
		{"default", "local", "2.0 KB", "unlimited"},
		{"archive", "s3", "1.0 KB", "10.0 GB"},
	}
	for i, w := range want {
		for j := range w {
			if rows[i][j] != w[j] {
				t.Errorf("row %d col %d: want %q, got %q", i, j, w[j], rows[i][j])
			}
		}
	}
}

func TestBlobStoreRows_Empty(t *testing.T) {
	if rows := blobStoreRows(nil); len(rows) != 0 {
		t.Errorf("want no rows, got %v", rows)
	}
}

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{5 * 1024 * 1024 * 1024, "5.0 GB"},
		{3 * 1024 * 1024 * 1024 * 1024, "3.0 TB"},
	}
	for _, c := range cases {
		if got := humanBytes(c.in); got != c.want {
			t.Errorf("humanBytes(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}
