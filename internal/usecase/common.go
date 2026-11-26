package usecase

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// copyFile copies file from the source path to the destination path and returns error.
func copyFile(src_file_path, dst_file_path string) error {
	// Open the source file
	sourceFile, err := os.Open(src_file_path)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.Create(dst_file_path)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Efficiently stream the bytes from source to destination
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Flush the file system's in-memory buffers to stable storage
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}
	return nil
}

// safeString replaces characters in a string that are not allowed in file names with safe alternatives.
func safeString(input string) string {
	replacer := strings.NewReplacer(
		"/", "⧸",
		"?", "",
		":", "-",
		"*", "",
	)

	return replacer.Replace(input)

}

// findFiles performs resursive search from root_dir_path down for files matching regular expression.
// Hidden directories and files are skipped
// Returns a string slice of complete file paths
func findFiles(root_dir_path, file_regexp string) ([]string, error) {
	var matches []string

	err := filepath.WalkDir(root_dir_path, func(path string, d fs.DirEntry, err error) error {
		// Check for hidden files and directories (starting with a dot)
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir // Skip the entire directory
			}
			return nil // Skip the hidden file
		}

		// Check if it's a file and matches the target extension. Excludes apple nfo metadata files starting with "._" and other hidden files.
		// if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), file_extension) {
		// if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), file_extension) && !strings.HasPrefix(filepath.Base(path), "._") {
		if !d.IsDir() {
			match, err := regexp.MatchString(file_regexp, path)
			if err != nil {
				return err // Stop walking if a directory cannot be read
			}

			if match {
				matches = append(matches, path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
		return nil, err
	}

	return matches, nil
}

func findVideoFile(root_path, basename string) (string, error) {
	var match string
	video_extensions := []string{".mkv", ".mp4", ".webm"}
	err := filepath.WalkDir(root_path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Stop walking if a directory cannot be read
		}

		// Check if it's a file and matches any of the video file extensions
		if !d.IsDir() && slices.Contains(video_extensions, filepath.Ext(strings.ToLower(path))) && strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) == basename {
			match = path
			return fs.SkipDir
		} else {
			fmt.Printf("%s is not a video file or name does not container '%s'.\n", path, basename)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return match, nil

}
