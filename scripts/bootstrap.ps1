param()

$ErrorActionPreference = 'Stop'

function Read-ToolVersions {
  $map = @{}
  Get-Content '.tool-versions' | ForEach-Object {
    if (-not [string]::IsNullOrWhiteSpace($_)) {
      $parts = $_ -split '\s+', 2
      if ($parts.Length -eq 2) {
        $map[$parts[0]] = $parts[1]
      }
    }
  }
  return $map
}

function Ensure-CommandVersion {
  param(
    [string]$Command,
    [string]$VersionArgs,
    [string]$Expected,
    [string]$DisplayName
  )

  if (-not (Get-Command $Command -ErrorAction SilentlyContinue)) {
    throw "$DisplayName is required but not installed. Expected version $Expected."
  }

  $current = (& $Command $VersionArgs) 2>$null | Select-Object -First 1
  if (-not $current) {
    throw "Unable to determine $DisplayName version. Expected $Expected."
  }

  if ($current -notmatch [regex]::Escape($Expected)) {
    throw "$DisplayName version mismatch. Expected $Expected, got: $current"
  }
}

$tools = Read-ToolVersions

if (Get-Command mise -ErrorAction SilentlyContinue) {
  Write-Host 'Installing pinned tools with mise...'
  & mise install
} elseif (Get-Command asdf -ErrorAction SilentlyContinue) {
  Write-Host 'Installing pinned tools with asdf...'
  & asdf install
} else {
  Write-Host 'No supported version manager detected. Validating locally installed tools against .tool-versions...'
}

if ($tools.ContainsKey('nodejs')) {
  Ensure-CommandVersion -Command 'node' -VersionArgs '--version' -Expected $tools['nodejs'] -DisplayName 'Node.js'
  Ensure-CommandVersion -Command 'npm' -VersionArgs '--version' -Expected '11.8.0' -DisplayName 'npm'
}

if ($tools.ContainsKey('golang')) {
  Ensure-CommandVersion -Command 'go' -VersionArgs 'version' -Expected $tools['golang'] -DisplayName 'Go'
}

if ($tools.ContainsKey('terraform')) {
  Ensure-CommandVersion -Command 'terraform' -VersionArgs 'version' -Expected $tools['terraform'] -DisplayName 'Terraform'
}

if ($tools.ContainsKey('buf')) {
  Ensure-CommandVersion -Command 'buf' -VersionArgs '--version' -Expected $tools['buf'] -DisplayName 'Buf'
}

if (Test-Path 'package.json') {
  Write-Host 'Installing npm dependencies...'
  & npm ci
}

if (Test-Path 'go.mod') {
  Write-Host 'Downloading Go module dependencies...'
  & go mod download
}

Write-Host 'Bootstrap completed.'
