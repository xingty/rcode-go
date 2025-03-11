#Requires -Version 5.0

# Configuration
$RepoOwner = "xingty"
$RepoName = "rcode-go"
$InstallDir = "$env:USERPROFILE"

# Text colors for console output
function Write-Status { param([string]$Message) Write-Host "[STATUS] $Message" -ForegroundColor Green }
function Write-Error1 { param([string]$Message) Write-Host "[ERROR] $Message" -ForegroundColor Red; exit 1 }
function Write-Warning1 { param([string]$Message) Write-Host "[WARNING] $Message" -ForegroundColor Yellow }

# Initial banner
Write-Status "GCode Installation Script for Windows"


# Detect architecture
$Arch = "unknown"
if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64" -or $env:PROCESSOR_ARCHITEW6432 -eq "AMD64") {
    $Arch = "amd64"
} elseif ($env:PROCESSOR_ARCHITECTURE -eq "x86" -and -not $env:PROCESSOR_ARCHITEW6432) {
    $Arch = "386"
} elseif ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $Arch = "arm64"
}

if ($Arch -eq "unknown") {
    Write-Error1 "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE. Cannot determine appropriate download."
}

Write-Status "Detected platform: windows, architecture: $Arch"

# Create temporary directory
$TempDir = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), "gcode_install_$([System.Guid]::NewGuid().ToString())")
try {
    New-Item -ItemType Directory -Path $TempDir -Force | Out-Null
} catch {
    Write-Error1 "Failed to create temporary directory: $_"
}

Write-Status "Fetching the latest release information from GitHub..."

# Get the latest release info
try {
    $ReleaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest" -Method Get
} catch {
    if ($_.Exception.Response.StatusCode -eq 403) {
        Write-Error1 "GitHub API rate limit exceeded. Please try again later."
    } else {
        Write-Error1 "Failed to fetch release information from GitHub: $_"
    }
}

$TagName = $ReleaseInfo.tag_name
if (-not $TagName) {
    Write-Error1 "Failed to get latest release tag. Please check repository name and owner."
}

Write-Status "Found latest release: $TagName"

# Set file pattern to look for
$FilePattern = "gcode-$TagName-windows-$Arch.tar.gz"
Write-Status "Looking for file matching pattern: $FilePattern"

# Find matching asset
$Asset = $ReleaseInfo.assets | Where-Object { $_.name -eq $FilePattern }
if (-not $Asset) {
    Write-Error1 "No matching release file found for your platform (windows) and architecture ($Arch)."
}

$AssetUrl = $Asset.browser_download_url
Write-Status "Found matching release file: $($Asset.name)"
Write-Status "Downloading from: $AssetUrl"

# Download the asset
$TarballPath = Join-Path -Path $TempDir -ChildPath "release.tar.gz"
try {
    Invoke-WebRequest -Uri $AssetUrl -OutFile $TarballPath
} catch {
    Write-Error1 "Failed to download the release file: $_"
}

Write-Status "Download complete. Extracting to $InstallDir..."

# Create installation directory if it doesn't exist
if (-not (Test-Path -Path $InstallDir)) {
    try {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    } catch {
        Write-Error1 "Failed to create installation directory: $_"
    }
}

# Extract the tarball
try {
    # Try using built-in tar command (Windows 10 1803+)
    if (Get-Command -Name tar -ErrorAction SilentlyContinue) {
        tar -xzf $TarballPath -C $InstallDir
        if ($LASTEXITCODE -ne 0) { throw "tar command failed with exit code $LASTEXITCODE" }
    } else {
        # Fallback to PowerShell extraction
        Write-Status "Built-in tar command not found. Using PowerShell for extraction..."
        
        # Extract tar.gz using .NET compression
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        
        # First, extract the gzip to get tar file
        $TarPath = Join-Path -Path $TempDir -ChildPath "release.tar"
        $input = New-Object System.IO.FileStream $TarballPath, ([IO.FileMode]::Open), ([IO.FileAccess]::Read), ([IO.FileShare]::Read)
        $output = New-Object System.IO.FileStream $TarPath, ([IO.FileMode]::Create), ([IO.FileAccess]::Write), ([IO.FileShare]::None)
        $gzipStream = New-Object System.IO.Compression.GZipStream $input, ([IO.Compression.CompressionMode]::Decompress)
        
        # Copy the decompressed bytes to the output file
        $buffer = New-Object byte[](1024)
        while ($true) {
            $read = $gzipStream.Read($buffer, 0, 1024)
            if ($read -le 0) { break }
            $output.Write($buffer, 0, $read)
        }
        
        $gzipStream.Close()
        $output.Close()
        $input.Close()
        
        # Now extract the tar file - requires 7Zip or third-party tools
        # Since Windows doesn't have native tar extraction in PowerShell, recommend manual extraction
        Write-Warning1 "PowerShell cannot extract tar files natively."
        Write-Warning1 "Please extract the downloaded file manually using 7-Zip or similar tools:"
        Write-Warning1 "File location: $TarballPath"
        Write-Warning1 "Extract to: $InstallDir"
        
        # Offer to open the temp directory
        $OpenTemp = Read-Host "Would you like to open the temp directory containing the downloaded file? (Y/N)"
        if ($OpenTemp -eq "Y" -or $OpenTemp -eq "y") {
            Start-Process $TempDir
        }
        
        # Don't clean up the temp directory since user needs to manually extract
        exit 0
    }
} catch {
    Write-Error1 "Failed to extract the archive: $_"
}

# Clean up
try {
    Remove-Item -Path $TempDir -Recurse -Force
} catch {
    Write-Warning1 "Failed to clean up temporary directory: $_"
}

Write-Status "Installation completed successfully!"

# Check if the bin directory exists
$BinPath = Join-Path -Path $InstallDir -ChildPath "/gcode/bin"
if (-not (Test-Path -Path $BinPath)) {
    Write-Warning1 "The bin directory does not exist at $BinPath. Please check the extracted contents."
} else {
    Write-Status "Adding installation directory to PATH for this session..."
    $env:PATH = "$env:PATH;$BinPath"
    
    Write-Status "To permanently add GCode to your PATH, you have two options:"
    Write-Host ""
    Write-Host "  1. Run the following command in an Administrator PowerShell:"
    Write-Host "     [Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ';$BinPath', 'User')" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  2. Manually add through Windows Settings:"
    Write-Host "     - Press Win+I to open Settings"
    Write-Host "     - Search for 'Edit environment variables'"
    Write-Host "     - Edit the 'Path' variable under 'User variables'"
    Write-Host "     - Add '$BinPath'"
    Write-Host ""
    
    # Offer to update PATH automatically
    $UpdatePath = Read-Host "Would you like to update your PATH automatically now? (Y/N)"
    if ($UpdatePath -eq "Y" -or $UpdatePath -eq "y") {
        try {
            [Environment]::SetEnvironmentVariable(
                "Path", 
                [Environment]::GetEnvironmentVariable("Path", "User") + ";$BinPath", 
                "User"
            )
            Write-Status "PATH updated successfully!"
        } catch {
            Write-Error1 "Failed to update PATH: $_"
        }
    }
    
    Write-Host ""
    Write-Status "After updating your PATH, you can run the program by typing 'gcode' in your terminal."
}

Write-Status "Installation process complete!"
Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")