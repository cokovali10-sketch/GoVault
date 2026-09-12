@echo off
echo Building GoVault...
go build -o govault.exe -ldflags="-s -w" .
echo Done: govault.exe
