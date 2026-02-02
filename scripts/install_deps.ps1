# NPU-IA - Script de Instalação de Dependências
# Windows PowerShell

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "    NPU-IA - Instalação de Dependências" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Verifica se está rodando como admin
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "AVISO: Algumas instalações podem precisar de permissão de administrador." -ForegroundColor Yellow
}

# Função para verificar comando
function Test-Command($cmd) {
    return [bool](Get-Command -Name $cmd -ErrorAction SilentlyContinue)
}

# 1. Verifica Go
Write-Host "[1/6] Verificando Go..." -ForegroundColor Green
if (Test-Command "go") {
    $goVersion = go version
    Write-Host "    ✓ Go instalado: $goVersion" -ForegroundColor Green
} else {
    Write-Host "    ✗ Go não encontrado. Instalando via winget..." -ForegroundColor Yellow
    winget install -e --id GoLang.Go
}

# 2. Verifica Git
Write-Host "[2/6] Verificando Git..." -ForegroundColor Green
if (Test-Command "git") {
    $gitVersion = git --version
    Write-Host "    ✓ Git instalado: $gitVersion" -ForegroundColor Green
} else {
    Write-Host "    ✗ Git não encontrado. Instalando via winget..." -ForegroundColor Yellow
    winget install -e --id Git.Git
}

# 3. Instala FFmpeg
Write-Host "[3/6] Verificando FFmpeg..." -ForegroundColor Green
if (Test-Command "ffmpeg") {
    Write-Host "    ✓ FFmpeg instalado" -ForegroundColor Green
} else {
    Write-Host "    ✗ FFmpeg não encontrado. Instalando via winget..." -ForegroundColor Yellow
    winget install -e --id Gyan.FFmpeg
}

# 4. Instala PortAudio
Write-Host "[4/6] Verificando PortAudio..." -ForegroundColor Green
$portaudioPath = "C:\portaudio"
if (Test-Path $portaudioPath) {
    Write-Host "    ✓ PortAudio encontrado em $portaudioPath" -ForegroundColor Green
} else {
    Write-Host "    Baixando PortAudio..." -ForegroundColor Yellow
    # Baixa binários pré-compilados
    $url = "https://github.com/nicenboim/portaudio-built/releases/download/v19.7.0/portaudio_x64.zip"
    $zipPath = "$env:TEMP\portaudio.zip"
    Invoke-WebRequest -Uri $url -OutFile $zipPath
    Expand-Archive -Path $zipPath -DestinationPath $portaudioPath -Force
    Write-Host "    ✓ PortAudio instalado em $portaudioPath" -ForegroundColor Green
}

# 5. Baixa Piper TTS
Write-Host "[5/6] Verificando Piper TTS..." -ForegroundColor Green
$piperPath = ".\piper"
if (Test-Path "$piperPath\piper.exe") {
    Write-Host "    ✓ Piper TTS instalado" -ForegroundColor Green
} else {
    Write-Host "    Baixando Piper TTS..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $piperPath -Force | Out-Null

    # Baixa Piper
    $piperUrl = "https://github.com/rhasspy/piper/releases/download/2023.11.14-2/piper_windows_amd64.zip"
    $piperZip = "$env:TEMP\piper.zip"
    Invoke-WebRequest -Uri $piperUrl -OutFile $piperZip
    Expand-Archive -Path $piperZip -DestinationPath $piperPath -Force

    # Baixa voz em português
    Write-Host "    Baixando voz em português brasileiro..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path "$piperPath\voices" -Force | Out-Null
    $voiceUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/pt/pt_BR/faber/medium/pt_BR-faber-medium.onnx"
    $voiceJsonUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/pt/pt_BR/faber/medium/pt_BR-faber-medium.onnx.json"
    Invoke-WebRequest -Uri $voiceUrl -OutFile "$piperPath\voices\pt_BR-faber-medium.onnx"
    Invoke-WebRequest -Uri $voiceJsonUrl -OutFile "$piperPath\voices\pt_BR-faber-medium.onnx.json"

    Write-Host "    ✓ Piper TTS instalado com voz pt_BR" -ForegroundColor Green
}

# 6. Verifica AMD Ryzen AI Software
Write-Host "[6/6] Verificando AMD Ryzen AI Software..." -ForegroundColor Green
$amdDriverPath = "C:\Program Files\AMD\Ryzen AI Software"
if (Test-Path $amdDriverPath) {
    Write-Host "    ✓ AMD Ryzen AI Software instalado" -ForegroundColor Green
} else {
    Write-Host "    ⚠ AMD Ryzen AI Software não encontrado." -ForegroundColor Yellow
    Write-Host "    Baixe em: https://www.amd.com/en/products/software/ryzen-ai-software.html" -ForegroundColor Yellow
}

# Instala dependências Go
Write-Host ""
Write-Host "Instalando dependências Go..." -ForegroundColor Cyan
go mod download
go mod tidy

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "    Instalação concluída!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Próximos passos:" -ForegroundColor Cyan
Write-Host "1. Execute: .\scripts\download_models.ps1" -ForegroundColor White
Write-Host "2. Configure: configs\config.yaml" -ForegroundColor White
Write-Host "3. Compile: go build -o npu-ia.exe ./cmd/npu-ia" -ForegroundColor White
Write-Host "4. Execute: .\npu-ia.exe" -ForegroundColor White
