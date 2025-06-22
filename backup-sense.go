package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	backupBaseDir   = "./backup"
	dirPermissions  = 0755
	filePermissions = 0644
	timestampFormat = "20060102-150405"
	defaultPort     = 80
	defaultMaxMB    = 10
)

var maxUploadSize int64

type PfSenseConfig struct {
	XMLName xml.Name `xml:"pfsense"`
	System  struct {
		Hostname string `xml:"hostname"`
		Domain   string `xml:"domain"`
	} `xml:"system"`
}

type OPNsenseConfig struct {
	XMLName xml.Name `xml:"opnsense"`
	System  struct {
		Hostname string `xml:"hostname"`
	} `xml:"system"`
}

func main() {
	port := flag.Int("p", defaultPort, "Listening Port")
	maxMB := flag.Int("m", defaultMaxMB, "Maximum upload size in MB")
	flag.Parse()

	maxUploadSize = int64(*maxMB) << 20

	http.HandleFunc("/upload", handleUpload)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Backup server started on port %d...", *port)
	log.Printf("Max upload size: %d MB (%d bytes)", *maxMB, maxUploadSize)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)

	if !validateHTTPMethod(w, r) {
		return
	}

	// Set the max size for the entire request
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse form with a reasonable buffer size
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		handleError(w, "File too large", err, http.StatusBadRequest)
		return
	}

	fileBytes, err := readUploadedFile(r)
	if err != nil {
		handleError(w, "Error processing file", err, http.StatusBadRequest)
		return
	}

	hostname, err := parseFirewallConfig(fileBytes, clientIP)
	if err != nil {
		handleError(w, "Invalid configuration", err, http.StatusBadRequest)
		return
	}

	filePath, err := saveBackupFile(hostname, fileBytes)
	if err != nil {
		handleError(w, "Backup failed", err, http.StatusInternalServerError)
		return
	}

	sendSuccessResponse(w, clientIP, filePath)
}

func getClientIP(r *http.Request) string {
	ip := strings.Split(r.RemoteAddr, ":")[0]
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = strings.Split(forwarded, ",")[0]
	}
	return ip
}

func validateHTTPMethod(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func readUploadedFile(r *http.Request) ([]byte, error) {
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("error retrieving file: %w", err)
	}
	defer file.Close()

	// Additional validation of file size
	if fileHeader.Size > maxUploadSize {
		return nil, fmt.Errorf("file too large (max %d MB)", maxUploadSize>>20)
	}
	return io.ReadAll(file)
}

func parseFirewallConfig(fileBytes []byte, clientIP string) (string, error) {
	content := string(fileBytes)

	switch {
	case strings.Contains(content, "<pfsense>"):
		return parsePfSenseConfig(fileBytes)
	case strings.Contains(content, "<opnsense>"):
		return parseOPNSenseConfig(fileBytes)
	default:
		return "", fmt.Errorf("unsupported XML type")
	}
}

func parsePfSenseConfig(fileBytes []byte) (string, error) {
	var config PfSenseConfig
	if err := xml.Unmarshal(fileBytes, &config); err != nil {
		return "", fmt.Errorf("pfSense parse error: %w", err)
	}

	if config.System.Hostname == "" {
		return "", fmt.Errorf("missing hostname in pfSense config")
	}

	hostname := config.System.Hostname
	if config.System.Domain != "" {
		hostname += "." + config.System.Domain
	}
	return hostname, nil
}

func parseOPNSenseConfig(fileBytes []byte) (string, error) {
	var config OPNsenseConfig
	if err := xml.Unmarshal(fileBytes, &config); err != nil {
		return "", fmt.Errorf("OPNSense parse error: %w", err)
	}

	if config.System.Hostname == "" {
		return "", fmt.Errorf("missing hostname in OPNSense config")
	}
	return config.System.Hostname, nil
}

func saveBackupFile(hostname string, content []byte) (string, error) {
	backupDir := filepath.Join(backupBaseDir, hostname)
	if err := os.MkdirAll(backupDir, dirPermissions); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	filename := fmt.Sprintf("%s-%s.xml", hostname, time.Now().Format(timestampFormat))
	filePath := filepath.Join(backupDir, filename)

	if err := os.WriteFile(filePath, content, filePermissions); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return filePath, nil
}

func sendSuccessResponse(w http.ResponseWriter, clientIP, filePath string) {
	response := fmt.Sprintf("Backup received from %s Saved to: %s", clientIP, filePath)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Backup sent successfully\n"))
	log.Println(response)
}

func handleError(w http.ResponseWriter, message string, err error, status int) {
	log.Printf("%s: %v", message, err)
	http.Error(w, fmt.Sprintf("%s: %v", message, err), status)
}

func firstN(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
