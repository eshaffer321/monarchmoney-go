package graphql

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"sync"
)

//go:embed queries/*
var queriesFS embed.FS

// QueryLoader loads GraphQL queries from embedded files
type QueryLoader struct {
	cache map[string]string
	mu    sync.RWMutex
}

// NewQueryLoader creates a new query loader
func NewQueryLoader() *QueryLoader {
	return &QueryLoader{
		cache: make(map[string]string),
	}
}

// Load loads a query from the embedded filesystem
func (l *QueryLoader) Load(queryPath string) (string, error) {
	// Check cache first
	l.mu.RLock()
	if query, ok := l.cache[queryPath]; ok {
		l.mu.RUnlock()
		return query, nil
	}
	l.mu.RUnlock()

	// Load from embedded FS
	fullPath := path.Join("queries", queryPath)
	content, err := queriesFS.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to load query %s: %w", queryPath, err)
	}

	query := string(content)

	// Cache the query
	l.mu.Lock()
	l.cache[queryPath] = query
	l.mu.Unlock()

	return query, nil
}

// LoadAll loads all queries from a directory
func (l *QueryLoader) LoadAll(dir string) (map[string]string, error) {
	queries := make(map[string]string)

	dirPath := path.Join("queries", dir)
	entries, err := queriesFS.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), ".graphql") {
			queryPath := path.Join(dir, entry.Name())
			query, err := l.Load(queryPath)
			if err != nil {
				return nil, err
			}

			// Use filename without extension as key
			key := strings.TrimSuffix(entry.Name(), ".graphql")
			queries[key] = query
		}
	}

	return queries, nil
}

// MustLoad loads a query and panics on error (for initialization)
func (l *QueryLoader) MustLoad(queryPath string) string {
	query, err := l.Load(queryPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load required query %s: %v", queryPath, err))
	}
	return query
}

// List returns all available queries
func (l *QueryLoader) List() ([]string, error) {
	var queries []string

	err := fs.WalkDir(queriesFS, "queries", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".graphql") {
			// Remove "queries/" prefix
			queryPath := strings.TrimPrefix(path, "queries/")
			queries = append(queries, queryPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list queries: %w", err)
	}

	return queries, nil
}

// Global loader instance
var defaultLoader = NewQueryLoader()

// Load is a convenience function using the default loader
func Load(queryPath string) (string, error) {
	return defaultLoader.Load(queryPath)
}

// MustLoad is a convenience function using the default loader
func MustLoad(queryPath string) string {
	return defaultLoader.MustLoad(queryPath)
}
