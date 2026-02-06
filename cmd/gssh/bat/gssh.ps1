param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ArgsFromCaller
)

$gsshCore = Join-Path $PSScriptRoot "gssh-core.exe"
if (-not (Test-Path $gsshCore)) {
    $gsshCore = "gssh-core"
}

# Prefer core for gssh's own help/version. `-h` is reserved for ssh passthrough.
if ($ArgsFromCaller.Count -eq 1) {
    if ($ArgsFromCaller[0] -eq "--help") {
        & $gsshCore "--help"
        exit $LASTEXITCODE
    }
    if ($ArgsFromCaller[0] -eq "--version") {
        & $gsshCore "--version"
        exit $LASTEXITCODE
    }
}

$preparedOutput = & $gsshCore "__prepare" @ArgsFromCaller
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
