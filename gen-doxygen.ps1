$ErrorActionPreference = "Stop"

$ProjectDir = Split-Path -Parent $MyInvocation.MyCommand.Path

$Doxyfile = Join-Path $ProjectDir "Doxyfile"

if (-not (Test-Path $Doxyfile)) {
    Write-Host "Brak Doxyfile, generuję domyślny..."
    Push-Location $ProjectDir
    doxygen -g
    Pop-Location
}

Write-Host "Generuję dokumentację..."

Push-Location $ProjectDir
doxygen $Doxyfile
Pop-Location

Write-Host "Gotowe."