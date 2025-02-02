package main

import (
	"embed"
	"fmt"
	"genai"
	"html/template"
	"log"
	"net/http"
	"os/exec"

	"dev-team/internal/config"
	"dev-team/internal/handlers"
	"dev-team/internal/settings"
	"dev-team/internal/state"
	"dev-team/pkg/repository"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

//go:embed templates
var templatesFS embed.FS

var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	appState, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize GenAI provider
	genaiProvider, err := genai.NewProvider(appState.Settings.Provider, appState.Settings.APIKey)
	if err != nil {
		log.Printf("Error initializing GenAI provider: %v", err)
	}
	appState.GenAI = genaiProvider
	state.State = appState

	// Code validation
	log.Println("Running code validation...")

	// Run go fmt
	cmd := exec.Command("go", "fmt", "./...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("go fmt failed: %v\n%s", err, string(out))
	} else {
		log.Println("go fmt passed")
	}


	// Run go mod tidy
	cmd = exec.Command("go", "mod", "tidy")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("go mod tidy failed: %v\n%s", err, string(out))
	} else {
		log.Println("go mod tidy passed")
	}

	// Run golangci-lint
	cmd = exec.Command("./bin/golangci-lint", "run")
	out, err = cmd.CombinedOutput()
	if err != nil {
        log.Printf("golangci-lint failed: %v\n%s", err, string(out))
		fmt.Printf("golangci-lint failed: %v\n%s", err, string(out))
		return // Exit if lint fails
	}
	log.Println("golangci-lint passed")


	// Run go test
	cmd = exec.Command("go", "test", "./...")
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("go test failed: %v\n%s", err, string(out))
		fmt.Printf("go test failed: %v\n%s", err, string(out))
		return // Exit if tests fail
	}
    log.Println("go test passed")


	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/repositories", handlers.HandleListRepositories).Methods("GET")
	api.HandleFunc("/repositories", handlers.HandleAddRepository).Methods("POST")
	api.HandleFunc("/repositories", handlers.HandleDeleteRepository).Methods("DELETE")
	api.HandleFunc("/repositories/update", handlers.HandleUpdateRepository).Methods("POST")
	api.HandleFunc("/repositories/clone", handlers.HandleCloneRepository).Methods("POST")
	api.HandleFunc("/repositories/commit", handlers.HandleCommit).Methods("POST")
	api.HandleFunc("/repositories/push", handlers.HandlePush).Methods("POST")
	api.HandleFunc("/repositories/pr", handlers.HandleCreatePR).Methods("POST")
	api.HandleFunc("/repositories/sync", handlers.HandleSyncRepository).Methods("POST")
	api.HandleFunc("/settings", handlers.HandleGetSettings).Methods("GET")
	api.HandleFunc("/settings", handlers.HandleUpdateSettings).Methods("POST")
	api.HandleFunc("/gemini/models", handlers.HandleGeminiModels).Methods("GET")
	api.HandleFunc("/github/repositories", handlers.HandleGitHubRepositories).Methods("GET")
	api.HandleFunc("/github/issues", handlers.HandleGitHubIssues).Methods("GET")

	// Web routes
	r.HandleFunc("/", handleHome).Methods("GET")
	r.HandleFunc("/settings", handleSettingsPage).Methods("GET")

	// Configure CORS for API routes
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Start the scheduler
	state.State.Scheduler.Start()
	defer state.State.Scheduler.Stop()

	handler := c.Handler(r)
	log.Printf("Server starting on http://0.0.0.0:8083")
	log.Fatal(http.ListenAndServe("0.0.0.0:8083", handler))
}

type PageData struct {
	Page         string
	Repositories map[string]*repository.Repository
	Settings     settings.Settings
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	state.State.Mu.RLock()
	data := PageData{
		Page:         "home",
		Repositories: state.State.Repositories,
		Settings:     state.State.Settings,
	}
	state.State.Mu.RUnlock()

	err := templates.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleSettingsPage(w http.ResponseWriter, r *http.Request) {
	state.State.Mu.RLock()
	data := PageData{
		Page:         "settings",
		Repositories: state.State.Repositories,
		Settings:     state.State.Settings,
	}
	state.State.Mu.RUnlock()

	err := templates.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
