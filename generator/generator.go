package generator

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

type Generator struct {
	templates embed.FS
}

type ProjectConfig struct {
	// Core
	ProjectName string
	Module      string
	Description string
	GoVersion   string

	// Structure
	Structure   string // "standard", "flat", "feature", "hexagonal"
	ProjectType string // "rest-api", "cli", "grpc", "library"

	// Dependencies
	Router string // "chi", "gin", "echo", "fiber", "stdlib"
	Logger string // "zerolog", "zap", "slog", "logrus", "stdlib"

	// Optional Features
	UseDocker   bool
	UseGitHub   bool
	UseConfig   bool
	UseLogger   bool
	UseDatabase bool
	UseRedis    bool
	UseJWT      bool
	UseAir      bool

	// Dependencies list
	Dependencies []string
}

func New(templates embed.FS) *Generator {
	return &Generator{
		templates: templates,
	}
}

// Generate creates a zip file containing the generated project
func (g *Generator) Generate(config ProjectConfig) ([]byte, error) {
	// Create a buffer to write our zip to
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Get file mappings for the selected structure
	mappings := GetFileMappings(config.Structure)

	// Generate each file
	for _, mapping := range mappings {
		// Check condition
		if mapping.Condition != nil && !mapping.Condition(config) {
			continue
		}

		// Read template
		templatePath := "templates/" + mapping.TemplatePath
		templateData, err := g.templates.ReadFile(templatePath)
		if err != nil {
			// Skip files that don't exist
			continue
		}

		// Process output path (replace template variables)
		outputPath := g.processPath(mapping.OutputPath, config)

		// Parse and execute template
		tmpl, err := template.New(mapping.TemplatePath).Parse(string(templateData))
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", mapping.TemplatePath, err)
		}

		var content bytes.Buffer
		if err := tmpl.Execute(&content, config); err != nil {
			return nil, fmt.Errorf("failed to execute template %s: %w", mapping.TemplatePath, err)
		}

		// Add file to zip
		fullPath := filepath.Join(config.ProjectName, outputPath)
		f, err := zipWriter.Create(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create zip entry %s: %w", fullPath, err)
		}

		if _, err := f.Write(content.Bytes()); err != nil {
			return nil, fmt.Errorf("failed to write to zip entry %s: %w", fullPath, err)
		}
	}

	// Generate go.mod
	if err := g.generateGoMod(zipWriter, config); err != nil {
		return nil, fmt.Errorf("failed to generate go.mod: %w", err)
	}

	// Generate go.sum (empty file)
	goSumPath := filepath.Join(config.ProjectName, "go.sum")
	if _, err := zipWriter.Create(goSumPath); err != nil {
		return nil, fmt.Errorf("failed to create go.sum: %w", err)
	}

	// Close the zip writer
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// GetFileList returns a list of files that would be generated
func (g *Generator) GetFileList(config ProjectConfig) []string {
	var files []string

	mappings := GetFileMappings(config.Structure)

	for _, mapping := range mappings {
		if mapping.Condition != nil && !mapping.Condition(config) {
			continue
		}

		outputPath := g.processPath(mapping.OutputPath, config)
		files = append(files, outputPath)
	}

	files = append(files, "go.mod", "go.sum")

	return files
}

// processPath replaces template variables in the path
func (g *Generator) processPath(path string, config ProjectConfig) string {
	path = strings.ReplaceAll(path, "{{.ProjectName}}", config.ProjectName)
	return path
}

// generateGoMod creates a go.mod file with the appropriate dependencies
func (g *Generator) generateGoMod(zipWriter *zip.Writer, config ProjectConfig) error {
	goModPath := filepath.Join(config.ProjectName, "go.mod")
	f, err := zipWriter.Create(goModPath)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("module %s\n\n", config.Module))
	buf.WriteString(fmt.Sprintf("go %s\n", config.GoVersion))

	// Collect dependencies
	deps := g.getDependencies(config)
	if len(deps) > 0 {
		buf.WriteString("\nrequire (\n")
		for pkg, version := range deps {
			buf.WriteString(fmt.Sprintf("\t%s %s\n", pkg, version))
		}
		buf.WriteString(")\n")
	}

	_, err = f.Write(buf.Bytes())
	return err
}

