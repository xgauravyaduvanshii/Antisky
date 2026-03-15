package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/net/websocket"
)

func main() {
	port := getEnv("PORT", "8091")

	log.Println("╔══════════════════════════════════════╗")
	log.Println("║   Antisky Terminal Proxy v1.0        ║")
	log.Println("╚══════════════════════════════════════╝")

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"healthy","service":"terminal-proxy"}`)
	})

	// WebSocket terminal endpoint
	mux.Handle("/ws/terminal", websocket.Handler(terminalHandler))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Printf("🖥️  Terminal proxy listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Terminal proxy failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Terminal proxy stopped")
}

func terminalHandler(ws *websocket.Conn) {
	defer ws.Close()

	log.Printf("Terminal session opened from %s", ws.Request().RemoteAddr)

	// Start a bash shell
	cmd := exec.Command("/bin/bash", "-i")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	// Get stdin pipe
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("Failed to get stdin: %v", err)
		return
	}

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to get stdout: %v", err)
		return
	}

	// Get stderr pipe
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Failed to get stderr: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start shell: %v", err)
		return
	}

	var wg sync.WaitGroup

	// stdout → WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				return
			}
			msg, _ := json.Marshal(map[string]string{"type": "output", "data": string(buf[:n])})
			websocket.Message.Send(ws, string(msg))
		}
	}()

	// stderr → WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := stderr.Read(buf)
			if err != nil {
				return
			}
			msg, _ := json.Marshal(map[string]string{"type": "output", "data": string(buf[:n])})
			websocket.Message.Send(ws, string(msg))
		}
	}()

	// WebSocket → stdin
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			var msg string
			if err := websocket.Message.Receive(ws, &msg); err != nil {
				cmd.Process.Kill()
				return
			}

			var input struct {
				Type string `json:"type"`
				Data string `json:"data"`
			}
			if err := json.Unmarshal([]byte(msg), &input); err != nil {
				// Raw text input
				stdin.Write([]byte(msg))
				continue
			}

			switch input.Type {
			case "input":
				stdin.Write([]byte(input.Data))
			case "resize":
				// TODO: Handle terminal resize
			}
		}
	}()

	cmd.Wait()
	wg.Wait()
	log.Printf("Terminal session closed from %s", ws.Request().RemoteAddr)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
