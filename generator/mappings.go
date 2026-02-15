package generator

// TemplateMapping defines which template files map to which output files
// based on project structure and configuration

type FileMapping struct {
	TemplatePath string
	OutputPath   string
	Condition    func(config ProjectConfig) bool // Optional: only include if condition is true
}

/*type ProjectConfig struct {
	Structure    string // "standard", "flat", "feature", "hexagonal"
	ProjectType  string // "rest-api", "cli", "grpc", "library"
	ProjectName  string
	Module       string
	Description  string
	GoVersion    string
	Router       string // "chi", "gin", "echo", "fiber", "stdlib"
	Logger       string // "zerolog", "zap", "slog", "logrus", "stdlib"
	UseDocker    bool
	UseGitHub    bool
	UseConfig    bool
	UseLogger    bool
	UseDatabase  bool
	UseRedis     bool
	UseJWT       bool
	UseAir       bool
}*/

// GetFileMappings returns the file mappings for a given project structure
func GetFileMappings(structure string) []FileMapping {
	switch structure {
	case "standard":
		return standardLayoutMappings()
	case "flat":
		return flatLayoutMappings()
	case "feature":
		return featureLayoutMappings()
	case "hexagonal":
		return hexagonalLayoutMappings()
	default:
		return standardLayoutMappings()
	}
}

func standardLayoutMappings() []FileMapping {
	return []FileMapping{
		// Main application
		{
			TemplatePath: "standard/cmd_main.go.tmpl",
			OutputPath:   "cmd/{{.ProjectName}}/main.go",
		},
		// Internal packages
		{
			TemplatePath: "standard/internal_handler.go.tmpl",
			OutputPath:   "internal/handler/handler.go",
			Condition:    func(c ProjectConfig) bool { return c.ProjectType == "rest-api" },
		},
		{
			TemplatePath: "standard/internal_config.go.tmpl",
			OutputPath:   "internal/config/config.go",
			Condition:    func(c ProjectConfig) bool { return c.UseConfig },
		},
		{
			TemplatePath: "standard/internal_middleware.go.tmpl",
			OutputPath:   "internal/middleware/logger.go",
			Condition:    func(c ProjectConfig) bool { return c.UseLogger && c.Router == "chi" },
		},
		// Pkg (shared libraries)
		{
			TemplatePath: "standard/pkg_logger.go.tmpl",
			OutputPath:   "pkg/logger/logger.go",
			Condition:    func(c ProjectConfig) bool { return c.UseLogger },
		},
		// Configuration files
		{
			TemplatePath: "standard/README.md.tmpl",
			OutputPath:   "README.md",
		},
		{
			TemplatePath: "standard/Makefile.tmpl",
			OutputPath:   "Makefile",
		},
		{
			TemplatePath: "standard/gitignore.tmpl",
			OutputPath:   ".gitignore",
		},
		{
			TemplatePath: "standard/env.example.tmpl",
			OutputPath:   ".env.example",
		},
		// Docker
		{
			TemplatePath: "standard/Dockerfile.tmpl",
			OutputPath:   "Dockerfile",
			Condition:    func(c ProjectConfig) bool { return c.UseDocker },
		},
		{
			TemplatePath: "standard/docker-compose.yaml.tmpl",
			OutputPath:   "docker-compose.yaml",
			Condition:    func(c ProjectConfig) bool { return c.UseDocker },
		},
		// CI/CD
		{
			TemplatePath: "standard/github_ci.yaml.tmpl",
			OutputPath:   ".github/workflows/ci.yml",
			Condition:    func(c ProjectConfig) bool { return c.UseGitHub },
		},
	}
}

func flatLayoutMappings() []FileMapping {
	return []FileMapping{
		{
			TemplatePath: "flat/main.go.tmpl",
			OutputPath:   "main.go",
		},
		{
			TemplatePath: "flat/README.md.tmpl",
			OutputPath:   "README.md",
		},
		{
			TemplatePath: "standard/gitignore.tmpl",
			OutputPath:   ".gitignore",
		},
		{
			TemplatePath: "standard/Dockerfile.tmpl",
			OutputPath:   "Dockerfile",
			Condition:    func(c ProjectConfig) bool { return c.UseDocker },
		},
	}
}

