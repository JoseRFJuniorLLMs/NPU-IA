package router

import (
	"log"
	"sync"
	"time"
)

// MemoryManager gerencia modelos na memória de forma inteligente
type MemoryManager struct {
	router    *Router
	lastUsed  map[string]time.Time
	mu        sync.RWMutex
	ttl       time.Duration // Tempo para descarregar modelo inativo
	ticker    *time.Ticker
	stopChan  chan struct{}

	// Modelos que NUNCA descarrega
	persistent map[string]bool
}

// NewMemoryManager cria um gerenciador de memória
func NewMemoryManager(router *Router, ttl time.Duration) *MemoryManager {
	mm := &MemoryManager{
		router:     router,
		lastUsed:   make(map[string]time.Time),
		ttl:        ttl,
		stopChan:   make(chan struct{}),
		persistent: map[string]bool{
			"whisper": true, // STT sempre ativo
			"phi":     true, // Modelo rápido sempre ativo
		},
	}

	// Inicia goroutine de limpeza
	mm.ticker = time.NewTicker(30 * time.Second)
	go mm.cleanupLoop()

	return mm
}

// Touch marca modelo como usado
func (mm *MemoryManager) Touch(name string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.lastUsed[name] = time.Now()
}

// cleanupLoop verifica periodicamente modelos inativos
func (mm *MemoryManager) cleanupLoop() {
	for {
		select {
		case <-mm.ticker.C:
			mm.cleanup()
		case <-mm.stopChan:
			return
		}
	}
}

// cleanup descarrega modelos inativos
func (mm *MemoryManager) cleanup() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	now := time.Now()

	for name, lastUsed := range mm.lastUsed {
		// Pula modelos persistentes
		if mm.persistent[name] {
			continue
		}

		// Verifica se passou do TTL
		if now.Sub(lastUsed) > mm.ttl {
			mm.unloadModel(name)
			delete(mm.lastUsed, name)
		}
	}
}

// unloadModel descarrega um modelo da memória
func (mm *MemoryManager) unloadModel(name string) {
	log.Printf("💤 Descarregando %s (inativo por %v)", name, mm.ttl)

	mm.router.mu.Lock()
	defer mm.router.mu.Unlock()

	switch name {
	case "llama":
		if mm.router.llama != nil {
			mm.router.llama.Close()
			mm.router.llama = nil
			mm.router.loaded["llama"] = false
		}
	case "qwen":
		if mm.router.qwen != nil {
			mm.router.qwen.Close()
			mm.router.qwen = nil
			mm.router.loaded["qwen"] = false
		}
	case "vision":
		if mm.router.vision != nil {
			mm.router.vision.Close()
			mm.router.vision = nil
			mm.router.loaded["vision"] = false
		}
	case "coder":
		if mm.router.coder != nil {
			mm.router.coder.Close()
			mm.router.coder = nil
			mm.router.loaded["coder"] = false
		}
	}

	// Força garbage collection
	// runtime.GC()
}

// GetStats retorna estatísticas de memória
func (mm *MemoryManager) GetStats() map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	stats := make(map[string]interface{})

	loaded := []string{}
	for name, isLoaded := range mm.router.loaded {
		if isLoaded {
			loaded = append(loaded, name)
		}
	}
	stats["loaded_models"] = loaded
	stats["total_loaded"] = len(loaded)

	// Tempo desde último uso
	lastUsed := make(map[string]string)
	for name, t := range mm.lastUsed {
		lastUsed[name] = time.Since(t).String()
	}
	stats["last_used"] = lastUsed

	return stats
}

// Stop para o gerenciador
func (mm *MemoryManager) Stop() {
	mm.ticker.Stop()
	close(mm.stopChan)
}

// SetPersistent define se um modelo deve ficar sempre na memória
func (mm *MemoryManager) SetPersistent(name string, persistent bool) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.persistent[name] = persistent
}

// SetTTL define o tempo de inatividade para descarregar
func (mm *MemoryManager) SetTTL(ttl time.Duration) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.ttl = ttl
}
