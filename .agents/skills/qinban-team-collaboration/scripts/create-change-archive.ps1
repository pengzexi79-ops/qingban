[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidateSet('director', 'frontend-core', 'frontend-chat', 'frontend-experience', 'backend-lead')]
    [string]$Role,

    [Parameter(Mandatory = $true)]
    [ValidateNotNullOrEmpty()]
    [string]$Member,

    [Parameter(Mandatory = $true)]
    [ValidateNotNullOrEmpty()]
    [string]$TaskId,

    [Parameter(Mandatory = $true)]
    [ValidateNotNullOrEmpty()]
    [string]$Summary,

    [string]$Details = '未补充具体变化。',
    [string]$Impact = '无已知 API、数据库或兼容性影响。',
    [string]$Tests = '未执行测试。',
    [string]$Handoff = '无。',
    [switch]$TechnicalApproval,
    [switch]$DryRun
)

$ErrorActionPreference = 'Stop'

function Invoke-Git {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)
    $output = & git @Arguments 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "git $($Arguments -join ' ') failed:`n$($output -join "`n")"
    }
    return @($output)
}

function Convert-ToSafeName([string]$Value) {
    $safe = $Value.Trim() -replace '[^A-Za-z0-9._-]+', '-'
    $safe = $safe.Trim('-')
    if ([string]::IsNullOrWhiteSpace($safe)) { return 'unknown' }
    return $safe
}

function Convert-ToYamlQuoted([string]$Value) {
    $escaped = $Value.Replace('\', '\\').Replace('"', '\"').Replace("`r", '').Replace("`n", '\n')
    return '"{0}"' -f $escaped
}

function Get-Sha256Text([string]$Value) {
    $sha = [System.Security.Cryptography.SHA256]::Create()
    try {
        $bytes = [System.Text.UTF8Encoding]::new($false).GetBytes($Value)
        return ([System.BitConverter]::ToString($sha.ComputeHash($bytes))).Replace('-', '').ToLowerInvariant()
    }
    finally {
        $sha.Dispose()
    }
}

$repoRoot = (Invoke-Git rev-parse --show-toplevel | Select-Object -First 1).Trim()
Set-Location -LiteralPath $repoRoot

$branch = (Invoke-Git branch --show-current | Select-Object -First 1).Trim()
if ([string]::IsNullOrWhiteSpace($branch)) {
    throw 'Detached HEAD is not supported for team archives.'
}
if ($branch -eq 'master') {
    throw 'Refusing to archive work developed directly on master. Create a task branch first.'
}

$statusLines = @(Invoke-Git -c core.quotepath=false status --porcelain=v1 --untracked-files=all)
$records = foreach ($line in $statusLines) {
    if ([string]::IsNullOrWhiteSpace($line) -or $line.Length -lt 4) { continue }

    $path = $line.Substring(3)
    if ($path -match '^docs/development-archive/\d{4}-\d{2}-\d{2}/') { continue }

    [pscustomobject]@{
        Status = $line.Substring(0, 2)
        Path = $path
    }
}
$records = @($records | Sort-Object Path)

if (-not $records) {
    throw 'No unarchived working-tree changes were found.'
}

$backendPatterns = @(
    '^backend/', '^build/', '^bruno/', '^tests/',
    '^docs/openapi(?:\.|$)', '^docs/PHASE1_API\.md$'
)
$backendOwned = @($records | Where-Object {
    $candidate = $_.Path
    $backendPatterns | Where-Object { $candidate -match $_ } | Select-Object -First 1
})
$frontendOwned = @($records | Where-Object { $_.Path -match '^frontend/' })

if ($Role -like 'frontend-*' -and $backendOwned -and -not $TechnicalApproval) {
    throw "Frontend role changed backend-owned files: $($backendOwned.Path -join ', '). Obtain backend-lead approval or split the task."
}
if ($Role -eq 'backend-lead' -and $frontendOwned -and -not $TechnicalApproval) {
    throw "Backend role changed frontend files: $($frontendOwned.Path -join ', '). Record explicit cross-boundary approval or split the task."
}
if ($Role -eq 'director' -and ($backendOwned -or $frontendOwned) -and -not $TechnicalApproval) {
    throw 'Director role changed implementation files. Declare a frontend execution role or record backend-lead approval.'
}

$baseCommit = (Invoke-Git rev-parse --short HEAD | Select-Object -First 1).Trim()

$fingerprintRows = foreach ($record in $records) {
    $workingPath = $record.Path
    if ($workingPath -match ' -> ') {
        $workingPath = ($workingPath -split ' -> ', 2)[1]
    }

    $absoluteWorkingPath = Join-Path $repoRoot ($workingPath -replace '/', [System.IO.Path]::DirectorySeparatorChar)
    if (Test-Path -LiteralPath $absoluteWorkingPath -PathType Leaf) {
        $contentHash = (Get-FileHash -LiteralPath $absoluteWorkingPath -Algorithm SHA256).Hash.ToLowerInvariant()
    }
    elseif (Test-Path -LiteralPath $absoluteWorkingPath -PathType Container) {
        $contentHash = 'directory-or-submodule'
    }
    else {
        $contentHash = 'deleted'
    }

    '{0}|{1}' -f $record.Path, $contentHash
}
$fingerprintSeed = @(
    "branch=$branch"
    "base_commit=$baseCommit"
    "task_id=$TaskId"
) + @($fingerprintRows | Sort-Object)
$changeFingerprint = Get-Sha256Text ($fingerprintSeed -join "`n")

