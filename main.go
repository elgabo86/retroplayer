package main

import (
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
)

var (
	baseSites = []string{
		"https://spc.joshw.info",
		"https://nsf.joshw.info",
		"https://usf.joshw.info",
		"https://smd.joshw.info",
		"https://psf.joshw.info",
		"https://psf2.joshw.info",
		"https://psf3.joshw.info",
		"https://gbs.joshw.info",
		"https://gsf.joshw.info",
		"https://2sf.joshw.info",
		"https://3sf.joshw.info",
		"https://dsf.joshw.info",
		"https://wii.joshw.info",
		"https://wiiu.joshw.info",
		"https://gcn.joshw.info",
		"https://ssf.joshw.info",
		"https://xbox.joshw.info",
		"https://x360.joshw.info",
		"https://psp.joshw.info",
		"https://pc.joshw.info",
		"https://psf4.joshw.info",
		"https://switch.joshw.info",
		"https://vita.joshw.info",
		"https://3do.joshw.info",
	}
	
	// Mapping des sites vers leurs noms abrégés
	siteNames = map[string]string{
		"https://spc.joshw.info":  "SNES",
		"https://nsf.joshw.info":  "NES",
		"https://usf.joshw.info":  "N64",
		"https://smd.joshw.info":  "GEN",
		"https://psf.joshw.info":  "PS1",
		"https://psf2.joshw.info": "PS2",
		"https://psf3.joshw.info": "PS3",
		"https://gbs.joshw.info":  "GB",
		"https://gsf.joshw.info":  "GBA",
		"https://2sf.joshw.info":  "NDS",
		"https://3sf.joshw.info":  "3DS",
		"https://dsf.joshw.info":  "DC",
		"https://wii.joshw.info":  "Wii",
		"https://wiiu.joshw.info": "WiiU",
		"https://gcn.joshw.info":  "GCN",
		"https://ssf.joshw.info":  "SAT",
		"https://xbox.joshw.info": "XBX",
		"https://x360.joshw.info": "X360",
		"https://psp.joshw.info":  "PSP",
		"https://pc.joshw.info":   "PC",
		"https://psf4.joshw.info":  "PS4",
		"https://switch.joshw.info": "SWITCH",
		"https://vita.joshw.info":  "VITA",
		"https://3do.joshw.info":  "3DO",
	}
	
	app         *tview.Application
	fileList    *tview.List
	inputField  *tview.InputField
	statusText  *tview.TextView
	progressBar *tview.TextView
	allFiles    []File
	db          *sql.DB
	dbFile      string
	tempDir     string
	downloadCancel chan bool
)

type File struct {
	ID       int
	Site     string
	Folder   string
	Name     string
	Path     string
	FullName string
}

