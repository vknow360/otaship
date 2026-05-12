param(
    [string]$Version = "release"
)

$BuildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$LDFlags = "-X github.com/vknow360/otaship/cli/internal/commands.Version=$Version -X github.com/vknow360/otaship/cli/internal/commands.BuildDate=$BuildDate"

New-Item -ItemType Directory -Force -Path dist | Out-Null
Write-Host "Building release binaries for v$Version" -ForegroundColor Green

$env:GOOS="linux"; $env:GOARCH="amd64"; go build -ldflags $LDFlags -o dist/otaship-linux-amd64 cmd/otaship/main.go
$env:GOOS="linux"; $env:GOARCH="arm64"; go build -ldflags $LDFlags -o dist/otaship-linux-arm64 cmd/otaship/main.go
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -ldflags $LDFlags -o dist/otaship-darwin-amd64 cmd/otaship/main.go
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -ldflags $LDFlags -o dist/otaship-darwin-arm64 cmd/otaship/main.go
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -ldflags $LDFlags -o dist/otaship-windows-amd64.exe cmd/otaship/main.go

Write-Host "✓ Binaries built in dist/" -ForegroundColor Green
Get-ChildItem dist/