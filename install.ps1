# Configuration
$REPO_OWNER = "xingty"
$REPO_NAME = "rcode-go"
$INSTALL_DIR = $env:USERPROFILE
$BIN_PATH = "$($env:USERPROFILE)\gcode\bin"  

function Print-Status { 
    param([string]$Message) 
    Write-Host "[STATUS] $Message" -ForegroundColor Green 
}

function Print-Error { 
    param([string]$Message) 
    Write-Host "[ERROR] $Message" -ForegroundColor Red; 
    exit 1 
}

function Print-Warning { 
    param([string]$Message) 
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow 
}


function Detect-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE.ToLower()
    
    switch -Wildcard ($arch) {
        "amd64" { return "amd64" }
        "arm64" { return "arm64" }
        "arm*" { return "arm" }
        "x86" { return "386" }
        default { return "unknown" }
    }
}

function Compare-Version {
    param(
        [string]$Version1,
        [string]$Version2
    )
    
    $v1 = $Version1.Split('.')
    $v2 = $Version2.Split('.')
    
    $maxLength = [Math]::Max($v1.Length, $v2.Length)
    
    for ($i = 0; $i -lt $maxLength; $i++) {
        $num1 = if ($i -lt $v1.Length) { [int]$v1[$i] } else { 0 }
        $num2 = if ($i -lt $v2.Length) { [int]$v2[$i] } else { 0 }
        
        if ($num1 -gt $num2) {
            return 1
        }
        elseif ($num1 -lt $num2) {
            return 2
        }
    }
    
    return 0
}

$ARCH = Detect-Architecture

Print-Status "Detected platform: $PLATFORM, architecture: $ARCH"

if (Get-Command gcode -ErrorAction SilentlyContinue) {
    Print-Status "gcode is already installed. Checking for updates..."
    $CURRENT_VERSION = (gcode -v | Select-String -Pattern '\d+\.\d+\.\d+').Matches.Value
    
    if (-not $CURRENT_VERSION) {
        Print-Warning "Could not determine current version. Will proceed with installation."
    }
    else {
        Print-Status "Current version: $CURRENT_VERSION"
    }
}
else {
    Print-Status "gcode is not installed. Will proceed with installation."
    $CURRENT_VERSION = ""
}

Print-Status "Fetching the latest version information..."

# Get the latest version from VERSION file
$VERSION_URL = "https://raw.githubusercontent.com/$REPO_OWNER/$REPO_NAME/refs/heads/main/VERSION"
try {
    $VERSION = (Invoke-WebRequest -Uri $VERSION_URL -UseBasicParsing).Content.Trim()
}
catch {
    Print-Error "Failed to fetch version information. Please check your internet connection."
}

# Validate version format
if (-not ($VERSION -match '^\d+\.\d+\.\d+$')) {
    Print-Error "Invalid version format: $VERSION"
}

$TAG_NAME = "v$VERSION"
Print-Status "Found latest version: $TAG_NAME"

if ($CURRENT_VERSION) {
    $cmp = Compare-Version -Version1 $CURRENT_VERSION -Version2 $VERSION
    if ($cmp -eq 0) {
        Print-Status "gcode is already at the latest version. No update necessary."
        exit 0
    }
    elseif ($cmp -eq 1) {
        Print-Status "Installed gcode version ($CURRENT_VERSION) is ahead of the latest release ($VERSION)."
        exit 0
    }
    else {
        Print-Status "Update available: $CURRENT_VERSION -> $VERSION"
    }
}

$FILE_PATTERN = "gcode-$TAG_NAME-windows-$ARCH.tar.gz"
Print-Status "Looking for file matching pattern: $FILE_PATTERN"

$ASSET_URL = "https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$TAG_NAME/$FILE_PATTERN"

Print-Status "Downloading release file from: $ASSET_URL"

# Create a temporary directory
$TMP_DIR = Join-Path -Path $env:TEMP -ChildPath ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $TMP_DIR | Out-Null

try {
    # Download the asset
    $randomName = [System.IO.Path]::GetRandomFileName().Split('.')[0]
    $TARBALL = Join-Path -Path $TMP_DIR -ChildPath "$randomName.tar.gz"
    Print-Status "Downloading release file to: $TARBALL"
    # $TARBALL = Join-Path -Path $TMP_DIR -ChildPath "release.tar.gz"
    Invoke-WebRequest -Uri $ASSET_URL -OutFile $TARBALL -UseBasicParsing
    
    Print-Status "Download complete. Extracting to $INSTALL_DIR..."
    
    # Extract the tarball to the install directory
    if (-not (Test-Path -Path $INSTALL_DIR)) {
        New-Item -ItemType Directory -Path $INSTALL_DIR | Out-Null
    }
    
    if ($PSVersionTable.PSVersion.Major -ge 5) {
        # PowerShell 5+ has Expand-Archive
        $tempExtract = Join-Path -Path $TMP_DIR -ChildPath "extracted"
        New-Item -ItemType Directory -Path $tempExtract | Out-Null
        tar -xzf $TARBALL -C $tempExtract
        Move-Item -Path "$tempExtract\*" -Destination $INSTALL_DIR -Force
    }
    else {
        # Fallback for older PowerShell versions
        tar -xzf $TARBALL -C $INSTALL_DIR
    }
}
catch {
    Print-Error "Failed to download or extract the release file: $_"
}

# Clean up
Remove-Item -Path $TMP_DIR -Recurse -Force

Print-Status "Installation completed successfully!"

# Check if the bin directory exists
if (-not (Test-Path -Path $BIN_PATH)) {
    Print-Warning "The bin directory $BIN_PATH does not exist. Please check the extracted contents."
}
else {
    Print-Status "Please add the following to your PowerShell profile:"
    Write-Host "`$env:Path += `";$BIN_PATH`""
    
    $profilePath = $PROFILE.CurrentUserAllHosts
    if (-not (Test-Path -Path $profilePath)) {
        New-Item -ItemType File -Path $profilePath -Force | Out-Null
    }
    
    Print-Status "You can do this by running:"
    Write-Host "Add-Content -Path `"$profilePath`" -Value '`$env:Path += `";$BIN_PATH`"'"
    Write-Host ". `"$profilePath`""
    
    Print-Status "After updating your PATH, you can run the program by typing 'gcode' in your terminal."
}

Print-Status "Installation process complete!"