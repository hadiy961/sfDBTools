// file: internal/config/default_config/default_dbconfig.go
// Description: Default configuration untuk DBConfig
// Author: Hadiyatna Muflihun

package defaultconfig

import "sfDBTools/utils/common/structs"

func GetDBConfigGenerateDefaults() (*structs.DBConfigGenerateOptions, error) {

	// Definisikan semua Default Settings dalam satu Struct Literal yang bersih.
	DBConfigGenerateOptions := &structs.DBConfigGenerateOptions{
		ConnectionOptions: structs.ConnectionOptions{
			Host: "localhost",
			Port: 3306,
			User: "root",
		},
		EncryptionOptions: structs.EncryptionOptions{
			Required: true,
		},
		Name: "local_db",
	}

	// 2. Kembalikan struct dengan default values.
	return DBConfigGenerateOptions, nil
}
