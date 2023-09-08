package main

import (
	"strconv"
	"sync"
	"bufio"
	"strings"
	"fmt"
	"io/ioutil"
	"os"
	"io"
	"unicode"
)

func selectPasswordFile() (string, error) {
    // Check if the 'dict' directory exists
    if _, err := os.Stat("./dict"); os.IsNotExist(err) {
        return "", fmt.Errorf("\n[!] The 'dict' directory does not exist.")
    }

	// List all dictionaries found
	files, err := ioutil.ReadDir("./dict")
	if err != nil {
		return "", fmt.Errorf("\n[!] Error reading the 'dict' directory: %v", err)
	}

	for {
		fmt.Println("\nPlease choose a password dictionary from the following options :")
		fmt.Println("0. merge all dictionaries below")
		for i, file := range files {
			fmt.Printf("%d. %s\n", i+1, file.Name())
		}

		var choice int
		fmt.Print("\nEnter the number of your choice: ")
		fmt.Scan(&choice)
		
		if choice == 0 {
			return "MERGE_ALL", nil
		} else if choice > 0 && choice <= len(files) {
			return "./dict/" + files[choice-1].Name(), nil
		} else {
			fmt.Println("\nInvalid choice. Please try again.")
		}
	}
}



func parsePasswordFile(path string, cfg *Config,outputFilePath string ) (int, int, error) {
	batchSize := 50000
	fmt.Println("Processing started, parsing ",path)
	
	totalPasswords := 0
	validPasswords := 0

	file, err := os.Open(path)
	if err != nil {
		return totalPasswords, validPasswords, fmt.Errorf("\n[!] Error opening the file: %v", err)
		
	}
	defer file.Close()

	var batch []string
	const bufferSize = 65536  // 64 KB
	
	totalLines := 0
	buf := make([]byte, bufferSize)
	for {
		bytesRead, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return totalPasswords, validPasswords, fmt.Errorf("\n[!] Error parsing file")
		}
		if bytesRead == 0 {
			break
		}
		for _, b := range buf[:bytesRead] {
			if b == '\n' {
				totalLines++
			}
		}
	}
	
	fmt.Println("Total passwords to process: ", formatNumberWithCommas(totalLines))

	file.Seek(0, 0)  // Reset file pointer to the start
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		password := scanner.Text()
		totalPasswords++
		if isValidPassword(password, cfg) {
			batch = append(batch, password)
			validPasswords++
		}

		// If cache is full write
		if len(batch) >= batchSize {
			err = saveValidPasswords(batch,outputFilePath)
			if err != nil {
				return totalPasswords, validPasswords, err
			}
			batch = batch[:0] 
		}
	}

	if len(batch) > 0 {
		err = saveValidPasswords(batch,outputFilePath)		
		if err != nil {
			return totalPasswords, validPasswords, err
		}
	}

	if err := scanner.Err(); err != nil {
		return totalPasswords, validPasswords, fmt.Errorf("\n[!] Error reading the file: %v", err)
	}

	return totalPasswords, validPasswords, nil
}

func isValidPassword(password string, cfg *Config) bool {
	uppercaseCount := 0
	specialCharCount := 0
	digitCount := 0

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			uppercaseCount++
		case unicode.IsDigit(char):
			digitCount++
		case contains(cfg.SpecialChars, char):
			specialCharCount++
		}
	}

	if len(password) < cfg.MinLength {
		return false
	}
	if uppercaseCount < cfg.MinUppercase {
		return false
	}
	if specialCharCount < cfg.MinSpecialChars {
		return false
	}
	if digitCount < cfg.MinDigits {
		return false
	}
	return true
}

func contains(slice []rune, item rune) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func saveValidPasswords(passwords []string,outputFilePath string) error {
	// Ensure the 'output' directory exists
	if _, err := os.Stat("output"); os.IsNotExist(err) {
		os.Mkdir("output", 0755)
	}
	file, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("\n[!] Error opening the output file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, password := range passwords {
		_, err := writer.WriteString(password + "\n")
		if err != nil {
			return fmt.Errorf("\n[!] Error writing to the output file: %v", err)
		}
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("\n[!] Error flushing the write buffer: %v", err)
	}
	
	return nil
}

func mergeDictionaries() (string, error) {
    files, err := ioutil.ReadDir("./dict")
    if err != nil {
        return "", fmt.Errorf("\n[!] Error reading the 'dict' directory: %v", err)
    }

    mergedPasswords := make(map[string]bool)
    var mu sync.Mutex
    var wg sync.WaitGroup

    for _, file := range files {
        wg.Add(1)
        go func(filename string) {
            defer wg.Done()
            path := "./dict/" + filename
            f, err := os.Open(path)
            if err != nil {
                fmt.Println(fmt.Errorf("\n[!] Error reading the dictionary file %s: %v", path, err))
                return
            }
            defer f.Close()
            scanner := bufio.NewScanner(f)
            for scanner.Scan() {
                password := scanner.Text()
                mu.Lock()
                mergedPasswords[password] = true
                mu.Unlock()
            }
        }(file.Name())
    }

    wg.Wait()

    mergedFilePath := "./output/dicts_merged.txt"
    f, err := os.Create(mergedFilePath)
    if err != nil {
        return "", fmt.Errorf("\n[!] Error creating the merged file: %v", err)
    }
    defer f.Close()

    buffer := make([]string, 0, len(mergedPasswords))
    for password := range mergedPasswords {
        buffer = append(buffer, password)
    }
    _, err = f.WriteString(strings.Join(buffer, "\n"))
    if err != nil {
        return "", fmt.Errorf("\n[!] Error writing to the merged file: %v", err)
    }

    return mergedFilePath, nil
}

func formatNumberWithCommas(n int) string {
    in := strconv.Itoa(n)
    var out []rune
    count := 0

    for i := len(in) - 1; i >= 0; i-- {
        if count > 0 && count%3 == 0 {
            out = append([]rune{','}, out...)
        }
        out = append([]rune{rune(in[i])}, out...)
        count++
    }
    return string(out)
}