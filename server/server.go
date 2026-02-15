package server

import (
	"bytes"
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/thirukguru/go-initializer/generator"
)

type Server struct {
	webFiles         embed.FS
	projectTemplates embed.FS
	generator        *generator.Generator
}

func New(webFiles, projectTemplates embed.FS) *Server {
	return &Server{
		webFiles:         webFiles,
		projectTemplates: projectTemplates,
		generator:        generator.New(projectTemplates),
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Serve static files
	staticFS, err := fs.Sub(s.webFiles, "web/static")
	if err != nil {
		log.Printf("Warning: Could not load static files: %v", err)
	} else {
		r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	}

	// Serve the main HTML page
	r.Get("/", s.handleIndex)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Post("/generate", s.handleGenerate)
		r.Post("/preview", s.handlePreview)
	})

	return r
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := s.webFiles.ReadFile("web/templates/index.html")
	if err != nil {
		http.Error(w, "Could not load page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

type Dependency struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Desc     string `json:"desc"`
	Pkg      string `json:"pkg"`
}
type GenerateRequest struct {
	// Core
	ProjectName string `json:"project_name"`
	Module      string `json:"module"`
	Description string `json:"description"`
	GoVersion   string `json:"go_version"`

	// Structure
	Structure   string `json:"structure"`
	ProjectType string `json:"project_type"`

	// Dependencies
	Router string `json:"router"`
	Logger string `json:"logger"`

	// Optional Features
	UseDocker   bool `json:"use_docker"`
	UseGitHub   bool `json:"use_github"`
	UseConfig   bool `json:"use_config"`
	UseLogger   bool `json:"use_logger"`
	UseDatabase bool `json:"use_database"`
	UseRedis    bool `json:"use_redis"`
	UseJWT      bool `json:"use_jwt"`
	UseAir      bool `json:"use_air"`

	// Dependencies array
	Dependencies []Dependency `json:"dependencies"`
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	// Read body fully first
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read body: %v", err)
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // Reset for decode

	log.Printf("Raw body: %s", bodyBytes) // This WILL print

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Decode error: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Received generate request: %+v", req) // Full req now logs
	// Validate request
	if req.ProjectName == "" {
		http.Error(w, "Project name is required", http.StatusBadRequest)
		return
	}
	if req.Module == "" {
		http.Error(w, "Module path is required", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.GoVersion == "" {
		req.GoVersion = "1.26.0"
	}
	if req.Structure == "" {
		req.Structure = "standard"
	}
	if req.ProjectType == "" {
		req.ProjectType = "rest-api"
	}
	if req.Router == "" {
		req.Router = "chi"
	}

	// Convert to generator config
	config := generator.ProjectConfig{
		ProjectName:  req.ProjectName,
		Module:       req.Module,
		Description:  req.Description,
		GoVersion:    req.GoVersion,
		Structure:    req.Structure,
		ProjectType:  req.ProjectType,
		Router:       req.Router,
		Logger:       req.Logger,
		UseDocker:    req.UseDocker,
		UseGitHub:    req.UseGitHub,
		UseConfig:    req.UseConfig,
		UseLogger:    req.UseLogger,
		UseDatabase:  req.UseDatabase,
		UseRedis:     req.UseRedis,
		UseJWT:       req.UseJWT,
		UseAir:       req.UseAir,
		Dependencies: []string{}, // Empty slice
	}

	for _, dep := range req.Dependencies {
		config.Dependencies = append(config.Dependencies, dep.Pkg)
	}

	log.Printf("Extracted deps: %v", config.Dependencies)
	zipData, err := s.generator.Generate(config)
	if err != nil {
		log.Printf("Error generating project: %v", err)
		http.Error(w, "Failed to generate project", http.StatusInternalServerError)
		return
	}

	// Send zip file
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+req.ProjectName+".zip")
	w.Write(zipData)
}

type PreviewResponse struct {
	Files []FilePreview `json:"files"`
}

type FilePreview struct {
	Path string `json:"path"`
	Size int    `json:"size"`
}

func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Structure == "" {
		req.Structure = "standard"
	}

	// Convert to generator config
	config := generator.ProjectConfig{
		ProjectName:  req.ProjectName,
		Module:       req.Module,
		Description:  req.Description,
		GoVersion:    req.GoVersion,
		Structure:    req.Structure,
		ProjectType:  req.ProjectType,
		Router:       req.Router,
		Logger:       req.Logger,
		UseDocker:    req.UseDocker,
		UseGitHub:    req.UseGitHub,
		UseConfig:    req.UseConfig,
		UseLogger:    req.UseLogger,
		UseDatabase:  req.UseDatabase,
		UseRedis:     req.UseRedis,
		UseJWT:       req.UseJWT,
		UseAir:       req.UseAir,
		Dependencies: make([]string, len(req.Dependencies)),
	}
	for i, dep := range req.Dependencies {
		config.Dependencies[i] = dep.Pkg // Use actual import path
	}

	// Get file list
	files := s.generator.GetFileList(config)

	var previews []FilePreview
	for _, file := range files {
		previews = append(previews, FilePreview{
			Path: file,
			Size: 0, // Size would be calculated if needed
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PreviewResponse{
		Files: previews,
	})
}
