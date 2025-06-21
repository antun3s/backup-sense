package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Estrutura para parse do pfSense
type PfSenseConfig struct {
	XMLName xml.Name `xml:"pfsense"`
	System  struct {
		Hostname string `xml:"hostname"`
		Domain   string `xml:"domain"` // Adicionado campo domain
	} `xml:"system"`
}

// Estrutura para parse do OPNSense
type OPNsenseConfig struct {
	XMLName xml.Name `xml:"opnsense"`
	System  struct {
		Hostname string `xml:"hostname"`
	} `xml:"system"`
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	// 1. Validar método HTTP
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// 2. Limitar tamanho do upload (10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "O arquivo é muito grande", http.StatusBadRequest)
		return
	}

	// 3. Extrair arquivo do formulário
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Erro ao obter o arquivo: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 4. Ler conteúdo do arquivo
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Erro na leitura do arquivo: %v", err)
		http.Error(w, "Erro ao processar o arquivo", http.StatusInternalServerError)
		return
	}

	// 5. Validar estrutura XML e detectar tipo
	var hostname string
	fileContent := string(fileBytes)

	switch {
	case strings.Contains(fileContent, "<pfsense>"):
		var config PfSenseConfig
		if err := xml.Unmarshal(fileBytes, &config); err != nil {
			log.Printf("Erro no parse XML pfSense: %v", err)
			http.Error(w, "XML pfSense inválido", http.StatusBadRequest)
			return
		}
		// Concatena hostname e domain com ponto
		if config.System.Domain != "" {
			hostname = config.System.Hostname + "." + config.System.Domain
		} else {
			hostname = config.System.Hostname
		}

	case strings.Contains(fileContent, "<opnsense>"):
		var config OPNsenseConfig
		if err := xml.Unmarshal(fileBytes, &config); err != nil {
			log.Printf("Erro no parse XML OPNSense: %v", err)
			http.Error(w, "XML OPNSense inválido", http.StatusBadRequest)
			return
		}
		hostname = config.System.Hostname

	default:
		http.Error(w, "Tipo de firewall não suportado", http.StatusBadRequest)
		return
	}

	// 6. Validar hostname extraído
	if hostname == "" {
		http.Error(w, "Hostname não encontrado no XML", http.StatusBadRequest)
		return
	}

	// 7. Criar diretório de backup
	backupDir := filepath.Join("./backup", hostname)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Printf("Erro ao criar diretório: %v", err)
		http.Error(w, "Falha ao criar diretório de backup", http.StatusInternalServerError)
		return
	}

	// 8. Gerar nome do arquivo com timestamp
	timestamp := time.Now().Format("20060102-1504")
	fileName := fmt.Sprintf("%s-%s.xml", hostname, timestamp)
	filePath := filepath.Join(backupDir, fileName)

	// 9. Salvar arquivo
	if err := os.WriteFile(filePath, fileBytes, 0644); err != nil {
		log.Printf("Erro ao salvar arquivo: %v", err)
		http.Error(w, "Falha ao salvar backup", http.StatusInternalServerError)
		return
	}

	// 10. Resposta de sucesso
	response := fmt.Sprintf("Backup recebido e armazenado em: %s", filePath)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
	log.Println(response)
}

func main() {
	// Configurar endpoint
	http.HandleFunc("/upload", handleUpload)

	// Iniciar servidor na porta 80
	log.Println("Servidor de backup iniciado na porta 80...")
	log.Fatal(http.ListenAndServe(":80", nil))
}
