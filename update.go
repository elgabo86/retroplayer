package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type fetchResult struct {
	site   string
	folder string
	links  []Link
}

func updateCache(db *sql.DB) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Erreur lors de l'obtention du répertoire home: %v\n", err)
		return
	}

	cacheDir := filepath.Join(homeDir, ".cache", "retroplayer")

	fmt.Println("Mise à jour du cache en cours...")
	fmt.Printf("Cela peut prendre quelques secondes.\n\n")

	folders := []string{"0-9"}
	for c := 'a'; c <= 'z'; c++ {
		folders = append(folders, string(c))
	}
	folders = append(folders, "zzz_prototypes", "zzz_unlicensed")

	totalJobs := len(baseSites) * len(folders)
	var completed atomic.Int64

	// Canal pour les résultats
	results := make(chan fetchResult, 100)

	// Pool de workers concurrents (1 worker par site)
	var wg sync.WaitGroup
	for _, site := range baseSites {
		wg.Add(1)
		go func(site string) {
			defer wg.Done()
			for _, folder := range folders {
				url := fmt.Sprintf("%s/%s/", site, folder)
				content, err := fetchURL(url)
				if err != nil {
					completed.Add(1)
					continue
				}
				links := extract7zLinks(content)
				results <- fetchResult{site: site, folder: folder, links: links}
				completed.Add(1)
			}
		}(site)
	}

	// Fermer le canal results quand tous les workers ont fini
	go func() {
		wg.Wait()
		close(results)
	}()

	// Afficher la progression pendant la récupération
	progressDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-progressDone:
				return
			case <-ticker.C:
				done := completed.Load()
				pct := int(float64(done) / float64(totalJobs) * 100)
				fmt.Printf("\r\033[KProgression: %d/%d (%d%%)", done, totalJobs, pct)
				os.Stdout.Sync()
				if int(done) >= totalJobs {
					fmt.Println()
					return
				}
			}
		}
	}()

	// Préparer la transaction pour l'insertion
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("\nErreur lors du démarrage de la transaction: %v\n", err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO files (site, folder, file_name, file_path) VALUES (?, ?, ?, ?)")
	if err != nil {
		fmt.Printf("\nErreur lors de la préparation de l'insertion: %v\n", err)
		return
	}
	defer stmt.Close()

	// Insérer les résultats au fur et à mesure qu'ils arrivent
	for result := range results {
		for _, link := range result.links {
			stmt.Exec(result.site, result.folder, link.name, link.path)
		}
	}

	close(progressDone)

	fmt.Println()

	// Valider la transaction
	err = tx.Commit()
	if err != nil {
		fmt.Printf("Erreur lors de la validation de la transaction: %v\n", err)
		return
	}

	fmt.Println("Mise à jour du cache terminée.")
	fmt.Printf("Les fichiers sont stockés dans: %s\n", cacheDir)
}

func fetchURL(url string) (string, error) {
	// Créer un client HTTP avec timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Faire la requête
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Vérifier le code de statut
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Lire le contenu (jusqu'à 1 Mo)
	limitedReader := io.LimitReader(resp.Body, 1024*1024)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

type Link struct {
	name string
	path string
}

func extract7zLinks(content string) []Link {
	var links []Link
	
	// Expression régulière pour trouver les liens .7z
	re := regexp.MustCompile(`href="([^"]*\.7z)">([^<]*)`)
	matches := re.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) == 3 {
			path := match[1]
			name := match[2]
			// Décoder les entités HTML dans le chemin
			path = decodeHTMLEntities(path)
			links = append(links, Link{name: name, path: path})
		}
	}
	
	return links
}

