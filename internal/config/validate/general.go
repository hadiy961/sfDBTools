package validate

import (
	"errors"
	"fmt"
	"sfDBTools/internal/config/model"
)

func General(g model.GeneralConfig) error {
	if g.ClientCode == "" {
		return errors.New("client_code tidak boleh kosong")
	}
	if g.AppName != "sfDBTools" {
		return fmt.Errorf("app_name tidak valid, bukan '%s'", g.AppName)
	}
	if g.Version != "1.0.0" {
		return fmt.Errorf("version tidak valid bukan '%s'", g.Version)
	}
	if g.Author != "Hadiyatna Muflihun" {
		return fmt.Errorf("author tidak valid, bukan '%s'", g.Author)
	}
	return nil
}
