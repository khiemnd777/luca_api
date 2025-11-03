// scripts/create_module/templates/ent_generate.go.tmpl
// cli: go run modules/auditlog/ent/generate.go
package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func main() {
	err := entc.Generate(utils.GetModulePath("auditlog", "ent", "schema"), &gen.Config{
		Target:  utils.GetModulePath("auditlog", "ent", "generated"),
		Package: "github.com/khiemnd777/andy_api/modules/auditlog/ent/generated",
	},
	)
	if err != nil {
		log.Fatalf("Ent code generation failed: %v", err)
	}
}
