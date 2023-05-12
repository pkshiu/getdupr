set dotenv-load := true
DB_PATH := "dupr.sqlite"

build:
	go build -o bin/getdupr *.go

buildwin:
	env GOOS=windows GOARCH=amd64 go build -o bin/getdupr.exe *.go


run:
	go run *.go


