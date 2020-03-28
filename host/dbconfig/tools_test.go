package dbconfig

import (
	"fmt"
	"testing"
)

func TestParseRowsToJSON(t *testing.T) {
	Init()
	rows, err := HostDB.Query("select * from nullTest")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	js, err:=ParseRowsToJSON(rows)
	if err!=nil {
		fmt.Println(err)
		return
	}
	fmt.Println(js)
	ans := `{"result":[["id","intid","time","nulltime","nullable"],["1",14,"2020-03-29T12:00:00Z",null,null],["2147483649",14,"2020-03-28T12:00:00Z",null,null]]}`
	if js!=ans {
		t.Errorf("answer is not right")
	}

}