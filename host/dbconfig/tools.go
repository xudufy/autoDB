package dbconfig

import "github.com/jmoiron/sqlx"

func PrepareNamedQuery(q string, arg interface{}) (string, []interface{}, error) {
	query, args, err := sqlx.Named(q, arg)
	if err != nil {
		return "", nil, err
	}
	query, args, err = sqlx.In(query, args...)
	return query, args, err
}
