# NPU-IA - Download de Modelos ONNX
# Windows PowerShell

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "    NPU-IA - Download de Modelos" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$modelsPath = ".\models"
New-Item -ItemType Directory -Path $modelsPath -Force | Out-Null

# Função para baixar com progresso
function Download-Model {
    param (
        [string]$Name,
        [string]$Url,
        [string]$Output
    )

    Write-Host "Baixando $Name..." -ForegroundColor Yellow

    try {
        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $Url -OutFile $Output
        Write-Host "    ✓ $Name baixado" -ForegroundColor Green
    }
    catch {
        Write-Host "    ✗ Erro ao baixar $Name" -ForegroundColor Red
        Write-Host "    URL: $Url" -ForegroundColor Gray
    }
}

Write-Host "Os modelos serão baixados do Hugging Face." -ForegroundColor Cyan
Write-Host "Isso pode demorar dependendo da sua conexão." -ForegroundColor Cyan
Write-Host ""

# Menu de seleção
Write-Host "Selecione os modelos para baixar:" -ForegroundColor White
Write-Host "[1] Mínimo (Whisper + Phi-3.5) - ~4GB" -ForegroundColor White
Write-Host "[2] Recomendado (Mínimo + Qwen + Llama) - ~8GB" -ForegroundColor White
Write-Host "[3] Completo (Todos os modelos) - ~13GB" -ForegroundColor White
Write-Host ""
$choice = Read-Host "Escolha (1/2/3)"

# Whisper Medium - Sempre baixa
Write-Host ""
Write-Host "[1/6] Whisper Medium (STT)" -ForegroundColor Cyan
if (-not (Test-Path "$modelsPath\whisper-medium.onnx")) {
    Write-Host "NOTA: Whisper precisa ser convertido para ONNX." -ForegroundColor Yellow
    Write-Host "Baixe de: https://huggingface.co/onnx-community/whisper-medium" -ForegroundColor Yellow
    Write-Host "Ou use: python -m optimum.exporters.onnx --model openai/whisper-medium models/whisper" -ForegroundColor Gray
} else {
    Write-Host "    ✓ Whisper já existe" -ForegroundColor Green
}

# Phi-3.5 Mini - Sempre baixa
Write-Host ""
Write-Host "[2/6] Phi-3.5 Mini (LLM rápido)" -ForegroundColor Cyan
if (-not (Test-Path "$modelsPath\phi-3.5-mini.onnx")) {
    Write-Host "NOTA: Phi-3.5 precisa ser convertido para ONNX." -ForegroundColor Yellow
    Write-Host "Baixe de: https://huggingface.co/microsoft/Phi-3.5-mini-instruct-onnx" -ForegroundColor Yellow
} else {
    Write-Host "    ✓ Phi-3.5 já existe" -ForegroundColor Green
}

if ($choice -ge 2) {
    # Qwen 2.5
    Write-Host ""
    Write-Host "[3/6] Qwen 2.5 3B (Ações)" -ForegroundColor Cyan
    if (-not (Test-Path "$modelsPath\qwen-2.5-3b.onnx")) {
        Write-Host "NOTA: Qwen precisa ser convertido para ONNX." -ForegroundColor Yellow
        Write-Host "Baixe de: https://huggingface.co/Qwen/Qwen2.5-3B-Instruct" -ForegroundColor Yellow
        Write-Host "Converta com: optimum-cli export onnx --model Qwen/Qwen2.5-3B-Instruct models/qwen" -ForegroundColor Gray
    } else {
        Write-Host "    ✓ Qwen já existe" -ForegroundColor Green
    }

    # Llama 3.2
    Write-Host ""
    Write-Host "[4/6] Llama 3.2 3B (Contexto)" -ForegroundColor Cyan
    if (-not (Test-Path "$modelsPath\llama-3.2-3b.onnx")) {
        Write-Host "NOTA: Llama precisa ser convertido para ONNX." -ForegroundColor Yellow
        Write-Host "Baixe de: https://huggingface.co/meta-llama/Llama-3.2-3B-Instruct" -ForegroundColor Yellow
    } else {
        Write-Host "    ✓ Llama já existe" -ForegroundColor Green
    }
}

if ($choice -ge 3) {
    # MiniCPM-V
    Write-Host ""
    Write-Host "[5/6] MiniCPM-V (Visão)" -ForegroundColor Cyan
    if (-not (Test-Path "$modelsPath\minicpm-v.onnx")) {
        Write-Host "NOTA: MiniCPM-V precisa ser convertido para ONNX." -ForegroundColor Yellow
        Write-Host "Baixe de: https://huggingface.co/openbmb/MiniCPM-V-2_6" -ForegroundColor Yellow
    } else {
        Write-Host "    ✓ MiniCPM-V já existe" -ForegroundColor Green
    }

    # Qwen-Coder
    Write-Host ""
    Write-Host "[6/6] Qwen-Coder 3B (Código)" -ForegroundColor Cyan
    if (-not (Test-Path "$modelsPath\qwen-coder-3b.onnx")) {
        Write-Host "NOTA: Qwen-Coder precisa ser convertido para ONNX." -ForegroundColor Yellow
        Write-Host "Baixe de: https://huggingface.co/Qwen/Qwen2.5-Coder-3B-Instruct" -ForegroundColor Yellow
    } else {
        Write-Host "    ✓ Qwen-Coder já existe" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "    Instruções de Conversão ONNX" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Para converter modelos para ONNX otimizado para NPU:" -ForegroundColor White
Write-Host ""
Write-Host "1. Instale o Optimum:" -ForegroundColor Yellow
Write-Host "   pip install optimum[onnxruntime]" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Converta o modelo:" -ForegroundColor Yellow
Write-Host "   optimum-cli export onnx --model NOME_MODELO ./models/NOME" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Para NPU AMD, quantize em INT4:" -ForegroundColor Yellow
Write-Host "   python -m onnxruntime.quantization.preprocess --input model.onnx --output model_prep.onnx" -ForegroundColor Gray
Write-Host "   python -m onnxruntime.quantization.quantize --input model_prep.onnx --output model_int4.onnx --per_channel" -ForegroundColor Gray
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "    Próximo passo: go build" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
