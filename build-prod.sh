CC=x86_64-w64-mingw32-gcc GOOS=windows CGO_ENABLED=1 GOARCH=amd64 go build -ldflags "-w -s -H windowsgui" -o prod.exe