package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	cfg, err := LoadOrCreateConfig()
	if err != nil {
		fmt.Println("\n[!] Error loading the configuration:", err)
		os.Exit(1)
	}

	path, err := selectPasswordFile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var outputFilePath = OUTPUT_DIR + "/filteredPasswords.txt"

	if _, err := os.Stat(outputFilePath); !os.IsNotExist(err) {
		// Ask the user if they want to overwrite the existing file
		for {
			var overwriteChoice string
			fmt.Println("\nThe file 'filteredPasswords.txt' already exists. Do you want to overwrite it? (yes/no)")
			fmt.Scanln(&overwriteChoice)
			if overwriteChoice == "yes" || overwriteChoice == "y" {

				if err := os.Remove(outputFilePath); err != nil {
					fmt.Println("\n[!] Error deleting the output file:", err)
					os.Exit(1)
				}
				break
			} else if overwriteChoice == "no" || overwriteChoice == "n" {
				// Generate a new filename based on the current date and time
				currentTime := time.Now()
				formattedTime := currentTime.Format("0102-150405")
				outputFilePath = "./output/filteredPassword_" + formattedTime + ".txt"
				break
			} else {
				fmt.Println("Invalid choice. Please enter 'yes' or 'no'.")
			}
		}
	}

	if path == "MERGE_ALL" {
		fmt.Println("\n[*] Please wait, merging dictionaries and removing duplicates...")
		path, err = mergeDictionaries()
		if err != nil {
			fmt.Println("\n[!] Error merging the dictionaries:", err)
			os.Exit(1)
		}
	}

	startTime := time.Now()
	total, valid, err := parsePasswordFile(path, cfg, outputFilePath)
	if err != nil {
		fmt.Println("\n[!] Error processing the password dictionary:", err)
		os.Exit(1)
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	fmt.Printf("\nProcessing statistics:")
	fmt.Printf("\n- Total number of passwords processed: %s", formatNumberWithCommas(total))
	fmt.Printf("\n- Number of passwords retained: %s (%.2f%%)", formatNumberWithCommas(valid), float64(valid)*100/float64(total))
	fmt.Printf("\n- Processing time: %.1f seconds\n", duration.Seconds())
	fmt.Printf("\nFile generated: %v\n", outputFilePath)
	os.Exit(0)
}