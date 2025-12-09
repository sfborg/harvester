package worldplants

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gnames/gn"
	"github.com/gnames/gnsys"
)

// numberedFile represents a CSV file with a numeric name.
type numberedFile struct {
	name   string
	number int
}

// Extract overrides base Extract to handle WFWP-specific file preparation.
// The path parameter can be either:
// - A zip file containing ferns.csv and numbered CSV files
// - A directory containing these files
func (wp *worldplants) Extract(path string) error {
	if path == "" {
		return fmt.Errorf(
			"WFWP requires --file option with path to zip file or directory",
		)
	}

	// Determine the source directory (either from zip or direct)
	sourceDir, cleanup, err := wp.prepareSourceDir(path)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	if err := wp.validateInputDir(sourceDir); err != nil {
		return err
	}

	extractDir := wp.cfg.ExtractDir

	if err := wp.prepareFerns(sourceDir, extractDir); err != nil {
		return err
	}

	if err := wp.preparePlants(sourceDir, extractDir); err != nil {
		return err
	}

	wp.logFileSizes(extractDir)

	return nil
}

// prepareSourceDir extracts zip if needed and returns the source directory.
// Returns the directory path and an optional cleanup function.
func (wp *worldplants) prepareSourceDir(
	path string,
) (string, func(), error) {
	// Check if path is a zip file
	if strings.HasSuffix(strings.ToLower(path), ".zip") {
		return wp.extractZipFile(path)
	}

	// Check if it's a directory
	info, err := os.Stat(path)
	if err != nil {
		return "", nil, fmt.Errorf("cannot access input path: %w", err)
	}

	if !info.IsDir() {
		return "", nil, fmt.Errorf(
			"input path must be a .zip file or directory: %s",
			path,
		)
	}

	// It's a directory, use it directly
	return path, nil, nil
}

// extractZipFile extracts a zip file to a temp directory in the download dir.
// Returns the extracted directory path and a cleanup function.
func (wp *worldplants) extractZipFile(
	zipPath string,
) (string, func(), error) {
	slog.Info("extracting zip file", "path", zipPath)
	gn.Info("Extracting <em>%s</em>", zipPath)

	// Create temp directory in download dir for extraction
	tempDir := filepath.Join(wp.cfg.DownloadDir, "wfwp-extracted")

	// Clean up any existing temp directory
	if err := os.RemoveAll(tempDir); err != nil {
		return "", nil, fmt.Errorf("cannot clean temp directory: %w", err)
	}

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", nil, fmt.Errorf("cannot create temp directory: %w", err)
	}

	// Extract zip file
	if err := gnsys.ExtractZip(zipPath, tempDir); err != nil {
		return "", nil, fmt.Errorf("cannot extract zip file: %w", err)
	}

	slog.Info("zip file extracted", "to", tempDir)
	gn.Info("Zip extracted to <em>%s</em>", tempDir)

	// Cleanup function to remove temp directory
	cleanup := func() {
		if err := os.RemoveAll(tempDir); err != nil {
			slog.Warn("failed to clean up temp directory", "error", err)
			gn.Warn("Unable to clean temp directory")
		}
	}

	return tempDir, cleanup, nil
}

// validateInputDir checks that the input directory contains required files.
func (wp *worldplants) validateInputDir(inputDir string) error {
	if err := wp.checkDirExists(inputDir); err != nil {
		return err
	}

	if err := wp.checkFernsExists(inputDir); err != nil {
		return err
	}

	if err := wp.checkNumberedFilesExist(inputDir); err != nil {
		return err
	}

	return nil
}

// checkDirExists verifies the input directory exists and is readable.
func (wp *worldplants) checkDirExists(inputDir string) error {
	info, err := os.Stat(inputDir)
	if err != nil {
		return fmt.Errorf("cannot access input directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("input path is not a directory: %s", inputDir)
	}
	return nil
}

// checkFernsExists verifies ferns.csv is present.
func (wp *worldplants) checkFernsExists(inputDir string) error {
	fernsPath := filepath.Join(inputDir, "ferns.csv")
	if _, err := os.Stat(fernsPath); err != nil {
		return fmt.Errorf("ferns.csv not found in input directory: %w", err)
	}
	return nil
}

// checkNumberedFilesExist verifies at least one numbered CSV file exists.
func (wp *worldplants) checkNumberedFilesExist(inputDir string) error {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("cannot read input directory: %w", err)
	}

	csvRegexp := regexp.MustCompile(`^\d+\.csv$`)
	for _, file := range files {
		if csvRegexp.MatchString(file.Name()) {
			return nil
		}
	}

	return fmt.Errorf(
		"no numbered CSV files (1.csv, 2.csv, etc.) found in: %s",
		inputDir,
	)
}

// prepareFerns copies ferns.csv to the extract directory.
func (wp *worldplants) prepareFerns(inputDir, extractDir string) error {
	fernsInput := filepath.Join(inputDir, "ferns.csv")
	fernsOutput := filepath.Join(extractDir, "ferns.csv")

	slog.Info("copying ferns.csv", "from", fernsInput, "to", fernsOutput)
	gn.Info("Copy ferns.csv from %s to %s", fernsInput, fernsOutput)

	data, err := os.ReadFile(fernsInput)
	if err != nil {
		return fmt.Errorf("cannot read ferns.csv: %w", err)
	}

	err = os.WriteFile(fernsOutput, data, 0644)
	if err != nil {
		return fmt.Errorf("cannot write ferns.csv: %w", err)
	}

	return nil
}

