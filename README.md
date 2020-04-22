# AutoDB

## How to run Docker container.
1. Download Docker
1. Open a terminal in the project folder and type `docker-compose build`
 to build project.
1. Then type `docker-compose up mysql` to start MySQL server.
1. After the MySQL server has finished starting open a new terminal and
 type `docker-compose up server` to start go server.

### How to Setup without Docker.
1. setup a MySQL Server.
1. setup a DBA user for that Server.
1. run sql/hostScheme.sql on the MySQL Server.
1. write the proper DB url and user credential in host/dbconfig/secret.go
1. build autodb/host and put the executable in bin/
1. run the executable

## Tests
1. run sql/testData.sql
1. run go test.


