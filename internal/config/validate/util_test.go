package validate

import (
	"os"
	"testing"
)

func TestInSlice(t *testing.T) {
	list := []string{"a", "b", "c"}
	if !InSlice("b", list) {
		t.Errorf("expected true when value exists")
	}
	if InSlice("d", list) {
		t.Errorf("expected false when value not in slice")
	}
}

func TestIsValidTimezone(t *testing.T) {
	if err := IsValidTimezone("Asia/Jakarta"); err != nil {
		t.Errorf("valid timezone returned error: %v", err)
	}
	if err := IsValidTimezone("Invalid/Zone"); err == nil {
		t.Errorf("expected error for invalid timezone")
	}
}

func TestDirExistsAndWritable(t *testing.T) {
	dir := t.TempDir()
	if err := DirExistsAndWritable(dir); err != nil {
		t.Fatalf("temp dir should be writable: %v", err)
	}

	nonexist := dir + "/nonexist"
	if err := DirExistsAndWritable(nonexist); err != nil {
		t.Fatalf("should create missing dir: %v", err)
	}
	if _, err := os.Stat(nonexist); err != nil {
		t.Errorf("directory should be created: %v", err)
	}

	// create a file and test
	file := dir + "/file"
	if err := os.WriteFile(file, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := DirExistsAndWritable(file); err == nil {
		t.Errorf("expected error when path is not directory")
	}
}
