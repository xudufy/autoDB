package dbconfig

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"github.com/jmoiron/sqlx"
)

func PrepareNamedQuery(q string, arg interface{}) (string, []interface{}, error) {
	query, args, err := sqlx.Named(q, arg)
	if err != nil {
		return "", nil, err
	}
	query, args, err = sqlx.In(query, args...)
	return query, args, err
}

//type rowValue struct {
//	data interface{}
//}

// Scan assigns a value from a database driver.
// See https://golang.org/pkg/database/sql/#Scanner
// The src value will be of one of the following types:
//
//    int64
//    float64
//    bool
//    []byte
//    string
//    time.Time
//    nil - for NULL values
//
// An error should be returned if the value cannot be stored
// without loss of information.
//
// Reference types such as []byte are only valid until the next call to Scan
// and should not be retained. Their underlying memory is owned by the driver.
// If retention is necessary, copy their values before the next call to Scan.
//
// in encoding/json, we can handle all of the types except []byte,
//func (v *rowValue) Scan(src interface{}) error {
//	switch src.(type) {
//	case []byte:
//		v.data = base64.StdEncoding.EncodeToString(src.([]byte))
//	default:
//		v.data = src
//	}
//	return nil
//}

func ParseRowsToJSON (rows *sql.Rows) (string, error) { //need test
	result := make([][]interface{}, 0, 50)
	cols, err:=rows.Columns()
	if err!=nil {
		return "", err
	}
	result = append(result, make([]interface{}, len(cols)))
	l := len(result) - 1
	for i := range result[l] {
		result[l][i] = cols[i]
	}
	rowResult := make([]*interface{}, len(cols))
	interfaceResult := make ([]interface{}, len(cols))
	for i := range rowResult {
		rowResult[i]=new(interface{})
		interfaceResult[i]=rowResult[i]
	}
	for rows.Next() {
		err:=rows.Scan(interfaceResult...)
		if err!=nil {
			return "", err
		}
		result = append(result, make([]interface{}, len(cols)))
		l := len(result) - 1
		for i := range result[l] {
			switch (*rowResult[i]).(type) {
			case []byte:
				*rowResult[i] = base64.StdEncoding.EncodeToString((*rowResult[i]).([]byte))
			}
			result[l][i] = *(rowResult[i])
		}
	}
	resultMap:=make(map[string][][]interface{})
	resultMap["result"] = result
	resultJSON, err := json.Marshal(resultMap)
	if err!=nil {
		panic(err)
	}
	return string(resultJSON), nil
}