package validate

import (
	"fmt"
	"sfDBTools/internal/config/model"
)

func All(cfg *model.Config) error {
	if err := General(cfg.General); err != nil {
		return fmt.Errorf("general: %w", err)
	}
	if err := Log(cfg.Log); err != nil {
		return fmt.Errorf("log: %w", err)
	}
	return nil
}
