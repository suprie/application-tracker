package main

import (
	"fmt"
	"os"

	_ "modernc.org/sqlite"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"suprie/application_tracker/internal/service"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Usage: ats <parse-jd|parse-cv> <file.pdf>")
		os.Exit(1)
	}

	command := os.Args[1] filename := os.Args[2]

	switch command { case "parse-cv": service.RunParseCV(filename) case
	"parse-jd":
		service.RunParseJD(filename)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Usage: ats <parse-cv|parse-jd> <file>\n")
		os.Exit(1)
	}

}
