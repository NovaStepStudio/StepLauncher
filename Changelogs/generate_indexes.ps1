$ErrorActionPreference = 'Stop'
$changelogs = $PSScriptRoot
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)

# Ordena los MD de una carpeta por su numero N (StepLauncher-Error-1.md, ...).
function Sort-MdByNumber {
    param([object[]]$Files)
    return @($Files | Sort-Object @{ Expression = {
        if ($_.Name -match '(\d+)\.md$') { [int]$Matches[1] } else { 0 }
    } })
}

# Carpetas de version "StepLauncher-X.Y.Z" ordenadas de la mas nueva a la mas antigua.
function Get-VersionDirs {
    param([string]$Folder)
    return @(Get-ChildItem $Folder -Directory | ForEach-Object {
        if ($_.Name -match '^StepLauncher-(\d+\.\d+\.\d+)$') {
            [PSCustomObject]@{ Version = [version]$Matches[1]; Name = $_.Name }
        }
    } | Sort-Object Version -Descending)
}

# Serializador JSON propio: compacta con ConvertTo-Json y re-indenta a 4
# espacios por nivel (sin el padding proporcional a la profundidad que mete
# ConvertTo-Json de PS 5.1 en formato expandido).
function ConvertTo-PrettyJson {
    param([Parameter(Mandatory = $true)]$Value)
    $compressed = $Value | ConvertTo-Json -Compress -Depth 20
    $sb = New-Object System.Text.StringBuilder
    $indent = 0
    $inString = $false
    $needBreak = $true
    $nl = [Environment]::NewLine
    for ($i = 0; $i -lt $compressed.Length; $i++) {
        $c = $compressed[$i]
        if ($inString) {
            if ($needBreak) {
                [void]$sb.Append($nl)
                [void]$sb.Append(' ' * (4 * $indent))
                $needBreak = $false
            }
            [void]$sb.Append($c)
            if ($c -eq '"' -and $compressed[$i - 1] -ne '\') { $inString = $false }
            continue
        }
        if ($c -eq '{' -or $c -eq '[') {
            if ($needBreak) {
                [void]$sb.Append($nl)
                [void]$sb.Append(' ' * (4 * $indent))
            }
            [void]$sb.Append($c)
            $indent++
            $needBreak = $true
        }
        elseif ($c -eq '}' -or $c -eq ']') {
            $indent = [Math]::Max(0, $indent - 1)
            [void]$sb.Append($nl)
            [void]$sb.Append(' ' * (4 * $indent))
            [void]$sb.Append($c)
            $needBreak = $false
        }
        elseif ($c -eq ',') {
            [void]$sb.Append($c)
            $needBreak = $true
        }
        elseif ($c -eq ':') {
            [void]$sb.Append($c)
            [void]$sb.Append(' ')
            $needBreak = $false
        }
        else {
            if ($needBreak) {
                [void]$sb.Append($nl)
                [void]$sb.Append(' ' * (4 * $indent))
                $needBreak = $false
            }
            [void]$sb.Append($c)
        }
    }
    return $sb.ToString().TrimStart()
}

function Save-Json {
    param([string]$Path, [object]$Object)
    [System.IO.File]::WriteAllText($Path, (ConvertTo-PrettyJson $Object), $utf8NoBom)
}

# --- Releases/index.json ---
$releasesContent = @()
foreach ($vd in (Get-VersionDirs (Join-Path $changelogs 'Releases'))) {
    $releasesContent += [ordered]@{
        version = $vd.Version.ToString()
        path    = './' + $vd.Name + '/news.json'
    }
}
$releasesIndex = [ordered]@{}
if ($releasesContent.Count -gt 0) { $releasesIndex['latest'] = $releasesContent[0].Version }
$releasesIndex['content'] = $releasesContent
Save-Json (Join-Path $changelogs 'Releases\index.json') $releasesIndex

# --- Errors/index.json ---
$errVersions = @()
foreach ($vd in (Get-VersionDirs (Join-Path $changelogs 'Errors'))) {
    $files = @(Sort-MdByNumber (Get-ChildItem (Join-Path $changelogs "Errors\$($vd.Name)")))
    $paths = @($files | ForEach-Object { './' + $vd.Name + '/' + $_.Name })
    $errVersions += [ordered]@{ version = $vd.Version.ToString(); errors = $paths }
}
Save-Json (Join-Path $changelogs 'Errors\index.json') @{ versions = $errVersions }

# --- Changes/index.json ---
$chgVersions = @()
foreach ($vd in (Get-VersionDirs (Join-Path $changelogs 'Changes'))) {
    $files = @(Sort-MdByNumber (Get-ChildItem (Join-Path $changelogs "Changes\$($vd.Name)")))
    $paths = @($files | ForEach-Object { './' + $vd.Name + '/' + $_.Name })
    $chgVersions += [ordered]@{ version = $vd.Version.ToString(); changes = $paths }
}
Save-Json (Join-Path $changelogs 'Changes\index.json') @{ versions = $chgVersions }

Write-Output 'INDEX OK'