func main() {
	// Initialiser le générateur de nombres aléatoires
	rand.Seed(time.Now().UnixNano())
	
	// Configurer les chemins
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Erreur lors de l'obtention du répertoire home: %v\n", err)
		os.Exit(1)
	}
	
	cacheDir := filepath.Join(homeDir, ".cache", "retroplayer")
	tempDir = cacheDir
	dbFile = filepath.Join(cacheDir, "retroplayer.db")
	
	// Créer le répertoire cache
	os.MkdirAll(cacheDir, 0755)

	// Vérifier les arguments de ligne de commande
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--update":
			updateDatabase()
			return
		case "--help", "-h":
			showHelp()
			return
		default:
			fmt.Printf("Argument inconnu: %s\n", os.Args[1])
			showHelp()
			os.Exit(1)
		}
	}

	// Initialiser la base de données
	db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		fmt.Printf("Erreur lors de l'ouverture de la base de données: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialiser la base de données
	initDB(db)

	// Vérifier si la base de données est vide
	count := 0
	err = db.QueryRow("SELECT COUNT(*) FROM files").Scan(&count)
	if err != nil || count == 0 {
		fmt.Println("Base de données vide. Veuillez lancer 'retroplayer --update' pour la peupler.")
		os.Exit(1)
	}

	// Charger tous les fichiers
	allFiles, err = fetchAllFiles(db)
	if err != nil {
		fmt.Printf("Erreur lors du chargement des fichiers: %v\n", err)
		os.Exit(1)
	}

	// Nettoyer les fichiers temporaires au démarrage
	cleanTempFiles()

	// Initialiser l'interface utilisateur
	initUI()

	// Mettre à jour la liste initiale
	updateList("")

	// Lancer l'application
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func showHelp() {
	fmt.Println("RetroPlayer - Lecteur de musiques rétro")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  retroplayer          # Lancer l'interface graphique")
	fmt.Println("  retroplayer --update # Mettre à jour la base de données")
	fmt.Println("  retroplayer --help   # Afficher cette aide")
	fmt.Println("")
}

func updateDatabase() {
	// Configurer les chemins
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Erreur lors de l'obtention du répertoire home: %v\n", err)
		os.Exit(1)
	}
	
	cacheDir := filepath.Join(homeDir, ".cache", "retroplayer")
	dbFile := filepath.Join(cacheDir, "retroplayer.db")
	
	// Créer le répertoire cache
	os.MkdirAll(cacheDir, 0755)

	// Initialiser la base de données
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		fmt.Printf("Erreur lors de l'ouverture de la base de données: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialiser la base de données
	initDB(db)

	// Mettre à jour le cache
	updateCache(db)
}

func initUI() {
	// Créer l'application
	app = tview.NewApplication()

	// Créer les composants
	fileList = tview.NewList()
	fileList.SetBorder(true).SetTitle(fmt.Sprintf("RetroPlayer - %d fichiers", len(allFiles)))
	fileList.ShowSecondaryText(false)

	inputField = tview.NewInputField().
		SetLabel("Recherche: ").
		SetFieldWidth(50)

	statusText = tview.NewTextView().
		SetText("Utilisez les flèches pour naviguer, Entrée pour jouer, ESC/Ctrl+C pour quitter").
		SetTextStyle(tcell.StyleDefault.Italic(true))
		
	progressBar = tview.NewTextView().
		SetText("").
		SetTextStyle(tcell.StyleDefault.Bold(true))

	// Configuration de la liste
	fileList.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		if i < len(allFiles) {
			searchTerm := strings.ToLower(inputField.GetText())
			filteredFiles := filterFiles(allFiles, searchTerm)
			
			if i < len(filteredFiles) {
				go playFile(filteredFiles[i])
			}
		}
	})

	// Gérer les événements de l'input field avec recherche non-bloquante
	inputField.SetChangedFunc(func(text string) {
		// Lancer la recherche de manière non-bloquante
		go func() {
			// Attendre un peu pour éviter les recherches trop fréquentes
			time.Sleep(300 * time.Millisecond)
			
			// Mettre à jour la liste dans le thread principal
			app.QueueUpdateDraw(func() {
				updateList(text)
			})
		}()
	})

	// Gérer les touches spéciales pour l'input field
	inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Ne rien faire, le focus ne doit jamais être sur ce champ
		return nil
	})

	// Gérer les événements de la liste
	fileList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			// Permettre le comportement par défaut pour naviguer vers le haut
			return event
		case tcell.KeyESC:
			// Annuler le téléchargement en cours s'il y en a un
			if downloadCancel != nil {
				select {
				case downloadCancel <- true:
				default:
				}
				statusText.SetText("Téléchargement annulé")
				progressBar.SetText("")
				// Nettoyer les fichiers temporaires
				cleanTempFiles()
			} else {
				app.Stop()
			}
			return nil
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		case tcell.KeyLeft:
			// Sauter de 5 entrées vers le haut
			currentIndex := fileList.GetCurrentItem()
			newIndex := currentIndex - 5
			if newIndex < 0 {
				newIndex = 0
			}
			fileList.SetCurrentItem(newIndex)
			return nil
		case tcell.KeyRight:
			// Sauter de 5 entrées vers le bas
			currentIndex := fileList.GetCurrentItem()
			totalItems := fileList.GetItemCount()
			newIndex := currentIndex + 5
			if newIndex >= totalItems {
				newIndex = totalItems - 1
			}
			fileList.SetCurrentItem(newIndex)
			return nil
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			// Effacer le dernier caractère du champ de recherche
			currentText := inputField.GetText()
			if len(currentText) > 0 {
				newText := currentText[:len(currentText)-1]
				inputField.SetText(newText)
				// Mettre à jour la liste avec le nouveau texte
				go func() {
					time.Sleep(50 * time.Millisecond)
					app.QueueUpdateDraw(func() {
						updateList(newText)
					})
				}()
			}
			return nil
		case tcell.KeyRune:
			// Ajouter le caractère au champ de recherche sans changer le focus
			if event.Rune() != 0 {
				currentText := inputField.GetText()
				newText := currentText + string(event.Rune())
				inputField.SetText(newText)
				// Mettre à jour la liste avec le nouveau texte
				go func() {
					time.Sleep(50 * time.Millisecond)
					app.QueueUpdateDraw(func() {
						updateList(newText)
					})
				}()
				return nil
			}
		case tcell.KeyCtrlU:
			// Effacer tout le texte de recherche avec Ctrl+U
			inputField.SetText("")
			go func() {
				time.Sleep(50 * time.Millisecond)
				app.QueueUpdateDraw(func() {
					updateList("")
				})
			}()
			return nil
		}
		
		return event
	})

	// Créer le layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(inputField, 3, 0, true).
		AddItem(fileList, 0, 1, true).
		AddItem(progressBar, 1, 0, false).
		AddItem(statusText, 2, 0, false)

	// Définir la fonction principale de l'application
	app.SetRoot(flex, true).SetFocus(fileList)
}

