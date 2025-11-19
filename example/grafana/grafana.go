package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	grafana "github.com/Telcoltar/options/example/grafana/pkg"
)

func main() {
	ghv := grafana.NewGrafanaHelmValues()

	schema := ghv.JSONSchemaWithMetadata()
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("valueSchema.json", schemaJSON, 0o600); err != nil {
		log.Fatal(err)
	}

	yamlPath := filepath.Join("example", "grafana", "test.yaml")
	yamlData, err := os.ReadFile(yamlPath)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	if err := ghv.Parse(yamlData); err != nil {
		log.Fatal(err)
	}

	if !ghv.IsValid() {
		log.Fatal("not valid, exiting")
	}

	log.Println("Build resources")
	resources, err := grafana.Build("central-grafana", ghv)
	if err != nil {
		log.Fatal(err)
	}
	resourceBytes, err := resources.AsYaml()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("resources.yaml", resourceBytes, 0o600); err != nil {
		log.Fatal(err)
	}
}
