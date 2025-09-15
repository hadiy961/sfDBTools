package template

// MariaDBConfigTemplate berisi template konfigurasi MariaDB
type MariaDBConfigTemplate struct {
	TemplatePath  string            `json:"template_path"`
	Content       string            `json:"content"`
	Placeholders  map[string]string `json:"placeholders"`
	DefaultValues map[string]string `json:"default_values"`
	CurrentConfig string            `json:"current_config"`
	CurrentPath   string            `json:"current_path"`
}