func initDB(db *sql.DB) {
	queries := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = 10000",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA mmap_size = 268435456",
		"CREATE TABLE IF NOT EXISTS files (id INTEGER PRIMARY KEY AUTOINCREMENT, site TEXT NOT NULL, folder TEXT NOT NULL, file_name TEXT NOT NULL, file_path TEXT NOT NULL, UNIQUE(site, folder, file_name))",
		"CREATE INDEX IF NOT EXISTS idx_files_site ON files(site)",
		"CREATE INDEX IF NOT EXISTS idx_files_folder ON files(folder)",
		"CREATE INDEX IF NOT EXISTS idx_files_name ON files(file_name)",
		"CREATE INDEX IF NOT EXISTS idx_files_composite ON files(site, folder, file_name)",
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			panic(err)
		}
	}
}

func fetchAllFiles(db *sql.DB) ([]File, error) {
	rows, err := db.Query("SELECT id, site, folder, file_name, file_path FROM files ORDER BY site, folder, file_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var f File
		err := rows.Scan(&f.ID, &f.Site, &f.Folder, &f.Name, &f.Path)
		if err != nil {
			return nil, err
		}
		
		// Décoder les entités HTML
		f.FullName = decodeHTMLEntities(f.Name)
		files = append(files, f)
	}

	return files, nil
}

