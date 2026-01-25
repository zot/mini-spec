// CRC: crc-Project.md | Seq: seq-init.md
package project

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds project configuration
type Config struct {
	DesignDir       string            `yaml:"design_dir"`
	SrcDir          string            `yaml:"src_dir"`
	CodeExtensions  []string          `yaml:"code_extensions"`
	CommentPatterns map[string]string `yaml:"comment_patterns"`
}

// Project represents a mini-spec project
type Project struct {
	RootPath  string
	DesignDir string
	SrcDir    string
	Config    Config
}

// DefaultCommentPatterns returns the default comment patterns per file extension
func DefaultCommentPatterns() map[string]string {
	return map[string]string{
		".go":   `//\s*`,
		".js":   `//\s*`,
		".ts":   `//\s*`,
		".c":    `//\s*`,
		".h":    `//\s*`,
		".cpp":  `//\s*`,
		".py":   `#\s*`,
		".lua":  `--\s*`,
		".sh":   `#\s*`,
		".bash": `#\s*`,
	}
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		DesignDir:       "design",
		SrcDir:          "src",
		CodeExtensions:  []string{".go", ".ts", ".js", ".lua", ".py", ".c", ".h", ".cpp", ".sh"},
		CommentPatterns: DefaultCommentPatterns(),
	}
}

// Detect finds a mini-spec project by walking up from cwd
func Detect() (*Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return DetectFrom(cwd)
}

// DetectFrom finds a mini-spec project by walking up from the given path
func DetectFrom(startPath string) (*Project, error) {
	dir := startPath
	for {
		designPath := filepath.Join(dir, "design")
		if info, err := os.Stat(designPath); err == nil && info.IsDir() {
			return loadProject(dir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil, fmt.Errorf("no design/ directory found (searched from %s)", startPath)
}

func loadProject(rootPath string) (*Project, error) {
	config := DefaultConfig()

	// Try to load .minispec.yaml
	configPath := filepath.Join(rootPath, ".minispec.yaml")
	if data, err := os.ReadFile(configPath); err == nil {
		var userConfig Config
		if err := yaml.Unmarshal(data, &userConfig); err != nil {
			return nil, fmt.Errorf("invalid .minispec.yaml: %w", err)
		}
		// Merge user config into defaults
		if userConfig.DesignDir != "" {
			config.DesignDir = userConfig.DesignDir
		}
		if userConfig.SrcDir != "" {
			config.SrcDir = userConfig.SrcDir
		}
		if len(userConfig.CodeExtensions) > 0 {
			config.CodeExtensions = userConfig.CodeExtensions
		}
		// Merge comment patterns: user patterns override defaults
		for ext, pattern := range userConfig.CommentPatterns {
			config.CommentPatterns[ext] = pattern
		}
	}

	return &Project{
		RootPath:  rootPath,
		DesignDir: filepath.Join(rootPath, config.DesignDir),
		SrcDir:    filepath.Join(rootPath, config.SrcDir),
		Config:    config,
	}, nil
}

// DesignPath resolves a path within the design directory
func (p *Project) DesignPath(filename string) string {
	return filepath.Join(p.DesignDir, filename)
}

// SrcPath resolves a path within the src directory
func (p *Project) SrcPath(filename string) string {
	return filepath.Join(p.SrcDir, filename)
}

// RequirementsPath returns the path to requirements.md
func (p *Project) RequirementsPath() string {
	return p.DesignPath("requirements.md")
}

// DesignMdPath returns the path to design.md
func (p *Project) DesignMdPath() string {
	return p.DesignPath("design.md")
}

// GlobCRCCards returns all CRC card files
func (p *Project) GlobCRCCards() ([]string, error) {
	return filepath.Glob(p.DesignPath("crc-*.md"))
}

// CommentPattern returns the comment regex pattern for the given file extension.
// Returns empty string if no pattern is configured for the extension.
func (p *Project) CommentPattern(ext string) string {
	return p.Config.CommentPatterns[ext]
}
