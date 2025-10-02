package flags

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
)

// DynamicAddFlags menggunakan reflection untuk mendaftarkan flags Cobra dari struct.
// sourceStruct harus berupa pointer ke struct yang telah diisi dengan nilai default.
// Nilai default diambil langsung dari field struct.
func DynamicAddFlags(cmd *cobra.Command, sourceStruct interface{}) error {

	val := reflect.ValueOf(sourceStruct).Elem()
	typ := val.Type()

	// Panggil fungsi rekursif untuk menangani embedded struct
	return addFlagsRecursive(cmd, val, typ)
}

func addFlagsRecursive(cmd *cobra.Command, val reflect.Value, typ reflect.Type) error {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Tangani Embedded/Nested Struct (Rekursif)
		if field.Anonymous || (fieldVal.Kind() == reflect.Struct && field.Type.Name() != "Time") {
			if fieldVal.CanAddr() {
				if err := addFlagsRecursive(cmd, fieldVal, fieldVal.Type()); err != nil {
					return err
				}
			}
			continue
		}

		flagName := field.Tag.Get("flag")
		if flagName == "" {
			continue
		}

		// Asumsi: Penggunaan "usage" telah diimplementasikan dengan benar
		usage := fmt.Sprintf("Option for %s", strings.ToLower(field.Name))

		// Pointer ke field struct untuk pendaftaran flag
		ptr := fieldVal.Addr().Interface()

		// PENTING: Gunakan fungsi NON-P (StringVar, IntVar, dll.)
		switch field.Type.Kind() {
		case reflect.String:
			// Diganti dari StringVarP menjadi StringVar
			cmd.Flags().StringVar(ptr.(*string), flagName, fieldVal.String(), usage)
		case reflect.Int:
			// Diganti dari IntVarP (jika digunakan) menjadi IntVar
			cmd.Flags().IntVar(ptr.(*int), flagName, int(fieldVal.Int()), usage)
		case reflect.Bool:
			// Diganti dari BoolVarP (jika digunakan) menjadi BoolVar
			cmd.Flags().BoolVar(ptr.(*bool), flagName, fieldVal.Bool(), usage)
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				defaultSlice := fieldVal.Interface().([]string)
				// Menggunakan StringSliceVar
				cmd.Flags().StringSliceVar(ptr.(*[]string), flagName, defaultSlice, usage)
			}
		default:
			// Anda dapat menambahkan logika error/warning di sini
		}
	}
	return nil
}
