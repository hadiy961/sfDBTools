package validate

import (
	"os"
	"testing"

	"sfDBTools/internal/config/model"
)

func validLogConfig(dir string) model.LogConfig {
	return model.LogConfig{
		Level:    "info",
		Format:   "text",
		Timezone: "Asia/Jakarta",
		Output: model.LogOutput{
			Console: true,
			File:    true,
		},
		File: model.LogFileSetting{
			Dir:           dir,
			RotateDaily:   false,
			RetentionDays: 1,
		},
	}
}

func TestLog(t *testing.T) {
	dir := t.TempDir()
	cfg := validLogConfig(dir)
	if err := Log(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	invalid := cfg
	invalid.Level = "bad"
	if err := Log(invalid); err == nil {
		t.Errorf("expected error for invalid level")
	}

	invalid = cfg
	invalid.Format = "xml"
	if err := Log(invalid); err == nil {
		t.Errorf("expected error for invalid format")
	}

	invalid = cfg
	invalid.Timezone = "Bad/Zone"
	if err := Log(invalid); err == nil {
		t.Errorf("expected error for invalid timezone")
	}

	invalid = cfg
	invalid.Output.File = true
	invalid.File.Dir = ""
	if err := Log(invalid); err == nil {
		t.Errorf("expected error when file output enabled without dir")
	}

	invalid = cfg
	invalid.Output.File = true
	invalid.File.Dir = dir + "/nonexist"
	if err := Log(invalid); err != nil {
		t.Fatalf("unexpected error for auto-create dir: %v", err)
	}
	if _, err := os.Stat(invalid.File.Dir); err != nil {
		t.Errorf("dir should be created: %v", err)
	}

	invalid = cfg
	invalid.File.RetentionDays = 0
	if err := Log(invalid); err == nil {
		t.Errorf("expected error for bad retention days")
	}
}
