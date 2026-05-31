# generate-docs.ps1

$ErrorActionPreference = "Stop"

$DocsDir = "docs/gomarkdoc"
$OutputFile = "$DocsDir/API.md"

Write-Host "Creating docs directory..."
New-Item -ItemType Directory -Force -Path $DocsDir | Out-Null

Write-Host "Generating Go documentation..."
go run github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest ./... > $OutputFile

Write-Host "Done."
Write-Host "Documentation saved to: $OutputFile"