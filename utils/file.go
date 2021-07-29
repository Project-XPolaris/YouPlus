package utils

import (
	"io"
	"log"
	"os"
)

func WriteLineToFile(path string, line string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	if _, err = f.WriteString(line); err != nil {
		log.Fatal(err)
	}
	if err = f.Close(); err != nil {
		log.Fatal(err)
	}
	return err
}

func WriteLinesToFile(path string, lines []string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	for _, line := range lines {
		_, err = io.WriteString(f, line+"\n")
		if err != nil {
			return err
		}
	}
	return err
}

func IsFileExist(target string) bool {
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return false
	}
	return true
}
