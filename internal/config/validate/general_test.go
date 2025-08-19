package validate

import (
	"sfDBTools/internal/config/model"
	"testing"
)

func TestGeneral(t *testing.T) {
	valid := model.GeneralConfig{
		ClientCode: "123",
		AppName:    "sfDBTools",
		Version:    "1.0.0",
		Author:     "Hadiyatna Muflihun",
	}
	if err := General(valid); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	invalid := valid
	invalid.ClientCode = ""
	if err := General(invalid); err == nil {
		t.Errorf("expected error for empty client_code")
	}
	invalid = valid
	invalid.AppName = "other"
	if err := General(invalid); err == nil {
		t.Errorf("expected error for wrong app_name")
	}
	invalid = valid
	invalid.Version = "0.1"
	if err := General(invalid); err == nil {
		t.Errorf("expected error for wrong version")
	}
	invalid = valid
	invalid.Author = "Someone"
	if err := General(invalid); err == nil {
		t.Errorf("expected error for wrong author")
	}
}
