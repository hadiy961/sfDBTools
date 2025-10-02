package parsing

import (
	"fmt"
	"reflect"
	"sfDBTools/utils/common"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// DynamicParseFlags mengiterasi struct target, membaca tags 'flag' dan 'env',
// dan mengisi nilai field menggunakan helper common.Get*FlagOrEnv.
func DynamicParseFlags(cmd *cobra.Command, target interface{}) error {

	val := reflect.ValueOf(target).Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Tangani Embedded/Nested Struct (Rekursif)
		if field.Anonymous || (fieldVal.Kind() == reflect.Struct && field.Type.Name() != "Time" /* hindari struct standar Go */) {
			// Jika field adalah embedded struct atau nested struct
			if fieldVal.Kind() == reflect.Struct && fieldVal.CanAddr() {
				if err := DynamicParseFlags(cmd, fieldVal.Addr().Interface()); err != nil {
					return err
				}
			}
			continue
		}

		flagName := field.Tag.Get("flag")
		envName := field.Tag.Get("env")

		if flagName == "" {
			continue // Lewati field yang tidak memiliki tag 'flag'
		}

		flag := cmd.Flag(flagName)
		if flag == nil {
			// Seharusnya tidak terjadi jika flags didaftarkan dengan benar di init()
			return fmt.Errorf("flag not registered: %s", flagName)
		}

		switch field.Type.Kind() {
		case reflect.String:
			// Ambil nilai default COBRA (sudah termasuk nilai dari config/init)
			defaultVal := flag.Value.String()
			parsedVal := common.GetStringFlagOrEnv(cmd, flagName, envName, defaultVal)
			fieldVal.SetString(parsedVal)

		case reflect.Int:
			// Ambil nilai default COBRA
			defaultVal, _ := cmd.Flags().GetInt(flagName)
			parsedVal := common.GetIntFlagOrEnv(cmd, flagName, envName, defaultVal)
			fieldVal.SetInt(int64(parsedVal))

		case reflect.Bool:
			// Ambil nilai default COBRA
			defaultVal, _ := cmd.Flags().GetBool(flagName)
			parsedVal := common.GetBoolFlagOrEnv(cmd, flagName, envName, defaultVal)
			fieldVal.SetBool(parsedVal)

		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				// Ambil nilai default COBRA
				defaultVal, _ := cmd.Flags().GetStringSlice(flagName)
				parsedVal := common.GetStringSliceFlagOrEnv(cmd, flagName, envName, defaultVal)

				// Assign slice
				sliceVal := reflect.MakeSlice(field.Type, len(parsedVal), len(parsedVal))
				for k, v := range parsedVal {
					sliceVal.Index(k).SetString(v)
				}
				fieldVal.Set(sliceVal)
			} else {
				return fmt.Errorf("unsupported slice type for flag %s: %s", flagName, field.Type)
			}
		default:
			return fmt.Errorf("unsupported field type for flag %s: %s", flagName, field.Type)
		}
	}
	return nil
}

// parseTagDefault mengkonversi string dari struct tag 'default' ke tipe data yang sesuai.
func ParseTagDefault(tag string, kind reflect.Kind) (interface{}, error) {
	if tag == "" {
		// Mengembalikan nilai nol Go jika tag kosong
		return reflect.Zero(reflect.TypeOf(kind)).Interface(), nil
	}

	switch kind {
	case reflect.String:
		return tag, nil
	case reflect.Bool:
		if strings.ToLower(tag) == "true" {
			return true, nil
		}
		if strings.ToLower(tag) == "false" {
			return false, nil
		}
		return nil, fmt.Errorf("invalid bool default: %s", tag)
	case reflect.Int:
		val, err := strconv.Atoi(tag)
		if err != nil {
			return nil, fmt.Errorf("invalid int default: %w", err)
		}
		return val, nil
	case reflect.Slice:
		if tag == "" {
			return []string{}, nil
		}
		// Diasumsikan format comma-separated string untuk []string
		slice := strings.Split(tag, ",")
		var cleanedSlice []string
		for _, s := range slice {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				cleanedSlice = append(cleanedSlice, trimmed)
			}
		}
		return cleanedSlice, nil
	default:
		return nil, fmt.Errorf("unsupported field type: %s", kind)
	}
}
