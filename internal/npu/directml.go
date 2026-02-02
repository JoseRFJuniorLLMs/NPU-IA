package npu

import (
	"fmt"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

// DirectML wrapper para AMD NPU
type DirectML struct {
	deviceID    int
	deviceName  string
	available   bool
	deviceInfo  *DeviceInfo
	dxgiFactory uintptr
}

// DeviceInfo informações do dispositivo NPU
type DeviceInfo struct {
	Name          string
	Vendor        string
	DriverVersion string
	MemoryMB      int64
	TOPs          float64
	DeviceID      uint32
	VendorID      uint32
	IsNPU         bool
	IsIntegrated  bool
}

// DXGI_ADAPTER_DESC estrutura para descrição do adapter
type DXGI_ADAPTER_DESC struct {
	Description           [128]uint16
	VendorId              uint32
	DeviceId              uint32
	SubSysId              uint32
	Revision              uint32
	DedicatedVideoMemory  uint64
	DedicatedSystemMemory uint64
	SharedSystemMemory    uint64
	AdapterLuid           [8]byte
}

var (
	dxgi                       = syscall.NewLazyDLL("dxgi.dll")
	procCreateDXGIFactory1     = dxgi.NewProc("CreateDXGIFactory1")
)

// GUIDs necessários
var (
	IID_IDXGIFactory1 = [16]byte{0x77, 0x0a, 0xae, 0x78, 0xf2, 0x6f, 0x4d, 0xba, 0xa8, 0x29, 0x25, 0x3c, 0x83, 0xd1, 0xb3, 0x87}
)

// NewDirectML inicializa DirectML
func NewDirectML() (*DirectML, error) {
	dm := &DirectML{
		deviceID: 0,
	}

	// Verifica disponibilidade do DirectML
	if err := dm.checkAvailability(); err != nil {
		return nil, err
	}

	// Enumera dispositivos para encontrar NPU/GPU
	if err := dm.enumerateDevices(); err != nil {
		// Não é erro fatal, pode usar info padrão
		fmt.Printf("Aviso: não foi possível enumerar dispositivos: %v\n", err)
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

// enumerateDevices enumera dispositivos DXGI para encontrar NPU
func (dm *DirectML) enumerateDevices() error {
	// Cria DXGI Factory
	var factory uintptr
	hr, _, _ := procCreateDXGIFactory1.Call(
		uintptr(unsafe.Pointer(&IID_IDXGIFactory1)),
		uintptr(unsafe.Pointer(&factory)),
	)

	if hr != 0 {
		return fmt.Errorf("falha ao criar DXGI Factory: 0x%x", hr)
	}

	dm.dxgiFactory = factory

	// Enumera adapters
	var bestDevice *DeviceInfo
	var npuDevice *DeviceInfo

	for i := uint32(0); i < 10; i++ {
		device, err := dm.getAdapterInfo(factory, i)
		if err != nil {
			break
		}

		// Prioriza NPU AMD
		if device.IsNPU && strings.Contains(strings.ToLower(device.Vendor), "amd") {
			npuDevice = device
		}

		// Guarda melhor dispositivo AMD
		if strings.Contains(strings.ToLower(device.Vendor), "amd") {
			if bestDevice == nil || device.MemoryMB > bestDevice.MemoryMB {
				bestDevice = device
			}
		}
	}

	// Usa NPU se encontrado, senão usa melhor dispositivo
	if npuDevice != nil {
		dm.deviceInfo = npuDevice
		dm.deviceName = npuDevice.Name
	} else if bestDevice != nil {
		dm.deviceInfo = bestDevice
		dm.deviceName = bestDevice.Name
	} else {
		// Fallback para info padrão
		dm.deviceInfo = dm.getDefaultDeviceInfo()
	}

	return nil
}

// getAdapterInfo obtém informações de um adapter específico
func (dm *DirectML) getAdapterInfo(factory uintptr, index uint32) (*DeviceInfo, error) {
	// Interface vtable para IDXGIFactory
	// EnumAdapters está no offset 7 da vtable
	factoryVtbl := *(*[20]uintptr)(unsafe.Pointer(factory))

	var adapter uintptr
	hr, _, _ := syscall.SyscallN(
		factoryVtbl[7], // EnumAdapters
		factory,
		uintptr(index),
		uintptr(unsafe.Pointer(&adapter)),
	)

	if hr != 0 {
		return nil, fmt.Errorf("adapter não encontrado")
	}

	// GetDesc está no offset 8 da vtable do adapter
	adapterVtbl := *(*[20]uintptr)(unsafe.Pointer(adapter))

	var desc DXGI_ADAPTER_DESC
	hr, _, _ = syscall.SyscallN(
		adapterVtbl[8], // GetDesc
		adapter,
		uintptr(unsafe.Pointer(&desc)),
	)

	// Release adapter
	syscall.SyscallN(adapterVtbl[2], adapter) // Release

	if hr != 0 {
		return nil, fmt.Errorf("falha ao obter descrição")
	}

	// Converte descrição para string
	name := utf16ToString(desc.Description[:])
	vendor := getVendorName(desc.VendorId)

	// Detecta se é NPU
	isNPU := detectNPU(name, desc.VendorId, desc.DeviceId)

	// Calcula TOPS estimado para NPU AMD
	tops := float64(0)
	if isNPU && desc.VendorId == 0x1002 { // AMD
		tops = estimateAMDNPUTOPs(name)
	}

	return &DeviceInfo{
		Name:          name,
		Vendor:        vendor,
		VendorID:      desc.VendorId,
		DeviceID:      desc.DeviceId,
		MemoryMB:      int64(desc.DedicatedVideoMemory / (1024 * 1024)),
		TOPs:          tops,
		IsNPU:         isNPU,
		IsIntegrated:  desc.DedicatedVideoMemory == 0,
		DriverVersion: "N/A",
	}, nil
}

// utf16ToString converte array UTF-16 para string Go
func utf16ToString(s []uint16) string {
	for i, v := range s {
		if v == 0 {
			s = s[:i]
			break
		}
	}
	return syscall.UTF16ToString(s)
}

// getVendorName retorna nome do fabricante pelo ID
func getVendorName(vendorID uint32) string {
	vendors := map[uint32]string{
		0x1002: "AMD",
		0x10DE: "NVIDIA",
		0x8086: "Intel",
		0x1414: "Microsoft",
		0x5143: "Qualcomm",
	}

	if name, ok := vendors[vendorID]; ok {
		return name
	}
	return fmt.Sprintf("Unknown (0x%04X)", vendorID)
}

// detectNPU detecta se o dispositivo é um NPU
func detectNPU(name string, vendorID, deviceID uint32) bool {
	nameLower := strings.ToLower(name)

	// Keywords que indicam NPU
	npuKeywords := []string{
		"npu", "neural", "ai engine", "ryzen ai",
		"xdna", "aipu", "neural processor",
	}

	for _, kw := range npuKeywords {
		if strings.Contains(nameLower, kw) {
			return true
		}
	}

	// Device IDs conhecidos de NPUs AMD
	// AMD Ryzen AI series NPU device IDs
	amdNPUDeviceIDs := map[uint32]bool{
		0x1502: true, // Phoenix NPU
		0x17F0: true, // Hawk Point NPU
		0x1640: true, // Strix Point NPU
	}

	if vendorID == 0x1002 && amdNPUDeviceIDs[deviceID] {
		return true
	}

	return false
}

// estimateAMDNPUTOPs estima TOPS do NPU AMD baseado no nome
func estimateAMDNPUTOPs(name string) float64 {
	nameLower := strings.ToLower(name)

	// Strix Point (Ryzen AI 300 series) = 50+ TOPS
	if strings.Contains(nameLower, "strix") ||
		strings.Contains(nameLower, "ryzen ai 3") ||
		strings.Contains(nameLower, "ai 300") {
		return 55.0
	}

	// Hawk Point (Ryzen 8000 series) = 16 TOPS
	if strings.Contains(nameLower, "hawk") ||
		strings.Contains(nameLower, "8000") {
		return 16.0
	}

	// Phoenix (Ryzen 7040 series) = 10 TOPS
	if strings.Contains(nameLower, "phoenix") ||
		strings.Contains(nameLower, "7040") {
		return 10.0
	}

	// Default para NPU desconhecido
	return 10.0
}

// getDefaultDeviceInfo retorna info padrão para AMD Ryzen AI
func (dm *DirectML) getDefaultDeviceInfo() *DeviceInfo {
	return &DeviceInfo{
		Name:          "AMD Ryzen AI NPU",
		Vendor:        "AMD",
		DriverVersion: "1.0.0",
		MemoryMB:      0, // NPU compartilha memória do sistema
		TOPs:          55.0,
		IsNPU:         true,
		IsIntegrated:  true,
	}
}

// GetDeviceInfo retorna informações do dispositivo NPU
func (dm *DirectML) GetDeviceInfo() (*DeviceInfo, error) {
	if !dm.available {
		return nil, fmt.Errorf("DirectML não disponível")
	}

	if dm.deviceInfo != nil {
		return dm.deviceInfo, nil
	}

	return dm.getDefaultDeviceInfo(), nil
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
	dm.deviceID = deviceID
	return nil
}

// GetDeviceName retorna nome do dispositivo
func (dm *DirectML) GetDeviceName() string {
	if dm.deviceInfo != nil {
		return dm.deviceInfo.Name
	}
	return dm.deviceName
}

// GetTOPs retorna a capacidade em TOPS do NPU
func (dm *DirectML) GetTOPs() float64 {
	if dm.deviceInfo != nil {
		return dm.deviceInfo.TOPs
	}
	return 55.0 // Default para AMD Ryzen AI
}

// Close libera recursos
func (dm *DirectML) Close() error {
	// Libera factory se foi criada
	if dm.dxgiFactory != 0 {
		factoryVtbl := *(*[20]uintptr)(unsafe.Pointer(dm.dxgiFactory))
		syscall.SyscallN(factoryVtbl[2], dm.dxgiFactory) // Release
		dm.dxgiFactory = 0
	}
	return nil
}

// PrintDeviceInfo imprime informações do dispositivo
func (dm *DirectML) PrintDeviceInfo() {
	info, err := dm.GetDeviceInfo()
	if err != nil {
		fmt.Printf("Erro ao obter info do dispositivo: %v\n", err)
		return
	}

	fmt.Println("╔═══════════════════════════════════════════╗")
	fmt.Println("║           NPU Device Information          ║")
	fmt.Println("╠═══════════════════════════════════════════╣")
	fmt.Printf("║ Nome:    %-32s ║\n", info.Name)
	fmt.Printf("║ Vendor:  %-32s ║\n", info.Vendor)
	fmt.Printf("║ TOPs:    %-32.1f ║\n", info.TOPs)
	if info.MemoryMB > 0 {
		fmt.Printf("║ VRAM:    %-28d MB ║\n", info.MemoryMB)
	} else {
		fmt.Printf("║ VRAM:    %-32s ║\n", "Shared (System RAM)")
	}
	fmt.Printf("║ NPU:     %-32v ║\n", info.IsNPU)
	fmt.Println("╚═══════════════════════════════════════════╝")
}
