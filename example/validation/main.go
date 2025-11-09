package main

import (
	"fmt"
	"log"

	"github.com/telcoltar/options/pkg/option"
)

// ServerConfig demonstrates validation with typed Container access
type ServerConfig struct {
	MinPort *option.Simple[int32]
	MaxPort *option.Simple[int32]
	Host    *option.Simple[string]

	*option.Container[ServerConfig]
}

func NewServerConfig() *ServerConfig {
	sc := &ServerConfig{
		MinPort: option.NewSimple[int32]("minPort").Default(8000),
		MaxPort: option.NewSimple[int32]("maxPort").Default(9000),
		Host:    option.NewSimple[string]("host").Default("localhost"),
	}

	sc.Container = option.NewContainer("ServerConfig", sc)

	// Add validation check with full typed access to the struct!
	sc.Container.AddCheck(func(s *ServerConfig) bool {
		// Check port range validation
		if s.MinPort.HasValue() && s.MaxPort.HasValue() {
			if s.MinPort.Get() >= s.MaxPort.Get() {
				return false
			}
		}

		// Check host is not empty
		if s.Host.HasValue() && s.Host.Get() == "" {
			return false
		}

		return true
	})

	return sc
}

func main() {
	fmt.Println("=== Validation Example ===")
	fmt.Println()

	// Test 1: Valid configuration
	fmt.Println("Test 1: Valid configuration")
	config1 := NewServerConfig()
	if !config1.IsValid() {
		log.Printf("Validation failed")
	} else {
		fmt.Printf("✓ Valid: minPort=%d, maxPort=%d, host=%s\n",
			config1.MinPort.Get(), config1.MaxPort.Get(), config1.Host.Get())
		fmt.Println()
	}

	// Test 2: Invalid configuration (minPort >= maxPort)
	fmt.Println("Test 2: Invalid configuration (minPort >= maxPort)")
	config2 := NewServerConfig()
	config2.MinPort.Set(9000)
	config2.MaxPort.Set(8000)
	if !config2.IsValid() {
		fmt.Printf("✓ Validation correctly failed: minPort=%d >= maxPort=%d\n",
			config2.MinPort.Get(), config2.MaxPort.Get())
		fmt.Println()
	} else {
		fmt.Println("✗ Validation should have failed but didn't")
		fmt.Println()
	}

	// Test 3: Parse YAML and validate
	fmt.Println("Test 3: Parse YAML and validate")
	config3 := NewServerConfig()
	yamlData := []byte(`
minPort: 3000
maxPort: 3500
host: "0.0.0.0"
`)
	if err := config3.Parse(yamlData); err != nil {
		log.Fatalf("Parse failed: %v", err)
	}

	if !config3.IsValid() {
		fmt.Printf("✗ Validation failed\n")
	} else {
		fmt.Printf("✓ Valid after parsing: minPort=%d, maxPort=%d, host=%s\n",
			config3.MinPort.Get(), config3.MaxPort.Get(), config3.Host.Get())
	}
}
