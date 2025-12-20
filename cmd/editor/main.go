package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/iceisfun/icescript/auxlib"
	"github.com/redis/go-redis/v9"
)

var (
	addr         = flag.String("addr", ":8080", "Address to listen on")
	redisHost    = flag.String("redis-host", "localhost", "Redis host")
	redisPort    = flag.Int("redis-port", 6379, "Redis port")
	scriptPrefix = flag.String("script-prefix", "icescript:", "Script prefix")
	staticDir    = flag.String("static", "./cmd/editor/static", "Path to static files")
)

func main() {
	flag.Parse()

	// 1. Setup Redis Client
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", *redisHost, *redisPort),
	})

	// 2. Setup Storage
	storage := auxlib.NewRedisStorage(rdb, *scriptPrefix)

	// 3. Setup Service
	svc := auxlib.NewService(storage)

	http.Handle("/", http.FileServer(http.Dir(*staticDir)))
	http.HandleFunc("/api/scripts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listScripts(w, r, svc)
		case http.MethodPost: // Create/Save
			http.Error(w, "Use POST /api/scripts/{name}", http.StatusMethodNotAllowed)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/scripts/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/scripts/")
		if name == "" {
			http.Error(w, "Script name required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			loadScript(w, r, svc, name)
		case http.MethodPost:
			saveScript(w, r, svc, name)
		case http.MethodDelete:
			deleteScript(w, r, svc, name)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		testScript(w, r, svc)
	})

	log.Printf("Listening on %s...", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func listScripts(w http.ResponseWriter, r *http.Request, svc auxlib.ScriptService) {
	scripts, err := svc.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sort.Strings(scripts)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scripts)
}

func loadScript(w http.ResponseWriter, r *http.Request, svc auxlib.ScriptService, name string) {
	content, err := svc.Load(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(content))
}

func saveScript(w http.ResponseWriter, r *http.Request, svc auxlib.ScriptService, name string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = svc.Save(r.Context(), name, string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func deleteScript(w http.ResponseWriter, r *http.Request, svc auxlib.ScriptService, name string) {
	err := svc.Delete(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func testScript(w http.ResponseWriter, r *http.Request, svc auxlib.ScriptService) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	result, err := svc.Test(r.Context(), string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