func filterFiles(files []File, searchTerm string) []File {
	if searchTerm == "" {
		return files
	}

	var filtered []File
	searchTerm = strings.ToLower(searchTerm)
	
	// Diviser le terme de recherche en mots
	searchWords := strings.Fields(searchTerm)
	
	for _, file := range files {
		// Convertir le nom du fichier en minuscules pour la comparaison
		fileNameLower := strings.ToLower(file.FullName)
		
		// Vérifier si tous les mots de recherche sont présents
		match := true
		for _, word := range searchWords {
			if !strings.Contains(fileNameLower, word) {
				match = false
				break
			}
		}
		
		if match {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

func updateList(searchTerm string) {
	// Sauvegarder l'index actuel
	currentIndex := fileList.GetCurrentItem()
	
	fileList.Clear()
	
	// Filtrer les fichiers selon le terme de recherche
	filteredFiles := filterFiles(allFiles, searchTerm)
	
	// Ajouter les fichiers filtrés à la liste (affichage avec site)
	for _, file := range filteredFiles {
		// Obtenir le nom abrégé du site
		siteName := siteNames[file.Site]
		if siteName == "" {
			siteName = "UNK"
		}
		
		// Supprimer l'extension .7z du nom
		displayName := strings.TrimSuffix(file.FullName, ".7z")
		
		// Afficher avec le format: (SITE) Nom du fichier
		fileList.AddItem(fmt.Sprintf("(%s) %s", siteName, displayName), "", 0, nil)
	}
	
	// Restaurer l'index si possible
	if currentIndex < len(filteredFiles) {
		fileList.SetCurrentItem(currentIndex)
	} else if len(filteredFiles) > 0 {
		fileList.SetCurrentItem(0)
	}
	
	// Mettre à jour le titre avec le nombre de fichiers
	fileList.SetTitle(fmt.Sprintf("RetroPlayer - %d fichiers", len(filteredFiles)))
}



func playFile(file File) {
	// Nettoyer les playlists existantes
	cleanPlaylists()
	
	// Afficher un message de statut avec le nom du fichier en rouge
	app.QueueUpdateDraw(func() {
		statusText.SetText(fmt.Sprintf("Téléchargement en cours: %s", file.FullName))
		statusText.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
		progressBar.SetText("Téléchargement: 0%")
	})

	// Construire l'URL du fichier
	url := buildFileURL(file)
	
	// Créer un nom de fichier temporaire
	tempFileName := fmt.Sprintf("file_%d.7z", rand.Intn(1000000))
	tempFile := filepath.Join(tempDir, tempFileName)
	
	// Canal pour annuler le téléchargement
	downloadCancel = make(chan bool, 1)
	
	// Télécharger le fichier avec progression
	err := downloadFileWithProgress(url, tempFile)
	if err != nil {
		app.QueueUpdateDraw(func() {
			statusText.SetText(fmt.Sprintf("Erreur de téléchargement: %v", err))
			statusText.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite))
			progressBar.SetText("")
		})
		return
	}

	// Afficher un message de statut avec le nom du fichier en rouge
	app.QueueUpdateDraw(func() {
		statusText.SetText(fmt.Sprintf("Extraction de l'archive: %s", file.FullName))
		statusText.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
		progressBar.SetText("Extraction...")
	})

	// Créer un répertoire d'extraction
	extractDir := filepath.Join(tempDir, fmt.Sprintf("extract_%d", rand.Intn(1000000)))
	os.MkdirAll(extractDir, 0755)
	
	// Extraire l'archive
	err = extractArchive(tempFile, extractDir)
	if err != nil {
		app.QueueUpdateDraw(func() {
			statusText.SetText(fmt.Sprintf("Erreur d'extraction: %v", err))
			statusText.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite))
			progressBar.SetText("")
		})
		// Nettoyer les fichiers temporaires
		os.Remove(tempFile)
		os.RemoveAll(extractDir)
		return
	}

	// Afficher un message de statut avec le nom du fichier en rouge
	app.QueueUpdateDraw(func() {
		statusText.SetText(fmt.Sprintf("Lecture en cours: %s", file.FullName))
		statusText.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
		progressBar.SetText("Lecture...")
	})

	// Lancer zxtune-qt dans une goroutine pour ne pas bloquer l'interface
	go func() {
		err := playWithZXTune(extractDir)
		app.QueueUpdateDraw(func() {
			if err != nil {
				statusText.SetText(fmt.Sprintf("Erreur de lecture: %v", err))
			} else {
				statusText.SetText("Lecture terminée. Utilisez les flèches pour naviguer, Entrée pour jouer, ESC/Ctrl+C pour quitter")
			}
			statusText.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite))
			progressBar.SetText("")
		})
		
		// Nettoyer les fichiers temporaires
		os.Remove(tempFile)
		os.RemoveAll(extractDir)
	}()
}

