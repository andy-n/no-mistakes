package verificationplan

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotSurvivesSourceRemovalAndRejectsTampering(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.txt")
	original := []byte("  Verify café\r\nExpected: exact bytes.\n\n")
	if err := os.WriteFile(source, original, 0600); err != nil {
		t.Fatal(err)
	}
	plan, err := Capture(filepath.Join(root, "inputs"), source, "repo", "branch", "head")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	resolved, err := Resolve(filepath.Join(root, "inputs"), plan.ID, "repo", "branch", "head")
	if err != nil {
		t.Fatal(err)
	}
	got, err := resolved.Read()
	if err != nil || !bytes.Equal(got, original) {
		t.Fatalf("snapshot changed: %q %v", got, err)
	}
	if _, err := Resolve(filepath.Join(root, "inputs"), plan.ID, "repo", "other", "head"); err == nil {
		t.Fatal("capture rebound across branches")
	}
	if err := os.Chmod(plan.Path, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan.Path, []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := plan.Read(); err == nil {
		t.Fatal("tampered evidence accepted")
	}
}
