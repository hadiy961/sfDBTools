package query

import (
	"database/sql"
	"fmt"
	"strings"

	dbwrap "sfDBTools/utils/database"
)

// Insert inserts a row into the given table using the provided data map.
// keys of data become columns; values are parameters.
func Insert(cfg dbwrap.Config, table string, data map[string]interface{}) (sql.Result, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided")
	}

	cols := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data))

	for k, v := range data {
		cols = append(cols, fmt.Sprintf("`%s`", k))
		placeholders = append(placeholders, "?")
		args = append(args, v)
	}

	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", table, strings.Join(cols, ","), strings.Join(placeholders, ","))

	db, err := dbwrap.GetDatabaseConnection(cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return db.Exec(query, args...)
}

// Select executes a SELECT query on the given table with optional columns and where clause.
// whereClause should not include the WHERE keyword (e.g. "id = ? AND active = ?").
// Returns a slice of maps (column name -> value).
func Select(cfg dbwrap.Config, table string, columns []string, whereClause string, args ...interface{}) ([]map[string]interface{}, error) {
	cols := "*"
	if len(columns) > 0 {
		esc := make([]string, 0, len(columns))
		for _, c := range columns {
			esc = append(esc, fmt.Sprintf("`%s`", c))
		}
		cols = strings.Join(esc, ",")
	}

	q := fmt.Sprintf("SELECT %s FROM `%s`", cols, table)
	if strings.TrimSpace(whereClause) != "" {
		q = q + " WHERE " + whereClause
	}

	db, err := dbwrap.GetDatabaseConnection(cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	colsNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)

	for rows.Next() {
		vals := make([]interface{}, len(colsNames))
		valPtrs := make([]interface{}, len(colsNames))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}

		if err := rows.Scan(valPtrs...); err != nil {
			return nil, err
		}

		m := make(map[string]interface{}, len(colsNames))
		for i, col := range colsNames {
			v := vals[i]
			// convert []byte to string for readability
			if b, ok := v.([]byte); ok {
				m[col] = string(b)
			} else {
				m[col] = v
			}
		}
		result = append(result, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetByID returns a single row by id column and value. columns is optional; pass nil for all columns.
func GetByID(cfg dbwrap.Config, table string, idCol string, id interface{}, columns []string) (map[string]interface{}, error) {
	where := fmt.Sprintf("%s = ?", idCol)
	rows, err := Select(cfg, table, columns, where, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	return rows[0], nil
}

// UpdateByID updates the row identified by idCol=id with the provided data map.
func UpdateByID(cfg dbwrap.Config, table string, idCol string, id interface{}, data map[string]interface{}) (sql.Result, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided")
	}

	sets := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data)+1)
	for k, v := range data {
		sets = append(sets, fmt.Sprintf("`%s` = ?", k))
		args = append(args, v)
	}
	args = append(args, id)

	query := fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s` = ?", table, strings.Join(sets, ","), idCol)

	db, err := dbwrap.GetDatabaseConnection(cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return db.Exec(query, args...)
}

// DeleteByID deletes a row identified by idCol=id.
func DeleteByID(cfg dbwrap.Config, table string, idCol string, id interface{}) (sql.Result, error) {
	query := fmt.Sprintf("DELETE FROM `%s` WHERE `%s` = ?", table, idCol)

	db, err := dbwrap.GetDatabaseConnection(cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return db.Exec(query, id)
}
