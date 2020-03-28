package dbconfig

import (
	"database/sql"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"strconv"
	"time"
)

func PrepareNamedQuery(q string, arg interface{}) (string, []interface{}, error) {
	query, args, err := sqlx.Named(q, arg)
	if err != nil {
		return "", nil, err
	}
	query, args, err = sqlx.In(query, args...)
	return query, args, err
}

//$culumntype	mysql.ColumnType.DatabaseColumnType format_in_json
//VARCHAR(%d)	"VARCHAR"							string (raw)
//MEDIUMTEXT	"MEDIUMTEXT"						string (raw)
//INT			"INT"								int
//BIGINT		"BIGINT"							string (json cannot handle int64 very well, let's leave it to frontend)
//DOUBLE		"DOUBLE"							double
//DATETIME		"DATETIME"							string (Time in RFC 3339 format in UTC)
//ENUM			"ENUM"								string (raw)
//SET			"SET"								string (raw)
// and don't forget about null value, which will make v be a real nil (not a nil value for another type).
func interpretColumnType(colType *sql.ColumnType , v interface{}) interface{} {

	//check for null
	vBytes , ok:=v.([]byte)
	if !ok {
		return v
	}

	switch colType.DatabaseTypeName() {
	case "INT":
		var vi32 int32
		vi64, _ := strconv.ParseInt(string(vBytes), 0, 32)
		vi32 = int32(vi64)
		return vi32
	case "BIGINT":
		var vi64 int64
		vi64, _ = strconv.ParseInt(string(vBytes), 0, 64)
		vsi64 := strconv.FormatInt(vi64, 10)
		return vsi64
	case "DOUBLE":
		var vf64 float64
		vf64, _ = strconv.ParseFloat(string(vBytes), 64)
		return vf64
	case "DATETIME": //convert the format to RFC 3339 format
		//pitfall: the time we retrieve from server will be convert to local timezone without a timezone identifier.
		var vNullTime mysql.NullTime
		_ = vNullTime.Scan(v)
		localT := time.Now()
		_, offset := localT.Zone()
		vTime := vNullTime.Time.Add(time.Duration(-offset) * time.Second)
		return vTime
	default:
		return string(v.([]byte))
	}

}



func ParseRowsToJSON (rows *sql.Rows) (string, error) { //need test for all types including null
	result := make([][]interface{}, 0, 50)
	cols, err:=rows.Columns()
	if err!=nil {
		return "", err
	}
	colTypes, err:= rows.ColumnTypes()

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
			result[l][i] = interpretColumnType(colTypes[i], *rowResult[i])
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