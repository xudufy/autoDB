# AutoDB

### How to Setup
1. setup a MySQL Server.
1. setup a DBA user for that Server.
1. run sql/hostScheme.sql on the MySQL Server.
1. write the proper DB url and user credential in host/dbconfig/secret.go
1. build autodb/host and put the executable in bin/
1. run the executable

## Tests
1. run sql/testData.sql
1. run go test.