// getDependencies returns a map of package -> version based on config
func (g *Generator) getDependencies(config ProjectConfig) map[string]string {
	deps := make(map[string]string)

	// Router dependencies
	switch config.Router {
	case "chi":
		deps["github.com/go-chi/chi/v5"] = "v5.0.11"
	case "gin":
		deps["github.com/gin-gonic/gin"] = "v1.9.1"
	case "echo":
		deps["github.com/labstack/echo/v4"] = "v4.11.4"
	case "fiber":
		deps["github.com/gofiber/fiber/v2"] = "v2.52.0"
	}

	// Logger dependencies
	switch config.Logger {
	case "zerolog":
		deps["github.com/rs/zerolog"] = "v1.32.0"
	case "zap":
		deps["go.uber.org/zap"] = "v1.26.0"
	case "logrus":
		deps["github.com/sirupsen/logrus"] = "v1.9.3"
	}

	// Database dependencies
	if config.UseDatabase {
		// Check if PostgreSQL is in dependencies
		hasPostgres := false
		for _, dep := range config.Dependencies {
			if strings.Contains(strings.ToLower(dep), "postgres") || strings.Contains(strings.ToLower(dep), "pgx") {
				hasPostgres = true
				deps["github.com/jackc/pgx/v5"] = "v5.5.1"
				break
			}
		}
		if !hasPostgres {
			// Default to pgx if database is enabled
			deps["github.com/jackc/pgx/v5"] = "v5.5.1"
		}
	}

	// Add UUID for hexagonal architecture (used in repository)
	if config.Structure == "hexagonal" {
		deps["github.com/google/uuid"] = "v1.5.0"
	}

	// Redis dependencies
	if config.UseRedis {
		deps["github.com/redis/go-redis/v9"] = "v9.4.0"
	}

	// JWT dependencies
	if config.UseJWT {
		deps["github.com/golang-jwt/jwt/v5"] = "v5.2.0"
	}

	// Process additional dependencies from the UI
	for _, dep := range config.Dependencies {
		// Map dependency names to packages
		switch dep {
		case "Chi Router":
			deps["github.com/go-chi/chi/v5"] = "v5.0.11"
		case "Gin Web Framework":
			deps["github.com/gin-gonic/gin"] = "v1.9.1"
		case "Echo":
			deps["github.com/labstack/echo/v4"] = "v4.11.4"
		case "Fiber":
			deps["github.com/gofiber/fiber/v2"] = "v2.52.0"
		case "Gorilla Mux":
			deps["github.com/gorilla/mux"] = "v1.8.1"
		case "Templ":
			deps["github.com/a-h/templ"] = "v0.2.543"
		case "Pongo2":
			deps["github.com/flosch/pongo2/v6"] = "v6.0.0"
		case "PostgreSQL Driver (pgx)":
			deps["github.com/jackc/pgx/v5"] = "v5.5.1"
		case "MySQL Driver":
			deps["github.com/go-sql-driver/mysql"] = "v1.7.1"
		case "GORM":
			deps["gorm.io/gorm"] = "v1.25.5"
		case "sqlx":
			deps["github.com/jmoiron/sqlx"] = "v1.3.5"
		case "SQLite Driver":
			deps["github.com/mattn/go-sqlite3"] = "v1.14.19"
		case "Redis Client (go-redis)":
			deps["github.com/redis/go-redis/v9"] = "v9.4.0"
		case "MongoDB Driver":
			deps["go.mongodb.org/mongo-driver"] = "v1.13.1"
		case "BadgerDB":
			deps["github.com/dgraph-io/badger/v4"] = "v4.2.0"
		case "Zerolog":
			deps["github.com/rs/zerolog"] = "v1.32.0"
		case "Zap":
			deps["go.uber.org/zap"] = "v1.26.0"
		case "Logrus":
			deps["github.com/sirupsen/logrus"] = "v1.9.3"
		case "Prometheus Client":
			deps["github.com/prometheus/client_golang"] = "v1.18.0"
		case "OpenTelemetry":
			deps["go.opentelemetry.io/otel"] = "v1.22.0"
		case "Jaeger Client":
			deps["github.com/jaegertracing/jaeger-client-go"] = "v2.30.0+incompatible"
		case "RabbitMQ Client":
			deps["github.com/rabbitmq/amqp091-go"] = "v1.9.0"
		case "Kafka Client (Sarama)":
			deps["github.com/IBM/sarama"] = "v1.42.2"
		case "NATS":
			deps["github.com/nats-io/nats.go"] = "v1.31.0"
		case "Gorilla WebSocket":
			deps["github.com/gorilla/websocket"] = "v1.5.1"
		case "JWT-Go":
			deps["github.com/golang-jwt/jwt/v5"] = "v5.2.0"
		case "Testify":
			deps["github.com/stretchr/testify"] = "v1.8.4"
		case "GoMock":
			deps["go.uber.org/mock"] = "v0.4.0"
		case "Ginkgo":
			deps["github.com/onsi/ginkgo/v2"] = "v2.15.0"
		}
	}

	return deps
}
