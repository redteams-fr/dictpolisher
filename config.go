package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"
)

// Config représente la structure de la configuration de filtrage des mots de passe.
type Config struct {
	MinLength       int    `json:"min_length"`
	MinUppercase    int    `json:"min_uppercase"`
	MinSpecialChars int    `json:"min_special_chars"`
	SpecialChars    []rune `json:"special_chars"`
	MinDigits       int    `json:"min_digits"`
}


func displayPolicy(cfg *Config) {
	fmt.Printf("   - Minimum password length: %d\n", cfg.MinLength)
	fmt.Printf("   - Minimum number of uppercase characters: %d\n", cfg.MinUppercase)
	fmt.Printf("   - Minimum number of special characters: %d\n", cfg.MinSpecialChars)
	fmt.Printf("   - Allowed special characters: %s\n", string(cfg.SpecialChars))
	fmt.Printf("   - Minimum number of digits: %d\n", cfg.MinDigits)
	fmt.Println()
}
func LoadOrCreateConfig() (*Config, error) {
	// Looking for config file
	files, err := filepath.Glob("config_*.json")
	if err != nil {
		return nil, err
	}

	if len(files) > 0 {
		for {
			fmt.Println("\nConfiguration(s) found:\n")
			for i, file := range files {
				// Load the config to display its policy
				cfg, err := loadConfig(file)
				if err == nil {
					fmt.Printf("%d. %s \n", i+1, file)
					displayPolicy(cfg)
				}
			}
			fmt.Println("Choose a configuration or press 'n' to create a new one.")
			var choice string
			fmt.Scanln(&choice)
			if choice == "n" {
				fmt.Println("\nDefine new policy")
				return createConfig()
			}
			fileChoice, err := strconv.Atoi(choice)
			if err != nil || fileChoice < 1 || fileChoice > len(files) {
				fmt.Println("\n[!] Error : invalid choice. Please select again.")
				continue
			}
			return loadConfig(files[fileChoice-1])
		}
	}
	fmt.Println("\nDefine new policy")

	return createConfig()
}


// loadConfig charge un fichier de configuration existant.
func loadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}


func createConfig() (*Config, error) {
	var cfg Config
	fmt.Print("👉 Minimum password length: ")
	fmt.Scanln(&cfg.MinLength)
	fmt.Print("👉 Minimum number of uppercase characters: ")
	fmt.Scanln(&cfg.MinUppercase)
	fmt.Print("👉 Minimum number of special characters: ")
	fmt.Scanln(&cfg.MinSpecialChars)
	
	// Only ask for special characters if the user specified a non-zero minimum
	if cfg.MinSpecialChars > 0 {
		fmt.Println("👉 Usable special characters (default: !@#$%^&*()_-): ")
		var specialChars string
		fmt.Scanln(&specialChars)
		if specialChars == "" {
			specialChars = "!@#$%^&*()_-"
		}
		cfg.SpecialChars = []rune(specialChars)
	} else {
		// If the user specified 0 for minimum special characters, make sure the list is empty
		cfg.SpecialChars = []rune{}
	}
	
	fmt.Print("👉 Minimum number of digits: ")
	fmt.Scanln(&cfg.MinDigits)

	filename := "config_" + time.Now().Format("20060102-150405") + ".json"
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\n[*] Configuration saved in file %s\n", filename)
	return &cfg, nil
}

