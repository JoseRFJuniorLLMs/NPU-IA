package vision

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"syscall"
	"unsafe"

	ort "github.com/yalue/onnxruntime_go"
	"github.com/JoseRFJuniorLLMs/NPU-IA/pkg/config"
)

// Model representa o modelo de visão MiniCPM-V
type Model struct {
	session *ort.DynamicAdvancedSession
	config  config.ModelConfig
}

// New cria um novo modelo de visão
func New(cfg config.ModelConfig) (*Model, error) {
	options, err := ort.NewSessionOptions()
	if err != nil {
		return nil, err
	}

	// Usa DirectML para NPU AMD
	err = options.AppendExecutionProviderDirectML(0)
	if err != nil {
		fmt.Println("DirectML não disponível para Vision, usando CPU")
	}

	// Carrega modelo MiniCPM-V ONNX
	session, err := ort.NewDynamicAdvancedSession(
		cfg.Path,
		[]string{"pixel_values", "input_ids"},
		[]string{"logits"},
		options,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar modelo de visão: %w", err)
	}

	return &Model{
		session: session,
		config:  cfg,
	}, nil
}

// CaptureScreen captura a tela atual
func (m *Model) CaptureScreen() ([]byte, error) {
	// Windows API para screenshot
	user32 := syscall.NewLazyDLL("user32.dll")
	gdi32 := syscall.NewLazyDLL("gdi32.dll")

	getDC := user32.NewProc("GetDC")
	releaseDC := user32.NewProc("ReleaseDC")
	getSystemMetrics := user32.NewProc("GetSystemMetrics")
	createCompatibleDC := gdi32.NewProc("CreateCompatibleDC")
	createCompatibleBitmap := gdi32.NewProc("CreateCompatibleBitmap")
	selectObject := gdi32.NewProc("SelectObject")
	bitBlt := gdi32.NewProc("BitBlt")
	deleteDC := gdi32.NewProc("DeleteDC")
	deleteObject := gdi32.NewProc("DeleteObject")

	// Obtém dimensões da tela
	width, _, _ := getSystemMetrics.Call(0)  // SM_CXSCREEN
	height, _, _ := getSystemMetrics.Call(1) // SM_CYSCREEN

	// Obtém DC da tela
	hdcScreen, _, _ := getDC.Call(0)
	defer releaseDC.Call(0, hdcScreen)

	// Cria DC compatível
	hdcMem, _, _ := createCompatibleDC.Call(hdcScreen)
	defer deleteDC.Call(hdcMem)

	// Cria bitmap
	hBitmap, _, _ := createCompatibleBitmap.Call(hdcScreen, width, height)
	defer deleteObject.Call(hBitmap)

	// Seleciona bitmap no DC
	selectObject.Call(hdcMem, hBitmap)

	// Copia tela para bitmap
	bitBlt.Call(
		hdcMem, 0, 0, width, height,
		hdcScreen, 0, 0,
		0x00CC0020, // SRCCOPY
	)

	// Converte para imagem Go
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

	// TODO: Copiar pixels do bitmap para img

	// Salva temporariamente e lê bytes
	tmpFile := os.TempDir() + "/screenshot.png"
	f, err := os.Create(tmpFile)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)
	defer f.Close()

	png.Encode(f, img)

	return os.ReadFile(tmpFile)
}

// Analyze analisa uma imagem com o prompt
func (m *Model) Analyze(ctx context.Context, imageData []byte, prompt string) (string, error) {
	// Pré-processa imagem
	pixelValues, err := m.preprocessImage(imageData)
	if err != nil {
		return "", err
	}

	// Tokeniza prompt
	inputIDs := m.tokenizePrompt(prompt)

	// Prepara tensores
	imgShape := ort.NewShape(1, 3, 384, 384) // MiniCPM-V input size
	imgTensor, err := ort.NewTensor(imgShape, pixelValues)
	if err != nil {
		return "", err
	}
	defer imgTensor.Destroy()

	textShape := ort.NewShape(1, int64(len(inputIDs)))
	textTensor, err := ort.NewTensor(textShape, inputIDs)
	if err != nil {
		return "", err
	}
	defer textTensor.Destroy()

	// Executa inferência
	// TODO: Implementar geração autoregressiva

	return "Análise da imagem", nil
}

// preprocessImage converte imagem para tensor
func (m *Model) preprocessImage(data []byte) ([]float32, error) {
	// TODO: Implementar pré-processamento real
	// - Decodificar PNG/JPEG
	// - Redimensionar para 384x384
	// - Normalizar pixels
	// - Converter para CHW format

	size := 3 * 384 * 384
	pixels := make([]float32, size)
	return pixels, nil
}

// tokenizePrompt tokeniza o prompt
func (m *Model) tokenizePrompt(prompt string) []int64 {
	// TODO: Implementar tokenização real
	tokens := make([]int64, len(prompt))
	return tokens
}

// Close libera recursos
func (m *Model) Close() error {
	if m.session != nil {
		return m.session.Destroy()
	}
	return nil
}

// Windows constants
const (
	SRCCOPY = 0x00CC0020
)

// Placeholder para remover warning de unused
var _ = unsafe.Pointer(nil)
