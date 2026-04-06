param(
    [string]$VersionOverride = ""
)

$ErrorActionPreference = 'Stop'

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    $fallbackGo = 'C:\Program Files\Go\bin'
    if (Test-Path (Join-Path $fallbackGo 'go.exe')) {
        $env:Path = "$fallbackGo;$env:Path"
    }
}

Get-Command go | Out-Null
go version
go env GOPATH GOROOT

$version = if ($VersionOverride) { $VersionOverride } else { (Get-Content version.txt).Trim() }
$ldflags = "-X main.version=$version -s -w"

Write-Host "Running local verification..."
go test ./...
go build ./...

New-Item -ItemType Directory -Force -Path bin/windows/amd64 | Out-Null
New-Item -ItemType Directory -Force -Path bin/linux/amd64 | Out-Null
New-Item -ItemType Directory -Force -Path bin/linux/arm64 | Out-Null
New-Item -ItemType Directory -Force -Path bin/darwin/amd64 | Out-Null
New-Item -ItemType Directory -Force -Path bin/darwin/arm64 | Out-Null

Write-Host "Building windows/amd64..."
$env:GOOS = 'windows'; $env:GOARCH = 'amd64'; $env:CGO_ENABLED = '0'
go build -ldflags $ldflags -o bin/windows/amd64/azd-drasi.exe .

Write-Host "Building linux/amd64..."
$env:GOOS = 'linux'; $env:GOARCH = 'amd64'; $env:CGO_ENABLED = '0'
go build -ldflags $ldflags -o bin/linux/amd64/azd-drasi .

Write-Host "Building linux/arm64..."
$env:GOOS = 'linux'; $env:GOARCH = 'arm64'; $env:CGO_ENABLED = '0'
go build -ldflags $ldflags -o bin/linux/arm64/azd-drasi .

Write-Host "Building darwin/amd64..."
$env:GOOS = 'darwin'; $env:GOARCH = 'amd64'; $env:CGO_ENABLED = '0'
go build -ldflags $ldflags -o bin/darwin/amd64/azd-drasi .

Write-Host "Building darwin/arm64..."
$env:GOOS = 'darwin'; $env:GOARCH = 'arm64'; $env:CGO_ENABLED = '0'
go build -ldflags $ldflags -o bin/darwin/arm64/azd-drasi .

if (-not (Test-Path 'bin/windows/amd64/azd-drasi.exe')) { throw 'Missing windows artifact' }
if (-not (Test-Path 'bin/linux/amd64/azd-drasi')) { throw 'Missing linux amd64 artifact' }
if (-not (Test-Path 'bin/linux/arm64/azd-drasi')) { throw 'Missing linux arm64 artifact' }
if (-not (Test-Path 'bin/darwin/amd64/azd-drasi')) { throw 'Missing darwin amd64 artifact' }
if (-not (Test-Path 'bin/darwin/arm64/azd-drasi')) { throw 'Missing darwin arm64 artifact' }

Write-Host "Done. Artifacts in bin/"
