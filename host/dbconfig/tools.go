package dbconfig

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func IsIdentifierLetter(ch rune) bool {
	return (ch>='0' && ch<='9') || (ch>='a' && ch<='z') || (ch>='A' && ch<='Z') || ch=='_'
}

func IsIdentifier(s string) bool {
	firstRune, _ := utf8.DecodeRuneInString(s)
	if unicode.IsNumber(firstRune) {
		return false
	}

	for _ , ch := range s {
		if !IsIdentifierLetter(ch) {
			return false
		}
	}

	return true
}

func PrepareNamedQuery(query string, args map[string]interface{}) (string, []interface{}, error) {

	q,a,err := ParseNamedQuery(query)
	if err!=nil {
		return "", nil, err
	}

	resultArgs := make([]interface{}, 0, len(a))

	for _, thisName:= range a {
		paraValue, ok := args[thisName]
		if !ok {
			return "", nil, errors.New(thisName + " is not provided.")
		}
		resultArgs = append(resultArgs, paraValue)
	}

	return q, resultArgs, nil

}

//because the source code of sqlx.Named() and sqlx.In() have major bugs,
//I decided to write it manually.
//here I only consider the colon in string literal and colon escape('::') outside the string literal is not a named
//parameter in one query.
//and '::' outside the string literal will be replaced as ':' in returned query.
//this function parse the query to ? parameterized and find out the list of args' name in that query.
func ParseNamedQuery(q string) (string, []string, error) {
	inStringLiteral := false
	inBackSlashEscape := false
	inQuoteEscapeInStringLiteral := false
	inColonEscape := false
	inName := false
	currentQuote := '"'

	var currentName strings.Builder
	var resultQuery strings.Builder
	resultArgs := make([]string, 0, 5)

	lenQ := 0
	for i, ch := range q {
		if i>lenQ {
			lenQ = i
		}
		if inQuoteEscapeInStringLiteral && ch!=currentQuote {
			inStringLiteral = false
			inQuoteEscapeInStringLiteral = false
		}

		if inColonEscape && ch!=':'{
			inColonEscape = false
		}

		if inName && IsIdentifierLetter(ch) {
			inColonEscape=false
			currentName.WriteRune(ch)
			continue
		}

		if ch=='?' && !inStringLiteral {
			return "",nil, errors.New("? outside string literals found in query:"+strconv.Itoa(i)+".")
		}

		if inName && !IsIdentifierLetter(ch) {
			thisName := currentName.String()
			currentName.Reset()
			if thisName == "" {
				if ch!=':' {
					return "", nil, errors.New("At query:" + strconv.Itoa(i) + ": expect ':' or identifier.")
				}
			} else {
				resultArgs = append(resultArgs, thisName)
				resultQuery.WriteRune('?')
				inName = false
			}
		}

		switch ch {
		case '\'':
			fallthrough
		case '"':
			if inBackSlashEscape {
				inBackSlashEscape = false
				resultQuery.WriteRune(ch)
				break
			}
			if !inStringLiteral {
				inStringLiteral = true
				currentQuote = ch
				resultQuery.WriteRune(ch)
				break
			} else if currentQuote == ch {
				if inQuoteEscapeInStringLiteral {
					inQuoteEscapeInStringLiteral = false
					resultQuery.WriteRune(ch)
					break
				} else {
					inQuoteEscapeInStringLiteral = true
					resultQuery.WriteRune(ch)
					break
				}
			} else {
				resultQuery.WriteRune(ch)
				break
			}
		case '\\':
			if inBackSlashEscape {
				inBackSlashEscape = false
				resultQuery.WriteRune(ch)
				break
			} else {
				inBackSlashEscape = true
				resultQuery.WriteRune(ch)
				break
			}
		case ':':
			if inBackSlashEscape {
				inBackSlashEscape = false
			}
			if inStringLiteral {
				resultQuery.WriteRune(ch)
				break
			}
			if inColonEscape {
				resultQuery.WriteRune(ch)
				inColonEscape = false
				inName = false
				break
			} else {
				inName = true
				inColonEscape = true
				break
			}
		default:
			if inBackSlashEscape {
				inBackSlashEscape = false
			}
			//inName has been processed above.
			resultQuery.WriteRune(ch)
		}
	}

	// check if the named parameter is at the end of query string.
	if inName {
		thisName := currentName.String()
		currentName.Reset()
		if thisName == "" {
			return "", nil, errors.New("At query:" + strconv.Itoa(lenQ+1) + ": expect ':' or identifier.")
		} else {
			resultArgs = append(resultArgs, thisName)
			resultQuery.WriteRune('?')
			inName = false
		}
	}

	return resultQuery.String(), resultArgs, nil
}

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

func ParseRowsToJSON (rows *sql.Rows) ([]byte, error) { //need test for all types including null
	result := make([][]interface{}, 0, 50)
	cols, err:=rows.Columns()
	if err!=nil {
		return nil, err
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
			return nil, err
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
	return resultJSON, nil
}