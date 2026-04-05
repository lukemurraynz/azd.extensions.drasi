$ErrorActionPreference = 'Stop'

$version = (Get-Content version.txt).Trim()
$ldflags = "-X main.version=$version -s -w"

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

Write-Host "Done. Artifacts in bin/"
