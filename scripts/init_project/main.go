package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("🚀 Initializing project...")

	// Step 0: Delete vendor folder if exists
	deleteFolder("vendor")

	// Step 1: Generate Ent for shared
	run("Generating Ent for shared", "go", "run", "-mod=mod", "entgo.io/ent/cmd/ent", "generate", "./shared/db/ent/schema", "--target", "./shared/db/ent/generated")

	// Step 1.1: Init database
	run("Initializing database", "go", "run", "./scripts/init_db")

	// Step 2: Auto generate Ent for all modules
	err := filepath.Walk("modules", func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}

		if filepath.Base(path) == "schema" && filepath.Base(filepath.Dir(path)) == "ent" {
			module := filepath.Dir(filepath.Dir(path))
			schema := "./" + path
			target := "./" + filepath.Join(module, "ent", "generated")
			run(fmt.Sprintf("Generating Ent for %s", module), "go", "run", "-mod=mod", "entgo.io/ent/cmd/ent", "generate", schema, "--target", target)
			migratePath := filepath.Join(module, "ent", "cmd", "migrate.go")
			run(
				fmt.Sprintf("⚙️ Running auto-migrate for %s", module),
				"go", "run", "-mod=mod", migratePath,
			)

		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to walk modules: %v\n", err)
		os.Exit(1)
	}

	// Step 3: go mod tidy & vendor
	run("Running go mod tidy", "go", "mod", "tidy")
	run("Running go mod vendor", "go", "mod", "vendor")

	// // Step 4: Init roles
	// run("Initializing roles", "go", "run", "./scripts/init_roles")

	// Step 5: Build all
	run("Building all modules", "go", "build", "./...")

	fmt.Println("✅ Project initialized successfully!")
}

func run(title string, name string, args ...string) {
	fmt.Println("👉 " + title)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed: %s %v\n", name, args)
		os.Exit(1)
	}
}

func deleteFolder(folder string) {
	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		fmt.Println("🧹 Removing folder:", folder)
		if err := os.RemoveAll(folder); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to delete folder %s: %v\n", folder, err)
			os.Exit(1)
		}
	}
}
