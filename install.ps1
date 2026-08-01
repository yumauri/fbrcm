& {
    # One-line installer for fbrcm.
    # Run: irm https://raw.githubusercontent.com/yumauri/fbrcm/main/install.ps1 | iex

    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'

    $repo = 'yumauri/fbrcm'
    $app = 'fbrcm'

    if ([System.Environment]::OSVersion.Platform -ne [System.PlatformID]::Win32NT) {
        throw 'This installer only supports Windows.'
    }

    if ($PSVersionTable.PSVersion.Major -lt 5) {
        throw 'PowerShell 5.0 or newer is required.'
    }

    if ([string]::IsNullOrWhiteSpace($env:INSTALL_DIR)) {
        if ([string]::IsNullOrWhiteSpace($env:LOCALAPPDATA)) {
            throw 'LOCALAPPDATA is not set. Set INSTALL_DIR to a writable directory and try again.'
        }
        $installDir = Join-Path $env:LOCALAPPDATA 'Programs\fbrcm\bin'
    } else {
        $installDir = [System.IO.Path]::GetFullPath($env:INSTALL_DIR)
    }

    $runtimeArchitecture = $null
    try {
        $runtimeArchitecture = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
    } catch {
        $runtimeArchitecture = if ($env:PROCESSOR_ARCHITEW6432) {
            $env:PROCESSOR_ARCHITEW6432
        } else {
            $env:PROCESSOR_ARCHITECTURE
        }
    }

    $architecture = switch -Regex ($runtimeArchitecture) {
        '^(AMD64|X64|x86_64)$' { 'x86_64'; break }
        '^(ARM64|Arm64|arm64)$' { 'arm64'; break }
        default { throw "Unsupported architecture: $runtimeArchitecture" }
    }

    # GitHub requires TLS 1.2 when Windows PowerShell uses an older .NET default.
    [System.Net.ServicePointManager]::SecurityProtocol =
        [System.Net.ServicePointManager]::SecurityProtocol -bor [System.Net.SecurityProtocolType]::Tls12

    $headers = @{
        Accept = 'application/vnd.github+json'
        'User-Agent' = 'fbrcm-installer'
    }
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest" -Headers $headers
    $latest = $release.tag_name
    if ([string]::IsNullOrWhiteSpace($latest)) {
        throw "Could not determine the latest release. Check https://github.com/$repo/releases"
    }

    $asset = "${app}_windows_${architecture}.zip"
    $baseUrl = "https://github.com/$repo/releases/download/$latest"
    $tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("$app-" + [System.Guid]::NewGuid())
    $archivePath = Join-Path $tempDir $asset
    $checksumsPath = Join-Path $tempDir 'checksums.txt'

    try {
        New-Item -ItemType Directory -Path $tempDir | Out-Null

        Write-Host "Downloading $app $latest for windows/$architecture..."
        Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/$asset" -OutFile $archivePath -Headers $headers
        Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/checksums.txt" -OutFile $checksumsPath -Headers $headers

        $checksumPattern = '^([A-Fa-f0-9]{64})\s+\*?' + [System.Text.RegularExpressions.Regex]::Escape($asset) + '$'
        $expectedChecksum = $null
        foreach ($line in Get-Content -LiteralPath $checksumsPath) {
            if ($line -match $checksumPattern) {
                $expectedChecksum = $Matches[1]
                break
            }
        }
        if ([string]::IsNullOrWhiteSpace($expectedChecksum)) {
            throw "Could not find a checksum for $asset."
        }

        $actualChecksum = (Get-FileHash -LiteralPath $archivePath -Algorithm SHA256).Hash
        if ($actualChecksum -ine $expectedChecksum) {
            throw "Checksum verification failed for $asset."
        }

        Expand-Archive -LiteralPath $archivePath -DestinationPath $tempDir -Force
        $source = Join-Path $tempDir "$app.exe"
        if (-not (Test-Path -LiteralPath $source -PathType Leaf)) {
            throw "The release archive does not contain $app.exe."
        }

        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        $destination = Join-Path $installDir "$app.exe"
        Write-Host "Installing to $destination..."
        Copy-Item -LiteralPath $source -Destination $destination -Force

        $normalizedInstallDir = $installDir.TrimEnd('\', '/')
        $userPath = [System.Environment]::GetEnvironmentVariable('Path', 'User')
        $pathContainsInstallDir = @($userPath -split ';') | Where-Object {
            $_.Trim().TrimEnd('\', '/') -ieq $normalizedInstallDir
        }
        if (-not $pathContainsInstallDir) {
            $newUserPath = if ([string]::IsNullOrWhiteSpace($userPath)) {
                $installDir
            } else {
                $userPath.TrimEnd(';') + ";$installDir"
            }
            [System.Environment]::SetEnvironmentVariable('Path', $newUserPath, 'User')
            Write-Host "Added $installDir to your user PATH."
        }

        $processPathContainsInstallDir = @($env:Path -split ';') | Where-Object {
            $_.Trim().TrimEnd('\', '/') -ieq $normalizedInstallDir
        }
        if (-not $processPathContainsInstallDir) {
            $env:Path = $env:Path.TrimEnd(';') + ";$installDir"
        }

        Write-Host "Done! Run: $app --help"
        Write-Host 'Open a new terminal if fbrcm is not found in an existing one.'
    } finally {
        if (Test-Path -LiteralPath $tempDir) {
            Remove-Item -LiteralPath $tempDir -Recurse -Force
        }
    }
}
