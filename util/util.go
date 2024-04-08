package util

import (
	"encoding/json"
	"fmt"
	"github.com/go-yaml/yaml"
	"golang.org/x/time/rate"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type Config struct {
	Protocol     string   `yaml:"protocol"`
	ListenAddr   string   `yaml:"listen_addr"`
	CertFile     string   `yaml:"cert_file"`
	KeyFile      string   `yaml:"key_file"`
	RateLimit    float64  `yaml:"rate_limit"`
	Blacklist    []string `yaml:"blacklist"`
	TemplateFile string   `yaml:"template_file"`
}

type VisitCounts struct {
	sync.RWMutex
	Counts map[string]int `yaml:"counts"`
}

type RateLimiter struct {
	sync.RWMutex
	Limiters map[string]*rate.Limiter
}

func (r *RateLimiter) getLimiter(ip string, limit rate.Limit) *rate.Limiter {
	r.Lock()
	defer r.Unlock()
	if limiter, ok := r.Limiters[ip]; ok {
		return limiter
	}
	newLimiter := rate.NewLimiter(limit, 1)
	r.Limiters[ip] = newLimiter
	return newLimiter
}

func LoadConfig(filename string) *Config {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading config file %s: %v", filename, err)
	}
	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		log.Fatalf("Error parsing config file %s: %v", filename, err)
	}
	return config
}

func OpenLogFile(filename string) *os.File {
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file %s: %v", filename, err)
	}
	return logFile
}

func LoadVisitCounts(filename string) *VisitCounts {
	visitCounts := &VisitCounts{Counts: make(map[string]int)}
	data, err := os.ReadFile(filename)
	if err == nil {
		err = yaml.Unmarshal(data, visitCounts)
		if err != nil {
			log.Printf("Error reading visit counts from %s: %v", filename, err)
		}
	} else if !os.IsNotExist(err) {
		log.Printf("Error reading visit counts from %s: %v", filename, err)
	}
	return visitCounts
}

func LoadTemplate(filename string) (*template.Template, error) {
	return template.ParseFiles(filename)
}
func HandleVisitCount(w http.ResponseWriter, r *http.Request, config *Config, visitCounts *VisitCounts, blacklist map[string]bool, limiter *RateLimiter, logFile *os.File) {
	ip := getIP(r)
	if blacklist[ip] {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	url := r.URL.Query().Get("url")
	visitCounts.Lock()
	count := visitCounts.Counts[url]
	visitCounts.Counts[url]++
	visitCounts.Unlock()
	if !limiter.getLimiter(ip, rate.Limit(config.RateLimit/60)).Allow() {
		log.Printf("IP %s exceeded rate limit", ip)
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(struct {
		Count int `json:"count"`
	}{
		Count: count,
	})
	if err != nil {
		log.Fatalf("Error encoding JSON response: %v", err)
		return
	}
	data, err := yaml.Marshal(visitCounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = os.WriteFile("data.yml", data, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logEntry := fmt.Sprintf("[%s] %s %s %s\n", time.Now().Format(time.RFC3339), ip, r.Method, r.URL.String())
	_, err = logFile.WriteString(logEntry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("%s %s %s Count: %d", ip, r.Method, r.URL.String(), count)
}
func HandleIndex(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}
	}
	cleanIp, _, err := net.SplitHostPort(ip)
	if err != nil {
		return ip
	}
	return cleanIp
}
