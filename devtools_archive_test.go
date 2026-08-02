package snapshot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveDirectoryProducesBytes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tool-state.json"), []byte(`{"version":1}`), 0644); err != nil {
		t.Fatal(err)
	}
	archive, err := ArchiveDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(archive) == 0 {
		t.Fatal("expected archive bytes")
	}
}
