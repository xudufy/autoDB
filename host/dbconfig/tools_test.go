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
	fmt.Println(string(js))
	ans := `{"result":[["id","intid","time","nulltime","nullable","nullstring"],["1",14,"2020-03-29T12:00:00Z",null,null,null],["2147483649",14,"2020-03-28T12:00:00Z",null,null,"?? ?"]]}`
	if string(js)!=ans {
		t.Errorf("answer is not right")
	}

}

func TestParseRowsToJSON2(t *testing.T) {
	Init()
	rows, err := HostDB.Query(`describe nulltest;`)
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
	fmt.Println(string(js))
}

func TestParseRowsToJSON4(t *testing.T) {
	Init()
	rows, err := HostDB.Query(`show indexes from nulltest;`)
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
	fmt.Println(string(js))
}

func TestParseRowsToJSON3(t *testing.T) {
	Init()
	rows, err := HostDB.Query(`insert into nullTest (id, time) values (11, '2020-03-29 08:00:00');describe nulltest;`)
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
	fmt.Println(string(js))
}

func TestNamedParameterizedQuery(t *testing.T) {
	a := map[string]interface{}{
		"n_1":"id",
		"n2":"intid",
		"n3":"time",
		"n4":"84",
		"n5":100,
		"n6":300.001,
	}

	q := []string {
		`:`,
		`::`,
		`"?"`,
		`'?'`,
		`:n_1 :: :n4:::n5 ::::\:n6?`,
		":n7",
		`":n7"`,
		`:n_1 :: 127,mr2h\vea :n4 :::n5 ::::\:n6`,
		`:n_1 :: 127,mr2h\vea :n4 :::n5 ::::\:`,
		`:n_1 ":n2" "':n3" "'':n3" ":n4"":n5"`,
		`:n_1 "":n2"" "\":n3":n4""":n5"`,
		`':n_1 '':n2'' '\':n3'':n4''':n5'`,
	}

	for _, qs := range q {
		a1, a2, err := PrepareNamedQuery(qs,a)
		if err!=nil {
			fmt.Println(err)
		}
		fmt.Println(a1)
		fmt.Println(a2)
		fmt.Println("------------")
	}

	for _, qs := range q {
		a1, a2, err := ParseNamedQuery(qs)
		if err!=nil {
			fmt.Println(err)
		}
		fmt.Println(a1)
		fmt.Println(a2)
		fmt.Println("------------")
	}
	/* correct answer:
	At query:1: expect ':' or identifier.

	[]
	------------
	:
	[]
	------------
	"?"
	[]
	------------
	'?'
	[]
	------------
	? outside string literals found in query:25.

	[]
	------------
	n7 is not provided.

	[]
	------------
	":n7"
	[]
	------------
	? : 127,mr2h\vea ? :? ::\?
	[id 84 100 300.001]
	------------
	At query:37: expect ':' or identifier.

	[]
	------------
	? ":n2" "':n3" "'':n3" ":n4"":n5"
	[id]
	------------
	? ""?"" "\":n3"?""":n5"
	[id intid 84]
	------------
	':n_1 '':n2'' '\'?''?''':n5'
	[time 84]
	------------
	At query:1: expect ':' or identifier.

	[]
	------------
	:
	[]
	------------
	"?"
	[]
	------------
	'?'
	[]
	------------
	? outside string literals found in query:25.

	[]
	------------
	?
	[n7]
	------------
	":n7"
	[]
	------------
	? : 127,mr2h\vea ? :? ::\?
	[n_1 n4 n5 n6]
	------------
	At query:37: expect ':' or identifier.

	[]
	------------
	? ":n2" "':n3" "'':n3" ":n4"":n5"
	[n_1]
	------------
	? ""?"" "\":n3"?""":n5"
	[n_1 n2 n4]
	------------
	':n_1 '':n2'' '\'?''?''':n5'
	[n3 n4]
	------------
	*/
}