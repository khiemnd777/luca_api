package gen

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/khiemnd777/andy_api/shared/flyway"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func GenerateEntClient() error {
	log.Println("⚙️  Generating Ent client...")

	entPath := utils.GetFullPath("shared", "db", "ent")

	cmd := exec.Command("go", "generate", entPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = utils.GetProjectRootDir()

	if err := cmd.Run(); err != nil {
		log.Fatalf("❌ Failed to generate Ent client: %v", err)
		return err
	}

	log.Println("✅ Ent Client generated successfully.")
	return nil
}

func RunFlyway(action string) error {
	var stdout, stderr bytes.Buffer

	log.Printf("🚀 Running flyway %s...", action)

	confPath := utils.GetFullPath("flyway", "flyway.conf")

	cmd := exec.Command("flyway", action, "-configFiles="+confPath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		combined := stdout.String() + stderr.String()
		// log.Fatal("❌ Failed to run flyway " + action + " error: " + combined)
		fmt.Println(combined)

		if strings.Contains(combined, "flyway_schema_history") {
			fmt.Println("⚠️ flyway_schema_history missing. Running baseline...")

			if baseErr := RunFlyway("baseline"); baseErr != nil {
				return fmt.Errorf("flyway baseline failed: %w", baseErr)
			}

			log.Println("✅ Flyway baseline completed. Retrying migrate...")
			return RunFlyway("migrate")
		}

		return err
	}

	log.Printf("✅ Flyway %s completed.\n", action)
	return nil
}

func GenerateFlywayConfig() error {
	if err := flyway.GenerateFlywayConf(); err != nil {
		log.Fatalf("❌ Failed to generate flyway.conf: %v\n", err)
		return err
	}
	log.Println("✅ flyway.conf generated successfully.")
	return nil
}