func featureLayoutMappings() []FileMapping {
	return []FileMapping{
		// Main
		{
			TemplatePath: "feature/cmd_main.go.tmpl",
			OutputPath:   "cmd/{{.ProjectName}}/main.go",
		},
		// User feature
		{
			TemplatePath: "feature/user_handler.go.tmpl",
			OutputPath:   "internal/user/handler.go",
		},
		{
			TemplatePath: "feature/user_service.go.tmpl",
			OutputPath:   "internal/user/service.go",
		},
		{
			TemplatePath: "feature/user_repository.go.tmpl",
			OutputPath:   "internal/user/repository.go",
		},
		{
			TemplatePath: "feature/user_model.go.tmpl",
			OutputPath:   "internal/user/model.go",
		},
		// Config
		{
			TemplatePath: "standard/internal_config.go.tmpl",
			OutputPath:   "pkg/config/config.go",
		},
		// Logger
		{
			TemplatePath: "standard/pkg_logger.go.tmpl",
			OutputPath:   "pkg/logger/logger.go",
			Condition:    func(c ProjectConfig) bool { return c.UseLogger },
		},
		// Root files
		{
			TemplatePath: "standard/README.md.tmpl",
			OutputPath:   "README.md",
		},
		{
			TemplatePath: "standard/Makefile.tmpl",
			OutputPath:   "Makefile",
		},
		{
			TemplatePath: "standard/gitignore.tmpl",
			OutputPath:   ".gitignore",
		},
		{
			TemplatePath: "standard/env.example.tmpl",
			OutputPath:   ".env.example",
		},
		{
			TemplatePath: "standard/Dockerfile.tmpl",
			OutputPath:   "Dockerfile",
			Condition:    func(c ProjectConfig) bool { return c.UseDocker },
		},
		{
			TemplatePath: "standard/docker-compose.yaml.tmpl",
			OutputPath:   "docker-compose.yaml",
			Condition:    func(c ProjectConfig) bool { return c.UseDocker },
		},
	}
}

func hexagonalLayoutMappings() []FileMapping {
	return []FileMapping{
		// Main application
		{
			TemplatePath: "hexagonal/cmd_main.go.tmpl",
			OutputPath:   "cmd/{{.ProjectName}}/main.go",
		},
		// Core - Domain
		{
			TemplatePath: "hexagonal/domain_user.go.tmpl",
			OutputPath:   "internal/core/domain/user.go",
		},
		// Core - Ports
		{
			TemplatePath: "hexagonal/port_repository.go.tmpl",
			OutputPath:   "internal/core/port/repository.go",
		},
		// Core - Services
		{
			TemplatePath: "hexagonal/service_user.go.tmpl",
			OutputPath:   "internal/core/service/user.go",
		},
		// Adapters - HTTP Handler
		{
			TemplatePath: "hexagonal/adapter_http_handler.go.tmpl",
			OutputPath:   "internal/adapters/http/handler/user.go",
			Condition:    func(c ProjectConfig) bool { return c.ProjectType == "rest-api" },
		},
		// Adapters - Repository
		{
			TemplatePath: "hexagonal/adapter_repository.go.tmpl",
			OutputPath:   "internal/adapters/repository/user.go",
		},
		// Infrastructure - Config
		{
			TemplatePath: "hexagonal/infra_config.go.tmpl",
			OutputPath:   "internal/infrastructure/config/config.go",
		},
		// Infrastructure - Logger
		{
			TemplatePath: "hexagonal/infra_logger.go.tmpl",
			OutputPath:   "internal/infrastructure/logger/logger.go",
			Condition:    func(c ProjectConfig) bool { return c.UseLogger },
		},
		// Documentation
		{
			TemplatePath: "hexagonal/README.md.tmpl",
			OutputPath:   "README.md",
		},
		// Configuration files (reuse from standard)
		{
			TemplatePath: "standard/Makefile.tmpl",
			OutputPath:   "Makefile",
		},
		{
			TemplatePath: "standard/gitignore.tmpl",
			OutputPath:   ".gitignore",
		},
		{
			TemplatePath: "standard/env.example.tmpl",
			OutputPath:   ".env.example",
		},
		{
			TemplatePath: "standard/Dockerfile.tmpl",
			OutputPath:   "Dockerfile",
			Condition:    func(c ProjectConfig) bool { return c.UseDocker },
		},
		{
			TemplatePath: "standard/docker-compose.yaml.tmpl",
			OutputPath:   "docker-compose.yaml",
			Condition:    func(c ProjectConfig) bool { return c.UseDocker },
		},
		{
			TemplatePath: "standard/github_ci.yaml.tmpl",
			OutputPath:   ".github/workflows/ci.yml",
			Condition:    func(c ProjectConfig) bool { return c.UseGitHub },
		},
	}
}
