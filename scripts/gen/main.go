package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/gen"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	config.Init(utils.GetFullPath("config.yaml"))

	switch os.Args[1] {
	case "ent":
		if len(os.Args) < 3 {
			fmt.Println("❌ Missing schema name. Example: go run scripts/gen.go ent User")
			return
		}
		generateEntSchema(os.Args[2])
	case "generate":
		gen.GenerateEntClient()
	case "seed":
		runSeeder()
	case "migrate":
		gen.RunFlyway("migrate")
	case "drop":
		gen.RunFlyway("clean")
	case "reset":
		gen.RunFlyway("clean")
		gen.RunFlyway("migrate")
	case "version":
		gen.RunFlyway("info")
	case "conf":
		if err := gen.GenerateFlywayConfig(); err != nil {
			return
		}
	default:
		fmt.Printf("❌ Unknown command: %s\n", os.Args[1])
		printHelp()
	}
}

func printHelp() {
	fmt.Println("\n📘 Dev CLI Helper Tool")
	fmt.Println("Usage:")
	fmt.Println("  go run ./scripts/gen ent <SchemaName>    📦 Create new schema and generate Ent client")
	fmt.Println("  go run ./scripts/gen generate             ⚙️  Only re-generate Ent client")
	fmt.Println("  go run ./scripts/gen seed                 🌱 Run seed logic")
	fmt.Println("  go run ./scripts/gen conf                 🛠 Generate flyway.conf from config.yaml")
	fmt.Println("  go run ./scripts/gen migrate              🚀 Run Flyway migrations")
	fmt.Println("  go run ./scripts/gen drop                 🧨 Drop all DB schema (clean)")
	fmt.Println("  go run ./scripts/gen reset                🔁 Drop & re-run migrations")
	fmt.Println("  go run ./scripts/gen version              🧾 Show migration info")
	fmt.Println()
}

func generateEntSchema(schema string) {
	fmt.Printf("📦 Creating schema: %s\n", schema)

	targetDir := filepath.Join(".", "shared", "db", "ent", "schema")

	cmd := exec.Command("ent", "new", schema, "--target", targetDir, "--feature", "sql/execquery")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to create schema: %v\n", err)
		return
	}

	gen.GenerateEntClient()
}

func runSeeder() {
	fmt.Println("🌱 Running seed logic (TODO: implement your seeder here)...")
}
