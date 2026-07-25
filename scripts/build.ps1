# macOS
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o bin/pyman_darwin_arm64
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o bin/pyman_darwin_amd64

# Linux
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/pyman_linux_amd64

# Windows
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o bin/pyman_windows_amd.exe
