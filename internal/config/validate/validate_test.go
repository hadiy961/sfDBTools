package validate

import (
	"sfDBTools/internal/config/model"
	"testing"
)

func TestAll(t *testing.T) {
	dir := t.TempDir()
	cfg := &model.Config{
		General: model.GeneralConfig{
			ClientCode: "123",
			AppName:    "sfDBTools",
			Version:    "1.0.0",
			Author:     "Hadiyatna Muflihun",
		},
		Log: model.LogConfig{
			Level:    "info",
			Format:   "text",
			Timezone: "Asia/Jakarta",
			Output:   model.LogOutput{Console: true, File: true},
			File: model.LogFileSetting{
				Dir:           dir,
				RetentionDays: 1,
			},
		},
	}
	if err := All(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