$archiveRoot = Join-Path $repoRoot 'docs\development-archive'
$matchingArchive = $null
if (Test-Path -LiteralPath $archiveRoot) {
    $matchingArchive = Get-ChildItem -LiteralPath $archiveRoot -Recurse -File -Filter '*.md' |
        Where-Object { $_.Directory.Name -match '^\d{4}-\d{2}-\d{2}$' } |
        Select-String -Pattern ('^change_fingerprint:\s*"?{0}"?\s*$' -f [regex]::Escape($changeFingerprint)) |
        Select-Object -First 1
}
if ($matchingArchive) {
    $existingRelative = [System.IO.Path]::GetRelativePath($repoRoot, $matchingArchive.Path).Replace('\', '/')
    throw "Current change state is already archived at $existingRelative"
}

$unstagedCheckOutput = & git diff --check 2>&1
$unstagedCheckPassed = $LASTEXITCODE -eq 0
$stagedCheckOutput = & git diff --cached --check 2>&1
$stagedCheckPassed = $LASTEXITCODE -eq 0
$diffCheckPassed = $unstagedCheckPassed -and $stagedCheckPassed
if ($diffCheckPassed) {
    $diffCheckText = '通过（已检查暂存区与未暂存区；未跟踪文件不在 git diff --check 范围内）。'
}
else {
    $checkFailures = @()
    if (-not $unstagedCheckPassed) {
        $checkFailures += "未暂存区：`n$($unstagedCheckOutput -join "`n")"
    }
    if (-not $stagedCheckPassed) {
        $checkFailures += "暂存区：`n$($stagedCheckOutput -join "`n")"
    }
    $diffCheckText = "未通过：`n$($checkFailures -join "`n`n")"
}

$now = Get-Date
$dateDirectory = $now.ToString('yyyy-MM-dd')
$stamp = $now.ToString('yyyyMMdd-HHmmss-fff')
$fileName = '{0}-{1}-{2}-{3}.md' -f $stamp, (Convert-ToSafeName $Role), (Convert-ToSafeName $Member), (Convert-ToSafeName $TaskId)
$relativePath = "docs/development-archive/$dateDirectory/$fileName"
$absolutePath = Join-Path $repoRoot ($relativePath -replace '/', [System.IO.Path]::DirectorySeparatorChar)

$fileLines = @($records | ForEach-Object {
    '- `{0}` `{1}`' -f $_.Status, $_.Path
})
$approvalText = if ($TechnicalApproval) { '是，已声明技术负责人批准跨界。' } else { '否。' }
$createdAt = $now.ToString('yyyy-MM-ddTHH:mm:sszzz')

$contentLines = @(
    '---'
    'archive_version: 1'
    ('created_at: {0}' -f (Convert-ToYamlQuoted $createdAt))
    ('member: {0}' -f (Convert-ToYamlQuoted $Member))
    ('role: {0}' -f (Convert-ToYamlQuoted $Role))
    ('task_id: {0}' -f (Convert-ToYamlQuoted $TaskId))
    ('branch: {0}' -f (Convert-ToYamlQuoted $branch))
    ('base_commit: {0}' -f (Convert-ToYamlQuoted $baseCommit))
    ('change_fingerprint: {0}' -f (Convert-ToYamlQuoted $changeFingerprint))
    '---'
    ''
    ('# 亲伴开发保存点：{0}' -f $TaskId)
    ''
    '## 本次目的'
    ''
    $Summary
    ''
    '## Git 检测到的变更'
    ''
)
$contentLines += $fileLines
$contentLines += @(
    ''
    '## 具体变化'
    ''
    $Details
    ''
    '## 接口、数据库与兼容影响'
    ''
    $Impact
    ''
    ('跨角色技术批准：{0}' -f $approvalText)
    ''
    '## 验证'
    ''
    $Tests
    ''
    ('git diff --check：{0}' -f $diffCheckText)
    ''
    '## 风险、未完成与交接'
    ''
    $Handoff
)
$content = $contentLines -join "`n"

if ($DryRun) {
    Write-Output "[DRY RUN] $relativePath"
    Write-Output $content
    exit 0
}

$directory = Split-Path -Parent $absolutePath
[System.IO.Directory]::CreateDirectory($directory) | Out-Null
[System.IO.File]::WriteAllText($absolutePath, $content.TrimEnd("`r", "`n") + "`n", [System.Text.UTF8Encoding]::new($false))

Write-Output $relativePath
