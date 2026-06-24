package browser

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// isURL checks if a string is an absolute URL (http://, https://, or file://)
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "file://")
}

// IsURL checks if a string is an absolute URL (exported for testing)
func IsURL(s string) bool {
	return isURL(s)
}

// isRelativeURL checks if a URL string is relative (not absolute)
func isRelativeURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return !parsedURL.IsAbs()
}

// IsRelativeURL checks if a URL string is relative (exported for testing)
func IsRelativeURL(rawURL string) bool {
	return isRelativeURL(rawURL)
}

// fixURL normalizes a URL by adding https:// prefix and /index.risor suffix as needed
func fixURL(rawURL string) (string, error) {
	// Handle file:// URLs
	if strings.HasPrefix(rawURL, "file://") {
		// Parse the file URL
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return "", fmt.Errorf("invalid file URL: %w", err)
		}

		// Add index.risor if no .risor extension
		if !strings.HasSuffix(parsed.Path, ".risor") {
			parsed.Path = filepath.ToSlash(filepath.Join(parsed.Path, "index.risor"))
		}

		return parsed.String(), nil
	}

	// Add https:// prefix if no scheme present
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	// Parse the URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Add index.risor if no .risor extension
	if !strings.HasSuffix(parsed.Path, ".risor") {
		parsed.Path = filepath.ToSlash(filepath.Join(parsed.Path, "index.risor"))
	}

	return parsed.String(), nil
}

// correctRelativeURL resolves a relative URL against a current URL
func correctRelativeURL(currentUrl, relativeUrl string) string {
	if isRelativeURL(relativeUrl) {
		// For file:// URLs, handle filesystem paths
		if strings.HasPrefix(currentUrl, "file://") {
			currentPath := strings.TrimPrefix(currentUrl, "file://")

			// If currentPath doesn't end with .risor, it's a directory URL
			// In this case, use it as-is rather than taking its parent
			var currentDir string
			if strings.HasSuffix(currentPath, ".risor") {
				currentDir = filepath.Dir(currentPath)
			} else {
				// It's already a directory path, use it directly
				currentDir = currentPath
			}

			resolvedPath := filepath.Join(currentDir, relativeUrl)
			absPath, err := filepath.Abs(resolvedPath)
			if err == nil {
				return "file://" + absPath
			}
			// On error, return original
			return relativeUrl
		}

		// For HTTP(S) URLs, parse properly
		parsedCurrent, err := url.Parse(currentUrl)
		if err != nil {
			return relativeUrl
		}

		// Ensure current URL is treated as a directory (not a file)
		// by adding trailing slash if not present and path doesn't end with .risor
		if !strings.HasSuffix(parsedCurrent.Path, "/") && !strings.HasSuffix(parsedCurrent.Path, ".risor") {
			parsedCurrent.Path += "/"
		}

		parsedRelative, err := url.Parse(relativeUrl)
		if err != nil {
			return relativeUrl
		}

		// Resolve the relative URL against the current URL
		resolved := parsedCurrent.ResolveReference(parsedRelative)
		return resolved.String()
	}
	return relativeUrl
}

// urlJoin joins a base URL with path segments
func urlJoin(baseURL string, paths ...string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	// Join all path elements
	allPaths := append([]string{base.Path}, paths...)
	joined := strings.Join(allPaths, "/")

	// Clean the path (removes double slashes, etc.)
	base.Path = filepath.ToSlash(filepath.Clean(joined))

	return base.String(), nil
}
