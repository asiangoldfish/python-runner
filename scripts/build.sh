# macOS
GOOS="darwin" GOARCH="arm64" go build -o bin/pyman_darwin_arm64 cmd/pyrun/pyrun.go
GOOS="darwin" GOARCH="amd64" go build -o bin/pyman_darwin_amd64 cmd/pyrun/pyrun.go

# Linux
GOOS="linux" GOARCH="amd64" go build -o bin/pyman_linux_amd64 cmd/pyrun/pyrun.go

# Windows
GOOS="windows" GOARCH="amd64" go build -o bin/pyman_windows_amd.exe cmd/pyrun/pyrun.go
