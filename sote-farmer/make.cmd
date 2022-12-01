@echo off

SET BINARY_NAME=sote-farmer.exe
SET GOPATH=%USERPROFILE%\go;%CD%

if exist bin\%BINARY_NAME% del /f bin\%BINARY_NAME%

go mod tidy
go build -o bin/%BINARY_NAME%  .\main.go