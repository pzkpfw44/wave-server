package main

import (
	"flag"
	"fmt"
	"os"

	migration_tool "github.com/pzkpfw44/wave-server/scripts"
)

func main() {
	sourceDir := flag.String("source", "./old_data", "Source directory for old data")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	isDev := flag.Bool("dev", true, "Development mode")

	flag.Parse()

	if err := migration_tool.RunMigrationTool(*sourceDir, *logLevel, *isDev); err != nil {
		fmt.Printf("Migration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Migration completed successfully")
}
