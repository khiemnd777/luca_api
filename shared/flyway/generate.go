package flyway

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/khiemnd777/andy_api/shared/config"
)

type FlywayConfTemplateData struct {
	URL      string
	User     string
	Password string
	Schemas  string
}

func GenerateFlywayConf() error {
	log.Println("🔧 Generating Flyway config...")

	cfg := config.Get().Database
	url, user, pass, schema, err := generateJDBCUrl(cfg.Provider, cfg)
	if err != nil {
		return err
	}

	data := FlywayConfTemplateData{
		URL:      url,
		User:     user,
		Password: pass,
		Schemas:  schema,
	}

	tmpl := template.Must(template.New("conf").Parse(`flyway.url={{.URL}}
flyway.user={{.User}}
flyway.password={{.Password}}
flyway.schemas={{.Schemas}}
flyway.locations=filesystem:{{.SQLPath}}
flyway.cleanDisabled=false
`))

	flywayDir := filepath.Join("flyway")
	sqlDir := filepath.Join(flywayDir, "sql")
	confPath := filepath.Join(flywayDir, "flyway.conf")

	dataWithPath := struct {
		FlywayConfTemplateData
		SQLPath string
	}{
		FlywayConfTemplateData: data,
		SQLPath:                sqlDir,
	}

	_ = os.MkdirAll(flywayDir, os.ModePerm)

	f, err := os.Create(confPath)
	if err != nil {
		return fmt.Errorf("create flyway.conf failed: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, dataWithPath); err != nil {
		return fmt.Errorf("write flyway.conf failed: %w", err)
	}

	log.Println("✅ Generated flyway.conf for", cfg.Provider)
	return nil
}

func generateJDBCUrl(provider string, cfg config.DatabaseConfig) (string, string, string, string, error) {
	switch provider {
	case "postgres":
		pg := cfg.Postgres
		url := fmt.Sprintf("jdbc:postgresql://%s:%d/%s", pg.Host, pg.Port, pg.Name)
		return url, pg.User, pg.Password, "public", nil
	default:
		return "", "", "", "", fmt.Errorf("unsupported database provider: %s", provider)
	}
}
