package npu

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

// DirectML wrapper para AMD NPU
type DirectML struct {
	deviceID   int
	deviceName string
	available  bool
}

// DeviceInfo informações do dispositivo NPU
type DeviceInfo struct {
	Name         string
	Vendor       string
	DriverVersion string
	MemoryMB     int64
	TOPs         float64
}

// NewDirectML inicializa DirectML
func NewDirectML() (*DirectML, error) {
	dm := &DirectML{
		deviceID: 0,
	}

	// Verifica disponibilidade do DirectML
	if err := dm.checkAvailability(); err != nil {
		return nil, err
	}

	return dm, nil
}

// checkAvailability verifica se DirectML está disponível
func (dm *DirectML) checkAvailability() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("DirectML só disponível no Windows")
	}

	// Tenta carregar DirectML.dll
	directml, err := syscall.LoadDLL("DirectML.dll")
	if err != nil {
		return fmt.Errorf("DirectML não encontrado: %w", err)
	}
	defer directml.Release()

	dm.available = true
	return nil
}

// GetDeviceInfo retorna informações do dispositivo NPU
func (dm *DirectML) GetDeviceInfo() (*DeviceInfo, error) {
	if !dm.available {
		return nil, fmt.Errorf("DirectML não disponível")
	}

	// TODO: Implementar query real do dispositivo via DXGI
	// Por enquanto retorna info hardcoded para AMD Ryzen AI

	return &DeviceInfo{
		Name:          "AMD Ryzen AI",
		Vendor:        "AMD",
		DriverVersion: "1.0.0",
		MemoryMB:      0, // NPU compartilha memória do sistema
		TOPs:          55.0,
	}, nil
}

// IsAvailable verifica se NPU está disponível
func (dm *DirectML) IsAvailable() bool {
	return dm.available
}

// GetDeviceID retorna ID do dispositivo
func (dm *DirectML) GetDeviceID() int {
	return dm.deviceID
}

// SetDevice seleciona dispositivo NPU
func (dm *DirectML) SetDevice(deviceID int) error {
	// TODO: Implementar seleção de dispositivo
	dm.deviceID = deviceID
	return nil
}

// Placeholder para warning unused
var _ = unsafe.Pointer(nil)
