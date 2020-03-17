package util

import (
	"crypto/md5"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func ScanInterface(rows *sql.Rows) ([]interface{}, error) {
	col, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	res := make([]interface{}, 0)
	colNum := len(col)
	for rows.Next() {
		row := make([]interface{}, colNum)
		for idx, _ := range col {
			row[idx] = new([]byte)
		}
		err := rows.Scan(row...)
		var byteRow []byte
		for _, cell := range row {
			byteCell := *(cell.(*[]byte))
			byteRow = append(byteRow, byteCell...)
		}
		rowHash := md5.Sum(byteRow)
		if err != nil {
			return nil, err
		}
		res = append(res, rowHash)
	}
	return res, nil
}
