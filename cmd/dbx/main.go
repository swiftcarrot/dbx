package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/swiftcarrot/dbx/migration"
)

const defaultMigrationsDir = "./migrations"

func main() {
	// Define flags
	migrationsDir := flag.String("migrations-dir", defaultMigrationsDir, "Directory containing migrations")
	dbUrl := flag.String("database", "", "Database connection URL")

	// Parse command line arguments
	flag.Parse()

	// Set up the migrations directory
	migration.SetMigrationsDir(*migrationsDir)

	// Get the subcommand
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]
	switch command {
	case "generate", "g":
		if len(args) < 2 {
			fmt.Println("Error: Missing migration name")
			fmt.Println("Usage: dbx generate <name>")
			os.Exit(1)
		}
		generateMigration(args[1])
	case "migrate", "m":
		if *dbUrl == "" {
			fmt.Println("Error: Database URL is required")
			fmt.Println("Usage: dbx migrate --database <url>")
			os.Exit(1)
		}
		var version string
		if len(args) > 1 {
			version = args[1]
		}
		runMigrations(*dbUrl, version)
	case "rollback", "r":
		if *dbUrl == "" {
			fmt.Println("Error: Database URL is required")
			fmt.Println("Usage: dbx rollback --database <url> [steps]")
			os.Exit(1)
		}
		steps := 1
		if len(args) > 1 {
			s, err := strconv.Atoi(args[1])
			if err == nil {
				steps = s
			}
		}
		rollbackMigrations(*dbUrl, steps)
	case "status", "s":
		if *dbUrl == "" {
			fmt.Println("Error: Database URL is required")
			fmt.Println("Usage: dbx status --database <url>")
			os.Exit(1)
		}
		showStatus(*dbUrl)
	case "help", "h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("DBX Migration Tool")
	fmt.Println("Usage:")
	fmt.Println("  dbx [options] <command> [arguments]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --migrations-dir <dir>   Directory containing migrations (default: ./migrations)")
	fmt.Println("  --database <url>         Database connection URL")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  generate, g <name>       Generate a new migration")
	fmt.Println("  migrate, m [version]     Run migrations (up to optional version)")
	fmt.Println("  rollback, r [steps]      Rollback migrations (default: 1 step)")
	fmt.Println("  status, s                Show migration status")
	fmt.Println("  help, h                  Show this help")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  dbx generate create_users")
	fmt.Println("  dbx --database \"postgres://user:pass@localhost/dbname\" migrate")
	fmt.Println("  dbx --database \"mysql://user:pass@localhost/dbname\" rollback 2")
}

func generateMigration(name string) {
	filePath, err := migration.CreateMigration(name)
	if err != nil {
		fmt.Printf("Error generating migration: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created migration: %s\n", filePath)
}

func connectToDatabase(dbUrl string) (*sql.DB, error) {
	// Parse the URL to determine the driver
	var driver string
	if strings.HasPrefix(dbUrl, "postgres://") {
		driver = "postgres"
	} else if strings.HasPrefix(dbUrl, "mysql://") {
		driver = "mysql"
		// Convert mysql URL to DSN format if needed
		dbUrl = strings.TrimPrefix(dbUrl, "mysql://")
		dbUrl = strings.Replace(dbUrl, "/", "?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci&database=", 1)
	} else if strings.HasPrefix(dbUrl, "sqlite://") {
		driver = "sqlite3"
		dbUrl = strings.TrimPrefix(dbUrl, "sqlite://")
	} else {
		return nil, fmt.Errorf("unsupported database URL: must start with postgres://, mysql:// or sqlite://")
	}

	// Connect to the database
	db, err := sql.Open(driver, dbUrl)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func runMigrations(dbUrl, version string) {
	db, err := connectToDatabase(dbUrl)
	if err != nil {
		fmt.Printf("Error connecting to database: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	err = migration.RunMigrations(db, version)
	if err != nil {
		fmt.Printf("Error running migrations: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Migrations applied successfully")
}

func rollbackMigrations(dbUrl string, steps int) {
	db, err := connectToDatabase(dbUrl)
	if err != nil {
		fmt.Printf("Error connecting to database: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	err = migration.RollbackMigration(db, steps)
	if err != nil {
		fmt.Printf("Error rolling back migrations: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Migrations rolled back successfully")
}

func showStatus(dbUrl string) {
	db, err := connectToDatabase(dbUrl)
	if err != nil {
		fmt.Printf("Error connecting to database: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	status, err := migration.GetMigrationStatus(db)
	if err != nil {
		fmt.Printf("Error getting migration status: %s\n", err)
		os.Exit(1)
	}

	if len(status) == 0 {
		fmt.Println("No migrations found")
		return
	}

	// Print status table
	fmt.Println("Migration Status:")
	fmt.Println("--------------------------------------------------------------------------------------------------------")
	fmt.Printf("%-14s | %-50s | %-10s | %s\n", "Version", "Name", "Status", "Applied At")
	fmt.Println("--------------------------------------------------------------------------------------------------------")

	for _, s := range status {
		fmt.Printf("%-14s | %-50s | %-10s | %s\n", s.Version, s.Name, s.Status, s.AppliedAt)
	}
	fmt.Println("--------------------------------------------------------------------------------------------------------")

	// Print current version
	currentVersion, err := migration.GetCurrentVersion(db)
	if err != nil {
		fmt.Printf("Error getting current version: %s\n", err)
		os.Exit(1)
	}

	if currentVersion == "" {
		fmt.Println("Current version: none")
	} else {
		fmt.Printf("Current version: %s\n", currentVersion)
	}
}
