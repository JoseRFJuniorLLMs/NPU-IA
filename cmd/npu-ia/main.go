package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/audio"
	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/router"
	"github.com/JoseRFJuniorLLMs/NPU-IA/internal/tts"
	"github.com/JoseRFJuniorLLMs/NPU-IA/pkg/config"
)

const banner = `
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║    ███╗   ██╗██████╗ ██╗   ██╗      ██╗ █████╗               ║
║    ████╗  ██║██╔══██╗██║   ██║      ██║██╔══██╗              ║
║    ██╔██╗ ██║██████╔╝██║   ██║█████╗██║███████║              ║
║    ██║╚██╗██║██╔═══╝ ██║   ██║╚════╝██║██╔══██║              ║
║    ██║ ╚████║██║     ╚██████╔╝      ██║██║  ██║              ║
║    ╚═╝  ╚═══╝╚═╝      ╚═════╝       ╚═╝╚═╝  ╚═╝              ║
║                                                               ║
║    Assistente IA 100%% Local - Powered by AMD NPU             ║
║    55 TOPS | 32GB RAM | Zero Cloud                           ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
`

func main() {
	fmt.Println(banner)
	log.Println("Iniciando NPU-IA...")

	// Carrega configurações
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Printf("Usando configurações padrão: %v", err)
		cfg = config.Default()
	}

	// Contexto para graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Inicializa o Router (carrega modelos)
	log.Println("Carregando modelos na NPU...")
	r, err := router.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Erro ao inicializar router: %v", err)
	}
	defer r.Close()

	// Inicializa TTS
	speaker, err := tts.New(cfg.TTS)
	if err != nil {
		log.Fatalf("Erro ao inicializar TTS: %v", err)
	}
	defer speaker.Close()

	// Inicializa captura de áudio
	mic, err := audio.NewCapture(cfg.Audio)
	if err != nil {
		log.Fatalf("Erro ao inicializar microfone: %v", err)
	}
	defer mic.Close()

	// Sinal de pronto
	log.Println("✓ NPU-IA pronto! Ouvindo...")
	speaker.Speak("Olá! Estou pronto para ajudar.")

	// Loop principal
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Captura áudio
				audioData, err := mic.Listen()
				if err != nil {
					log.Printf("Erro ao capturar áudio: %v", err)
					continue
				}

				// Processa com o router
				response, err := r.Process(ctx, audioData)
				if err != nil {
					log.Printf("Erro ao processar: %v", err)
					speaker.Speak("Desculpe, não entendi.")
					continue
				}

				// Responde
				if response.Text != "" {
					speaker.Speak(response.Text)
				}
			}
		}
	}()

	// Aguarda sinal de shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("\nDesligando NPU-IA...")
	speaker.Speak("Até logo!")
}
