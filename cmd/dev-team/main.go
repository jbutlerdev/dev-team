package main

import (
	"embed"
	"genai"
	"html/template"
	"log"
	"net/http"

	"dev-team/internal/config"
	"dev-team/internal/handlers"
	"dev-team/internal/settings"
	"dev-team/internal/state"
	"dev-team/pkg/repository"

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

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/repositories", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "GET":
            handlers.HandleListRepositories(w, r)
        case "POST":
            handlers.HandleAddRepository(w, r)
        case "DELETE":
            handlers.HandleDeleteRepository(w, r)
        }
    })
	mux.HandleFunc("/api/repositories/update", handlers.HandleUpdateRepository).Methods("POST")
	mux.HandleFunc("/api/repositories/clone", handlers.HandleCloneRepository).Methods("POST")
	mux.HandleFunc("/api/repositories/commit", handlers.HandleCommit).Methods("POST")
	mux.HandleFunc("/api/repositories/push", handlers.HandlePush).Methods("POST")
	mux.HandleFunc("/api/repositories/pr", handlers.HandleCreatePR).Methods("POST")
	mux.HandleFunc("/api/repositories/sync", handlers.HandleSyncRepository).Methods("POST")
	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "GET":
            handlers.HandleGetSettings(w, r)
        case "POST":
             handlers.HandleUpdateSettings(w, r)
        }
    })
	mux.HandleFunc("/api/gemini/models", handlers.HandleGeminiModels).Methods("GET")
    mux.HandleFunc("/api/github/repositories", handlers.HandleGitHubRepositories).Methods("GET")
	mux.HandleFunc("/api/github/issues", handlers.HandleGitHubIssues).Methods("GET")

	// Web routes
	mux.HandleFunc("/", handleHome).Methods("GET")
	mux.HandleFunc("/settings", handleSettingsPage).Methods("GET")

	// Configure CORS for API routes
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Start the scheduler
	state.State.Scheduler.Start()
	defer state.State.Scheduler.Stop()

	handler := c.Handler(mux)
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
