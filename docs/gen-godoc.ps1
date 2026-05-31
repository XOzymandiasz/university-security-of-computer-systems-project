# generate-godoc-html.ps1
# Generuje dokumentację GoDoc jako HTML w katalogu docs/godoc-html.
# Obsługuje polskie znaki przez UTF-8.

$ErrorActionPreference = "Stop"

# Wymuszenie UTF-8 w konsoli Windows.
chcp 65001 | Out-Null

$Utf8NoBom = [System.Text.UTF8Encoding]::new($false)

[Console]::InputEncoding = $Utf8NoBom
[Console]::OutputEncoding = $Utf8NoBom
$OutputEncoding = $Utf8NoBom

$env:LANG = "pl_PL.UTF-8"
$env:LC_ALL = "pl_PL.UTF-8"

# Ustawienie kodowania konsoli na UTF-8.
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)

$OutputDir = "godoc"

if (!(Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Force $OutputDir | Out-Null
}

$packages = @(
    @{ Path = "./cmd/client";                       Name = "cmd_client";           Title = "CMD Client" },
    @{ Path = "./cmd/server";                       Name = "cmd_server";           Title = "CMD Server" },
    @{ Path = "./cmd/third-part";                   Name = "cmd_third_part";       Title = "CMD TTP" },

    @{ Path = "./internal/shared/identity";         Name = "shared_identity";      Title = "Shared Identity" },
    @{ Path = "./internal/shared/protocol";         Name = "shared_protocol";      Title = "Shared Protocol" },

    @{ Path = "./internal/client";                  Name = "client_app";           Title = "Client App" },
    @{ Path = "./internal/client/client";           Name = "client_http_client";   Title = "Client HTTP Client" },
    @{ Path = "./internal/client/httpapi";          Name = "client_httpapi";       Title = "Client HTTP API" },
    @{ Path = "./internal/client/usecase";          Name = "client_usecase";       Title = "Client Use Cases" },

    @{ Path = "./internal/server";                  Name = "server_app";           Title = "Server App" },
    @{ Path = "./internal/server/client";           Name = "server_ttp_client";    Title = "Server TTP Client" },
    @{ Path = "./internal/server/httpapi";          Name = "server_httpapi";       Title = "Server HTTP API" },

    @{ Path = "./internal/third-part";              Name = "ttp_app";              Title = "TTP App" },
    @{ Path = "./internal/third-part/httpapi";      Name = "ttp_httpapi";          Title = "TTP HTTP API" },
    @{ Path = "./internal/third-part/ttpservice";   Name = "ttpservice";           Title = "TTP Service" }
)

function Convert-ToHtmlEscaped {
    param([string]$Text)

    return $Text.Replace("&", "&amp;").
            Replace("<", "&lt;").
            Replace(">", "&gt;")
}

function Write-Utf8NoBom {
    param(
        [string]$Path,
        [string]$Content
    )

    $utf8NoBom = [System.Text.UTF8Encoding]::new($false)
    [System.IO.File]::WriteAllText($Path, $Content, $utf8NoBom)
}

$indexLinks = ""

foreach ($pkg in $packages) {
    $path = $pkg.Path
    $name = $pkg.Name
    $title = $pkg.Title

    $outFile = Join-Path $OutputDir "$name.html"

    Write-Host "Generating GoDoc HTML for $path -> $outFile"

    $docText = (& go doc -all $path) -join "`n"
    $escaped = Convert-ToHtmlEscaped $docText

    $html = @"
<!DOCTYPE html>
<html lang="pl">
<head>
    <meta charset="UTF-8">
    <title>$title - GoDoc</title>
    <style>
        body {
            font-family: Segoe UI, Arial, sans-serif;
            margin: 0;
            background: #f6f8fa;
            color: #24292f;
        }
        header {
            background: #0d1117;
            color: white;
            padding: 20px 32px;
        }
        main {
            max-width: 1100px;
            margin: 32px auto;
            background: white;
            padding: 32px;
            border-radius: 12px;
            box-shadow: 0 4px 18px rgba(0,0,0,0.08);
        }
        a {
            color: #0969da;
            text-decoration: none;
        }
        pre {
            white-space: pre-wrap;
            font-family: Consolas, monospace;
            font-size: 14px;
            line-height: 1.45;
        }
        .back {
            display: inline-block;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
<header>
    <h1>$title</h1>
    <p>$path</p>
</header>
<main>
    <a class="back" href="index.html">← Powrót do indeksu</a>
    <pre>$escaped</pre>
</main>
</body>
</html>
"@

    Write-Utf8NoBom -Path $outFile -Content $html

    $indexLinks += "<li><a href='$name.html'>$title</a><span>$path</span></li>`n"
}

$indexHtml = @"
<!DOCTYPE html>
<html lang="pl">
<head>
    <meta charset="UTF-8">
    <title>GoDoc - Protokół TTP</title>
    <style>
        body {
            font-family: Segoe UI, Arial, sans-serif;
            margin: 0;
            background: #f6f8fa;
            color: #24292f;
        }
        header {
            background: #0d1117;
            color: white;
            padding: 28px 40px;
        }
        main {
            max-width: 1000px;
            margin: 32px auto;
            background: white;
            padding: 32px;
            border-radius: 12px;
            box-shadow: 0 4px 18px rgba(0,0,0,0.08);
        }
        h1 {
            margin: 0;
        }
        p {
            margin-top: 8px;
            color: #d0d7de;
        }
        ul {
            list-style: none;
            padding: 0;
        }
        li {
            padding: 14px 0;
            border-bottom: 1px solid #d8dee4;
        }
        a {
            color: #0969da;
            font-weight: 600;
            text-decoration: none;
            display: block;
        }
        span {
            color: #57606a;
            font-size: 13px;
        }
    </style>
</head>
<body>
<header>
    <h1>GoDoc - Protokół TTP</h1>
    <p>Dokumentacja pakietów Go dla projektu Client-Server-TTP</p>
</header>
<main>
    <h2>Pakiety</h2>
    <ul>
        $indexLinks
    </ul>
</main>
</body>
</html>
"@

Write-Utf8NoBom -Path (Join-Path $OutputDir "index.html") -Content $indexHtml

Write-Host ""
Write-Host "HTML GoDoc generated in $OutputDir"
Write-Host "Open: docs/godoc-html/index.html"