// preparePlants concatenates numbered CSV files into plants.csv.
func (wp *worldplants) preparePlants(inputDir, extractDir string) error {
	outputPath := filepath.Join(extractDir, "plants.csv")

	slog.Info(
		"concatenating plant CSV files",
		"input", inputDir,
		"output", outputPath,
	)
	gn.Info("Concatenating plant CSV files to %s", outputPath)

	if err := wp.removeExistingPlants(outputPath); err != nil {
		return err
	}

	files, err := wp.findNumberedFiles(inputDir)
	if err != nil {
		return err
	}

	if err := wp.concatenateFiles(inputDir, outputPath, files); err != nil {
		return err
	}

	slog.Info(
		"concatenated plant files",
		"count", len(files),
		"output", outputPath,
	)
	gn.Info("Concatenated plant files into <em>%s</em>", outputPath)

	return nil
}

// removeExistingPlants removes plants.csv if it already exists.
func (wp *worldplants) removeExistingPlants(outputPath string) error {
	if _, err := os.Stat(outputPath); err == nil {
		if err := os.Remove(outputPath); err != nil {
			return fmt.Errorf(
				"cannot remove existing plants.csv: %w",
				err,
			)
		}
	}
	return nil
}

// findNumberedFiles finds and sorts all numbered CSV files.
func (wp *worldplants) findNumberedFiles(
	inputDir string,
) ([]numberedFile, error) {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read input directory: %w", err)
	}

	csvRegexp := regexp.MustCompile(`^(\d+)\.csv$`)
	var numbered []numberedFile

	for _, file := range files {
		matches := csvRegexp.FindStringSubmatch(file.Name())
		if matches == nil {
			continue
		}

		num, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		numbered = append(numbered, numberedFile{
			name:   file.Name(),
			number: num,
		})
	}

	if len(numbered) == 0 {
		return nil, fmt.Errorf(
			"no numbered CSV files found in %s",
			inputDir,
		)
	}

	// Sort numerically (not alphabetically)
	sort.Slice(numbered, func(i, j int) bool {
		return numbered[i].number < numbered[j].number
	})

	return numbered, nil
}

// concatenateFiles writes all numbered files to a single output file.
func (wp *worldplants) concatenateFiles(
	inputDir string,
	outputPath string,
	files []numberedFile,
) error {
	outFile, err := os.OpenFile(
		outputPath,
		os.O_APPEND|os.O_WRONLY|os.O_CREATE,
		0644,
	)
	if err != nil {
		return fmt.Errorf("cannot create plants.csv: %w", err)
	}
	defer outFile.Close()

	for i, file := range files {
		if err := wp.appendFile(inputDir, outFile, file, i, len(files)); err != nil {
			return err
		}
	}

	return nil
}

// appendFile appends a single file's content to the output.
func (wp *worldplants) appendFile(
	inputDir string,
	outFile *os.File,
	file numberedFile,
	index int,
	total int,
) error {
	if (index+1)%10 == 0 || index == 0 {
		slog.Info(
			"processing plant file",
			"file", file.name,
			"progress", fmt.Sprintf("%d/%d", index+1, total),
		)
	}

	inputPath := filepath.Join(inputDir, file.name)
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", file.name, err)
	}

	// First file keeps header, subsequent files skip it
	isFirst := (index == 0)
	content := wp.processFileContent(data, isFirst, file.name)
	if content == nil {
		return nil // File was skipped (logged warning)
	}

	_, err = outFile.Write(content)
	if err != nil {
		return fmt.Errorf("cannot write to plants.csv: %w", err)
	}

	return nil
}

// processFileContent removes header from non-first files.
func (wp *worldplants) processFileContent(
	data []byte,
	isFirst bool,
	fileName string,
) []byte {
	if isFirst {
		return data
	}

	// Skip header line for subsequent files
	newlineIdx := bytes.IndexByte(data, '\n')
	if newlineIdx == -1 {
		slog.Warn("file has no newline, skipping", "file", fileName)
		return nil
	}

	return data[newlineIdx+1:]
}

// logFileSizes logs the sizes of prepared files for verification.
func (wp *worldplants) logFileSizes(extractDir string) {
	plantsPath := filepath.Join(extractDir, "plants.csv")
	fernsPath := filepath.Join(extractDir, "ferns.csv")

	if info, err := os.Stat(plantsPath); err == nil {
		sizeMB := float64(info.Size()) / 1024 / 1024
		slog.Info(
			"created plants.csv",
			"path", plantsPath,
			"size_mb", fmt.Sprintf("%.2f", sizeMB),
		)
		gn.Info("Created <em>%s</em>", plantsPath)
	}

	if info, err := os.Stat(fernsPath); err == nil {
		sizeMB := float64(info.Size()) / 1024 / 1024
		slog.Info(
			"copied ferns.csv",
			"path", fernsPath,
			"size_mb", fmt.Sprintf("%.2f", sizeMB),
		)
		gn.Info("Copied <em>%s</em>", fernsPath)
	}
}
