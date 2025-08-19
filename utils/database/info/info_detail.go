package info

import (
	"database/sql"
	"log"
	"strings"
)

func GetBaseTables(db *sql.DB) []string {
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'")
	if err != nil {
		log.Printf("Gagal ambil daftar tabel: %v", err)
		return nil
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil {
			tables = append(tables, t)
		}
	}
	return tables
}

func GetRowCount(db *sql.DB, table string) int64 {
	var count int64
	db.QueryRow("SELECT COUNT(*) FROM `" + table + "`").Scan(&count)
	return count
}

func GetPrimaryKey(db *sql.DB, table string) string {
	rows, err := db.Query("SHOW KEYS FROM `" + table + "` WHERE Key_name = 'PRIMARY'")
	if err != nil {
		return ""
	}
	defer rows.Close()

	var pkCols []string
	for rows.Next() {
		var table, nonUnique, keyName, seqInIndex, columnName, collation, cardinality, subPart, packed, null, indexType, comment, indexComment sql.NullString
		err := rows.Scan(&table, &nonUnique, &keyName, &seqInIndex, &columnName, &collation, &cardinality, &subPart, &packed, &null, &indexType, &comment, &indexComment)
		if err == nil && columnName.Valid {
			pkCols = append(pkCols, "`"+columnName.String+"`")
		}
	}
	return strings.Join(pkCols, ", ")
}

func GetViewNames(db *sql.DB) []string {
	rows, err := db.Query("SELECT table_name FROM information_schema.views WHERE table_schema = DATABASE()")
	if err != nil {
		log.Printf("Gagal ambil daftar view: %v", err)
		return nil
	}
	defer rows.Close()

	var views []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err == nil {
			views = append(views, v)
		}
	}
	return views
}
