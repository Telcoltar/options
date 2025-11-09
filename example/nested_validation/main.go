package main

import (
	"fmt"

	"github.com/telcoltar/options/pkg/option"
)

// DatabaseConfig with its own validation
type DatabaseConfig struct {
	Host     *option.Simple[string]
	Port     *option.Simple[int32]
	MaxConns *option.Simple[int32]

	*option.Container[DatabaseConfig]
}

func NewDatabaseConfig() *DatabaseConfig {
	dc := &DatabaseConfig{
		Host:     option.NewSimple[string]("host").Default("localhost"),
		Port:     option.NewSimple[int32]("port").Default(5432),
		MaxConns: option.NewSimple[int32]("maxConns").Default(100),
	}

	dc.Container = option.NewContainer("database", dc)

	// Validate database-specific rules
	dc.Container.AddCheck(func(d *DatabaseConfig) bool {
		if d.Port.HasValue() {
			port := d.Port.Get()
			if port < 1024 || port > 65535 {
				return false
			}
		}

		if d.MaxConns.HasValue() && d.MaxConns.Get() < 1 {
			return false
		}

		return true
	})

	return dc
}

// AppConfig with nested DatabaseConfig
type AppConfig struct {
	AppName  *option.Simple[string]
	Workers  *option.Simple[int32]
	Debug    *option.Simple[bool]
	Database *DatabaseConfig

	*option.Container[AppConfig]
}

func NewAppConfig() *AppConfig {
	ac := &AppConfig{
		AppName:  option.NewSimple[string]("appName").Default("MyApp"),
		Workers:  option.NewSimple[int32]("workers").Default(4),
		Debug:    option.NewSimple[bool]("debug").Default(false),
		Database: NewDatabaseConfig(),
	}

	ac.Container = option.NewContainer("AppConfig", ac)

	// Validate app-level rules
	ac.Container.AddCheck(func(a *AppConfig) bool {
		if a.Workers.HasValue() {
			workers := a.Workers.Get()
			if workers < 1 || workers > 128 {
				return false
			}
		}

		return true
	})

	return ac
}

func main() {
	fmt.Println("=== Nested Container Validation Example ===")
	fmt.Println()

	// Test 1: Valid nested configuration
	fmt.Println("Test 1: Valid nested configuration")
	config1 := NewAppConfig()
	if !config1.IsValid() {
		fmt.Printf("✗ Validation failed\n\n")
	} else {
		fmt.Printf("✓ Valid configuration\n")
		fmt.Printf("  App: %s, Workers: %d\n", config1.AppName.Get(), config1.Workers.Get())
		fmt.Printf("  DB: %s:%d, MaxConns: %d\n",
			config1.Database.Host.Get(),
			config1.Database.Port.Get(),
			config1.Database.MaxConns.Get())
		fmt.Println()
	}

	// Test 2: Invalid database port
	fmt.Println("Test 2: Invalid database port")
	config2 := NewAppConfig()
	config2.Database.Port.Set(100) // Invalid port
	if !config2.IsValid() {
		fmt.Printf("✓ Validation correctly failed: port=%d is invalid (must be 1024-65535)\n", config2.Database.Port.Get())
		fmt.Println()
	} else {
		fmt.Println("✗ Should have failed validation")
		fmt.Println()
	}

	// Test 3: Invalid workers count
	fmt.Println("Test 3: Invalid workers count")
	config3 := NewAppConfig()
	config3.Workers.Set(200) // Too many workers
	if !config3.IsValid() {
		fmt.Printf("✓ Validation correctly failed: workers=%d is invalid (must be 1-128)\n", config3.Workers.Get())
		fmt.Println()
	} else {
		fmt.Println("✗ Should have failed validation")
		fmt.Println()
	}

	// Test 4: Parse and validate YAML
	fmt.Println("Test 4: Parse and validate YAML")
	config4 := NewAppConfig()
	yamlData := []byte(`
appName: "ProductionApp"
workers: 8
debug: true
database:
  host: "prod-db.example.com"
  port: 5432
  maxConns: 200
`)

	if err := config4.Parse(yamlData); err != nil {
		fmt.Printf("✗ Parse failed: %v\n", err)
		return
	}

	if !config4.IsValid() {
		fmt.Printf("✗ Validation failed\n")
	} else {
		fmt.Printf("✓ Successfully parsed and validated YAML\n")
		fmt.Printf("  App: %s, Workers: %d, Debug: %v\n",
			config4.AppName.Get(),
			config4.Workers.Get(),
			config4.Debug.Get())
		fmt.Printf("  DB: %s:%d, MaxConns: %d\n",
			config4.Database.Host.Get(),
			config4.Database.Port.Get(),
			config4.Database.MaxConns.Get())
	}
}
