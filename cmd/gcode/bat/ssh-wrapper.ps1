param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ArgsFromCaller
)

$useGssh = $false
$sshArgs = New-Object System.Collections.Generic.List[string]
foreach ($arg in $ArgsFromCaller) {
    if ($arg -eq "--gssh") {
        $useGssh = $true
        continue
    }
    [void]$sshArgs.Add($arg)
}

if (-not $useGssh) {
    & ssh @sshArgs
    exit $LASTEXITCODE
}

$gsshCore = Join-Path $PSScriptRoot "gssh-core.exe"
$gssh = if (Test-Path $gsshCore) { $gsshCore } else { "gssh-core" }

$preparedOutput = & $gssh "__prepare" @sshArgs
$prepareExitCode = $LASTEXITCODE
if ($prepareExitCode -ne 0) {
    exit $prepareExitCode
}

if ([string]::IsNullOrWhiteSpace($preparedOutput)) {
    Write-Error "gssh-core __prepare returned empty output"
    exit 255
}

try {
    $prepared = $preparedOutput | ConvertFrom-Json
}
catch {
    Write-Error "failed to parse gssh-core __prepare output as JSON: $($_.Exception.Message)"
    exit 255
}

$program = if ($prepared.program) { [string]$prepared.program } else { "ssh" }
$preparedArgs = @()
if ($prepared.args) {
    $preparedArgs = @($prepared.args)
}

& $program @preparedArgs
exit $LASTEXITCODE
