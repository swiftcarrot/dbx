package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Generator is responsible for creating new migration files
type Generator struct {
	registry *Registry
}

// NewGenerator creates a new migration generator
func NewGenerator(registry *Registry) *Generator {
	return &Generator{
		registry: registry,
	}
}

// Generate creates a new migration file
func (g *Generator) Generate(name string) (string, error) {
	// Clean the name (no spaces, lowercase with underscores)
	name = strings.ToLower(strings.ReplaceAll(name, " ", "_"))

	// Generate version timestamp
	version := GenerateVersionTimestamp()

	// Create file path
	filename := fmt.Sprintf("%s_%s.go", version, name)
	filePath := filepath.Join(g.registry.GetMigrationsDir(), filename)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create migration file: %w", err)
	}
	defer file.Close()

	// Create migration from template
	migrationTemplate := template.Must(template.New("migration").Parse(migrationTmpl))

	packageName := filepath.Base(g.registry.GetMigrationsDir())

	data := struct {
		Version string
		Name    string
		Package string
	}{
		Version: version,
		Name:    name,
		Package: packageName,
	}

	if err := migrationTemplate.Execute(file, data); err != nil {
		return "", fmt.Errorf("failed to write migration file: %w", err)
	}

	return filePath, nil
}

// Template for new migration files
const migrationTmpl = `package {{ .Package }}

import (
	"github.com/swiftcarrot/dbx/migration"
	"github.com/swiftcarrot/dbx/schema"
)

func init() {
	migration.Register("{{ .Version }}", "{{ .Name }}", up{{ .Version }}, down{{ .Version }})
}

func up{{ .Version }}() *schema.Schema {
	s := schema.NewSchema()

	// Define your schema changes here
	// Example:
	// s.CreateTable("users", func(t *schema.Table) {
	//     t.Column("id", &schema.IntegerType{}, schema.PrimaryKey)
	//     t.Column("name", &schema.VarcharType{Length: 255})
	//     t.Column("email", &schema.VarcharType{Length: 255}, schema.NotNull)
	//     t.Column("created_at", &schema.TimestampType{})
	// })

	return s
}

func down{{ .Version }}() *schema.Schema {
	s := schema.NewSchema()

	// Define how to revert the changes here
	// Example:
	// s.DropTable("users")

	return s
}
`
