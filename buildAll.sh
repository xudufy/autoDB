cd $(dirname $0)
mkdir -p bin
go build -o autodb autodb/app
