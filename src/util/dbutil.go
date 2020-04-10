package util

import (
	"crypto/md5"
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
)

func ScanInterface(rows *sql.Rows) ([]interface{}, []interface{}, error) {
	col, err := rows.Columns()
	if col == nil {
		return nil, nil, errors.New("nothing in the result set")
	}
	if err != nil {
		return nil, nil, err
	}
	res := make([]interface{}, 0)
	realRes := make([]interface{}, 0)
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
			return nil, nil, err
		}
		res = append(res, rowHash)
		realRes = append(realRes, row)
	}
	return res, realRes, nil
}
