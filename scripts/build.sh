# macOS
GOOS="darwin" GOARCH="arm64" go build -o bin/pyman_darwin_arm64
GOOS="darwin" GOARCH="amd64" go build -o bin/pyman_darwin_amd64

# Linux
GOOS="linux" GOARCH="amd64" go build -o bin/pyman_linux_amd64

# Windows
GOOS="windows" GOARCH="amd64" go build -o bin/pyman_windows_amd.exe
