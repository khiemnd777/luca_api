// scripts/create_module/main.go
// cli: go run .\scripts\create_module\ --name={module_name} --server-port={server_port} --db-name={database_name}
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {
	name := flag.String("name", "", "Module name (e.g., order)")
	serverHost := flag.String("server-host", "127.0.0.1", "Server host")
	serverPort := flag.Int("server-port", 8081, "Server port")
	dbHost := flag.String("db-host", "127.0.0.1", "Database host")
	dbPort := flag.Int("db-port", 5432, "Database port")
	dbUser := flag.String("db-user", "postgres", "Database user")
	dbPass := flag.String("db-password", "postgres", "Database password")
	dbName := flag.String("db-name", "", "Database name (default: honvang_<module>)")
	dbSSL := flag.String("db-sslmode", "disable", "Database SSL mode")
	withEnt := flag.Bool("with-ent", false, "Whether to include Ent in this module")

	flag.Parse()

	if *name == "" {
		fmt.Println("❌ Module name is required. Usage: --name=order")
		os.Exit(1)
	}

	moduleName := strings.ToLower(*name)
	titleCase := cases.Title(language.English).String(moduleName)
	modulePath := filepath.Join("modules", moduleName)
	templatePath := filepath.Join("scripts", "create_module", "templates")

	dirs := []string{
		filepath.Join(modulePath, "config"),
		filepath.Join(modulePath, "handler"),
		filepath.Join(modulePath, "repository"),
		filepath.Join(modulePath, "service"),
		filepath.Join(modulePath, "txmanager"),
	}
	if *withEnt {
		dirs = append(dirs, filepath.Join(modulePath, "ent", "schema"))
		dirs = append(dirs, filepath.Join(modulePath, "ent", "bootstrap"))
		dirs = append(dirs, filepath.Join(modulePath, "ent", "cmd"))
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(err)
		}
	}

	actualDBName := *dbName
	if actualDBName == "" {
		actualDBName = "andy_" + moduleName
	}

	templates := map[string]string{
		filepath.Join("main.go"):                   "main.go.tmpl",
		filepath.Join("config.yaml"):               "config.yaml.tmpl",
		filepath.Join("config", "config.go"):       "config_config.go.tmpl",
		filepath.Join("handler", "http.go"):        "handler_http.go.tmpl",
		filepath.Join("repository", "repo.go"):     "repository_repo.go.tmpl",
		filepath.Join("service", "service.go"):     "service_service.go.tmpl",
		filepath.Join("txmanager", "txmanager.go"): "txmanager_txmanager.go.tmpl",
	}
	if *withEnt {
		templates[filepath.Join("ent", "generate.go")] = "ent_generate.go.tmpl"
		templates[filepath.Join("ent", "bootstrap", "bootstrap.go")] = "ent_bootstrap_bootstrap.go.tmpl"
		templates[filepath.Join("ent", "schema", "placeholder.go")] = "ent_schema_placeholder.go"
		templates[filepath.Join("ent", "cmd", "migrate.go")] = "ent_cmd_migrate.go.tmpl"
	}

	for relPath, tplFile := range templates {
		fullPath := filepath.Join(modulePath, relPath)
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("⚠️  Skipping existing file: %s\n", fullPath)
			continue
		}

		tpl := loadTemplate(filepath.Join(templatePath, tplFile))
		f, err := os.Create(fullPath)
		if err != nil {
			panic(err)
		}

		t := template.Must(template.New("").Parse(tpl))
		t.Execute(f, map[string]any{
			"Module":      titleCase,
			"module":      moduleName,
			"ServerHost":  *serverHost,
			"ServerPort":  *serverPort,
			"ServerRoute": fmt.Sprintf("/api/%s", moduleName),
			"DBHost":      *dbHost,
			"DBPort":      *dbPort,
			"DBUser":      *dbUser,
			"DBPassword":  *dbPass,
			"DBName":      actualDBName,
			"DBSSLMode":   *dbSSL,
			"WithEnt":     *withEnt,
		})
		f.Close()
	}

	fmt.Printf("✅ Module '%s' created successfully at %s\n", moduleName, modulePath)
}

func loadTemplate(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(data)
}
