package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"circles.diy/internal/config"
	"go.uber.org/zap"
)

// substituteCSSVariables replaces CSS variables with their actual values
// in media query contexts only, leaving other CSS custom properties intact
func substituteCSSVariables(cssContent string) (string, error) {
	// Get the breakpoint variable mappings
	variables := config.GetVariableMap()

	// Regex to match @media rules that contain CSS variables
	// This matches: @media (...var(--breakpoint-name)...)
	mediaQueryRegex := regexp.MustCompile(`(@media[^{]*var\(--breakpoint-[^}]*\)[^{]*\{)`)

	// Find all media query blocks that contain CSS variables
	result := mediaQueryRegex.ReplaceAllStringFunc(cssContent, func(mediaQuery string) string {
		// Replace each CSS variable in this media query
		processedQuery := mediaQuery
		for varName, value := range variables {
			// Create a regex for this specific variable: var(--breakpoint-name)
			varRegex := regexp.MustCompile(`var\(` + regexp.QuoteMeta(varName) + `\)`)
			processedQuery = varRegex.ReplaceAllString(processedQuery, value)
		}
		return processedQuery
	})

	return result, nil
}

func BuildCSS(logger *zap.Logger) error {
	// ITCSS Layer directories in correct specificity order
	cssDirectories := []string{
		"static/css/01-settings",
		"static/css/02-tools",
		"static/css/03-generic",
		"static/css/04-elements",
		"static/css/05-objects",
		"static/css/06-components",
		"static/css/07-utilities",
	}

	// Create output file
	outputFile := "static/css/style.css"
	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output CSS file: %v", err)
	}
	defer output.Close()

	// Write header comment
	if _, err := output.WriteString("/* Generated CSS - circles.diy */\n\n"); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	// Process each ITCSS layer directory
	for _, cssDir := range cssDirectories {
		if _, err := os.Stat(cssDir); os.IsNotExist(err) {
			continue
		}

		// Get all CSS files in the directory
		files, err := filepath.Glob(filepath.Join(cssDir, "*.css"))
		if err != nil {
			return fmt.Errorf("failed to glob CSS files in %s: %v", cssDir, err)
		}

		if len(files) == 0 {
			continue
		}

		// Write layer comment
		layerName := filepath.Base(cssDir)
		if _, err := output.WriteString(fmt.Sprintf("/* === ITCSS Layer: %s === */\n", layerName)); err != nil {
			return fmt.Errorf("failed to write layer comment: %v", err)
		}

		// Process each CSS file in the directory
		for _, cssFile := range files {
			file, err := os.Open(cssFile)
			if err != nil {
				return fmt.Errorf("failed to open CSS file %s: %v", cssFile, err)
			}

			// Write file marker comment
			fileName := filepath.Base(cssFile)
			if _, err := output.WriteString(fmt.Sprintf("/* --- %s --- */\n", fileName)); err != nil {
				file.Close()
				return fmt.Errorf("failed to write file marker: %v", err)
			}

			// Read file content
			content, err := io.ReadAll(file)
			if err != nil {
				file.Close()
				return fmt.Errorf("failed to read CSS file %s: %v", cssFile, err)
			}

			// Process CSS variables in media queries
			processedContent, err := substituteCSSVariables(string(content))
			if err != nil {
				file.Close()
				return fmt.Errorf("failed to process CSS variables in %s: %v", cssFile, err)
			}

			// Write processed content
			if _, err := output.WriteString(processedContent); err != nil {
				file.Close()
				return fmt.Errorf("failed to write processed CSS file %s: %v", cssFile, err)
			}

			// Add separator
			if _, err := output.WriteString("\n\n"); err != nil {
				file.Close()
				return fmt.Errorf("failed to write separator: %v", err)
			}

			file.Close()
		}
	}

	logger.Debug("CSS compiled successfully", zap.String("output", outputFile))
	return nil
}

func WatchCSSFiles(logger *zap.Logger) {
	cssDir := "static/css"
	lastModTime := time.Time{}

	for {
		hasChanges := false

		// Check if any CSS files have changed
		err := filepath.Walk(cssDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Ext(path) == ".css" && path != "static/css/style.css" {
				if info.ModTime().After(lastModTime) {
					hasChanges = true
					lastModTime = info.ModTime()
				}
			}
			return nil
		})

		if err != nil {
			logger.Warn("error watching CSS files", zap.Error(err))
		} else if hasChanges {
			logger.Debug("CSS files changed, rebuilding...")
			if err := BuildCSS(logger); err != nil {
				logger.Error("error rebuilding CSS", zap.Error(err))
			}
		}

		time.Sleep(1 * time.Second)
	}
}
