package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"encoding/json"
)

func passwordMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		password := os.Getenv("STARTUPCACHE_PASSWORD")
		if password == "" {
			http.Error(w, "STARTUPCACHE_PASSWORD not set", http.StatusInternalServerError)
			return
		}

		gotPassword := r.Header.Get("Password")
		if gotPassword != password {
			time.Sleep(3 * time.Second)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func rewriteFsPathMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Println("Rewriting path", r.URL.Path)
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/gcgo-startupcache")
		// log.Println("Rewritten path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func StartCacheServer(listenAddr string, startupCacheDir string) {
	log.Println("Starting HTTP server at", listenAddr)
	fs := http.FileServer(http.Dir(startupCacheDir))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/gcgo-startupcache/", rewriteFsPathMiddleware(passwordMiddleware(fs)))
	// List of chunks files for the given chain and table.
	chunksHandler := func(w http.ResponseWriter, r *http.Request) {
		chain := r.PathValue("chain")
		if chain == "" {
			http.Error(w, "No chain provided", http.StatusBadRequest)
			return
		}
		table := r.PathValue("table")
		if table == "" {
			http.Error(w, "No table provided", http.StatusBadRequest)
			return
		}
		tablePath := filepath.Join(startupCacheDir, chain, table)
		files, err := os.ReadDir(tablePath)
		if err != nil && os.IsNotExist(err) {
			respStr := fmt.Sprintf("No chunks found at %s", tablePath)
			http.Error(w, respStr, http.StatusNotFound)
		}
		if err != nil && !os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("Error reading chain directory: %s", err), http.StatusInternalServerError)
			return
		}
		chunks := []string{}
		for _, file := range files {
			if !file.IsDir() && strings.HasPrefix(file.Name(), table+"_") {
				chunks = append(chunks, file.Name())
			}
		}
		slices.SortFunc(chunks, func(i, j string) int {
			iBlockStr := strings.Split(strings.Split(i, "_")[1], "-")[0]
			iBlock, _ := strconv.Atoi(iBlockStr)
			jBlockStr := strings.Split(strings.Split(j, "_")[1], "-")[0]
			jBlock, _ := strconv.Atoi(jBlockStr)
			if iBlock < jBlock {
				return -1
			} else if iBlock > jBlock {
				return 1
			}
			return 0
		})
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(chunks)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding response: %s", err), http.StatusInternalServerError)
		}
	}
	mux.Handle("/gcgo-startupcache/{chain}/{table}/chunks.json", passwordMiddleware(http.HandlerFunc(chunksHandler)))

	// To simplify partially clearing the cache
	deleteHandler := func(w http.ResponseWriter, r *http.Request) {
		var err error
		chain := r.PathValue("chain")
		if chain == "" {
			http.Error(w, "No chain provided", http.StatusBadRequest)
			return
		}
		table := r.PathValue("table")
		afterStr := r.PathValue("after")
		after := 0
		if afterStr != "" {
			after, err = strconv.Atoi(afterStr)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid value for `after`: %s", err), http.StatusBadRequest)
				return
			}
		}
		log.Println("Deleting entries for", chain, table, "after", after)

		tables := []string{table}
		if table == "" {
			tables = tables[:0]
			tableFiles, err := os.ReadDir(filepath.Join(startupCacheDir, chain))
			if err != nil {
				http.Error(w, fmt.Sprintf("Error reading chain directory: %s", err), http.StatusInternalServerError)
				return
			}
			for _, file := range tableFiles {
				if file.IsDir() {
					tables = append(tables, file.Name())
				}
			}
		}

		deleted := []string{}
		for _, table := range tables {
			tablePath := filepath.Join(startupCacheDir, chain, table)
			chunks, err := os.ReadDir(tablePath)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error reading table directory: %s", err), http.StatusInternalServerError)
				return
			}
			for _, chunk := range chunks {
				if !chunk.IsDir() && strings.HasPrefix(chunk.Name(), table+"_") {
					chunkBlockStr := strings.Split(strings.Split(chunk.Name(), "_")[1], "-")[0]
					chunkBlock, err := strconv.Atoi(chunkBlockStr)
					if err != nil {
						http.Error(w, fmt.Sprintf("Error parsing block number for %s: %s", chunkBlockStr, err), http.StatusInternalServerError)
						return
					}

					if chunkBlock > after {
						err = os.Remove(filepath.Join(tablePath, chunk.Name()))
						deleted = append(deleted, chunk.Name())
						if err != nil {
							http.Error(w, fmt.Sprintf("Error deleting file: %s", err), http.StatusInternalServerError)
							return
						}
					}
				}
			}
		}

		fmt.Fprintf(w, "Deleted %d entries: %v", len(deleted), deleted)
	}

	// To overwrite the cache
	setHandler := func(w http.ResponseWriter, r *http.Request) {
		var err error
		chain := r.PathValue("chain")
		if chain == "" {
			http.Error(w, "No chain provided", http.StatusBadRequest)
			return
		}
		table := r.PathValue("table")
		if table == "" {
			http.Error(w, "No table provided", http.StatusBadRequest)
			return
		}
		filename := r.PathValue("filename")
		if filename == "" {
			http.Error(w, "No filename provided", http.StatusBadRequest)
			return
		}

		os.Mkdir(filepath.Join(startupCacheDir, chain), 0755)
		os.Mkdir(filepath.Join(startupCacheDir, chain, table), 0755)

		file, err := os.OpenFile(filepath.Join(startupCacheDir, chain, table, "."+filename), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error opening file: %s", err), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		_, err = io.Copy(file, r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error writing file: %s", err), http.StatusInternalServerError)
			return
		}

		err = os.Rename(filepath.Join(startupCacheDir, chain, table, "."+filename), filepath.Join(startupCacheDir, chain, table, filename))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error renaming file: %s", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Saved %s", filepath.Join(startupCacheDir, chain, table, filename))
	}

	mux.Handle("/gcgo-startupcache/delete/{chain}/{table}/{after}", passwordMiddleware(http.HandlerFunc(deleteHandler)))
	mux.Handle("/gcgo-startupcache/delete/{chain}/{table}/all", passwordMiddleware(http.HandlerFunc(deleteHandler)))
	mux.Handle("/gcgo-startupcache/delete/{chain}/all", passwordMiddleware(http.HandlerFunc(deleteHandler)))

	mux.Handle("/gcgo-startupcache/set/{chain}/{table}/{filename}", passwordMiddleware(http.HandlerFunc(setHandler)))
	go http.ListenAndServe(listenAddr, mux)
}
