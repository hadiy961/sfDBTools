package interactive

import (
	"sfDBTools/utils/interaction"
)

// InputCollector is a thin wrapper around the reusable interaction.InputCollector
// so we can keep existing function signatures in this package while using the
// shared implementation in `utils/interaction`.
type InputCollector struct {
	impl     *interaction.InputCollector
	Defaults *ConfigDefaults
}

// NewInputCollector creates a new interactive InputCollector backed by
// the shared `utils/interaction` implementation. `defaults` must implement
// the DefaultsResolver methods required by the shared collector.
func NewInputCollector(defaults *ConfigDefaults) *InputCollector {
	impl := interaction.NewInputCollector(defaults)
	return &InputCollector{impl: impl, Defaults: defaults}
}

// CollectString forwards to the shared implementation.
func (ic *InputCollector) CollectString(question string, currentValue string, templateKey string, hardcodedDefault string, validator func(string) error) (string, error) {
	return ic.impl.CollectString(question, currentValue, templateKey, hardcodedDefault, validator)
}

// CollectInt forwards to the shared implementation.
func (ic *InputCollector) CollectInt(question string, currentValue int, templateKey string, validator func(int) error) (int, error) {
	return ic.impl.CollectInt(question, currentValue, templateKey, validator)
}

// CollectBool forwards to the shared implementation.
func (ic *InputCollector) CollectBool(question string, defaultValue bool) bool {
	return ic.impl.CollectBool(question, defaultValue)
}

// CollectDirectory forwards to the shared implementation. Note: the shared
// implementation accepts a validator; we pass the package's ValidateAbsolutePath
// where callers previously relied on it.
func (ic *InputCollector) CollectDirectory(question string, currentValue string, templateKey string, hardcodedDefault string) (string, error) {
	return ic.impl.CollectDirectory(question, currentValue, templateKey, hardcodedDefault, ValidateAbsolutePath)
}
