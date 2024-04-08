package execute

import (
	"ViewCounts/util"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"os"
)

func ExecuteInit() {

}

func Execute() {
	config := util.LoadConfig("config.yml")
	logFile := util.OpenLogFile("log.txt")
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Fatalf("Error closing log file: %v", err)
			return
		}
	}(logFile)
	visitCounts := util.LoadVisitCounts("data.yml")
	blacklist := make(map[string]bool, len(config.Blacklist))
	for _, ip := range config.Blacklist {
		blacklist[ip] = true
	}
	limiter := &util.RateLimiter{Limiters: make(map[string]*rate.Limiter)}
	tmpl, err := util.LoadTemplate(config.TemplateFile)
	if err != nil {
		log.Fatalf("Error loading template: %v", err)
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/visit-count", func(w http.ResponseWriter, r *http.Request) {
		util.HandleVisitCount(w, r, config, visitCounts, blacklist, limiter, logFile)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		util.HandleIndex(w, r, tmpl)
	})
	server := &http.Server{
		Addr:    config.ListenAddr,
		Handler: mux,
	}
	log.Printf("Server is running on %s://%s", config.Protocol, config.ListenAddr)
	if config.Protocol == "https" {
		err = server.ListenAndServeTLS(config.CertFile, config.KeyFile)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil {
		log.Fatal(err)
		return
	}
}