func cleanPlaylists() {
	// Chemin vers le répertoire des playlists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	
	playlistsDir := filepath.Join(homeDir, ".local", "share", "ZXTune", "Playlists")
	
	// Vérifier si le répertoire existe
	if _, err := os.Stat(playlistsDir); os.IsNotExist(err) {
		return
	}
	
	// Supprimer tous les fichiers .xspf
	files, err := os.ReadDir(playlistsDir)
	if err != nil {
		return
	}
	
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".xspf" {
			os.Remove(filepath.Join(playlistsDir, file.Name()))
		}
	}
}

func cleanTempFiles() {
	// Nettoyer les fichiers temporaires dans le répertoire cache
	// mais ne pas supprimer la base de données
	
	// Lire le contenu du répertoire cache
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return
	}
	
	// Supprimer tous les fichiers sauf la base de données
	for _, file := range files {
		if file.Name() != "retroplayer.db" {
			fullPath := filepath.Join(tempDir, file.Name())
			if file.IsDir() {
				os.RemoveAll(fullPath)
			} else {
				os.Remove(fullPath)
			}
		}
	}
}

func buildFileURL(file File) string {
	// Pour les sites .joshw.info, utiliser la structure avec la première lettre du fichier
	if strings.Contains(file.Site, ".joshw.info") {
		// Décoder l'URL pour obtenir le nom réel du fichier
		decodedPath := file.Path
		
		// Extraire la première lettre du fichier
		firstLetter := strings.ToLower(string(decodedPath[0]))
		// Si c'est un chiffre, utiliser 0-9
		if firstLetter >= "0" && firstLetter <= "9" {
			firstLetter = "0-9"
		}
		
		// Construire l'URL
		return fmt.Sprintf("%s/%s/%s", file.Site, firstLetter, decodedPath)
	}
	
	// Pour tous les autres sites, utiliser le chemin direct
	return fmt.Sprintf("%s/%s", file.Site, file.Path)
}

func downloadFileWithProgress(url, filepath string) error {
	// Créer le fichier
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Faire la requête HTTP avec timeout
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Vérifier le code de statut
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Obtenir la taille totale
	total := resp.ContentLength
	var downloaded int64 = 0

	// Créer un buffer pour la copie
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		select {
		case <-downloadCancel:
			// Annulation demandée
			return fmt.Errorf("téléchargement annulé")
		default:
			// Lire les données
			n, err := resp.Body.Read(buf)
			if n > 0 {
				// Écrire les données dans le fichier
				out.Write(buf[:n])
				downloaded += int64(n)
				
				// Mettre à jour la barre de progression
				if total > 0 {
					percentage := int(float64(downloaded) / float64(total) * 100)
					app.QueueUpdateDraw(func() {
						progressBar.SetText(fmt.Sprintf("Téléchargement: %d%%", percentage))
					})
				}
			}
			
			if err != nil {
				if err == io.EOF {
					// Fin du téléchargement
					return nil
				}
				return err
			}
		}
	}
}

func extractArchive(archivePath, extractDir string) error {
	// Utiliser 7z pour extraire l'archive
	cmd := exec.Command("7z", "x", archivePath, "-o"+extractDir, "-y")
	return cmd.Run()
}

func playWithZXTune(dir string) error {
	// Vérifier que zxtune-qt est installé
	_, err := exec.LookPath("zxtune-qt")
	if err != nil {
		return fmt.Errorf("zxtune-qt n'est pas installé")
	}

	// Lancer zxtune-qt
	cmd := exec.Command("zxtune-qt", dir)
	return cmd.Run()
}