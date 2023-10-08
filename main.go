package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jbonadiman/cover-downloader/internal/downloader"
	"github.com/jbonadiman/cover-downloader/internal/parser"
)

const coverURLTemplate = "https://raw.githubusercontent.com/xlenore/ps2-covers/main/covers/%s.jpg"

type Args struct {
	Path string
}

func main() {
	args := parseArgs()
	installationPath := args.Path

	cacheFile := filepath.Join(installationPath, "cache", "gamelist.cache")
	gameIndexFile := filepath.Join(installationPath, "resources", "GameIndex.yaml")
	coversFolder := filepath.Join(installationPath, "covers")

	// TODO: translate game serial to name
	log.Println(gameIndexFile)

	allGamesSerials, err := readGameSerials(cacheFile)
	if err != nil {
		log.Fatalf("failed to parse game serials: %v", err)
	}
	log.Printf("found %d games\n", len(allGamesSerials))

	if (len(allGamesSerials)) == 0 {
		log.Fatalf("no games found in cache file %s", cacheFile)
	}

	existingCovers, err := getExistingCovers(coversFolder)
	if err != nil {
		log.Fatalf("failed to get existing covers: %v", err)
	}

	wg := sync.WaitGroup{}
	for serial := range allGamesSerials {
		if _, ok := existingCovers[serial]; ok {
			log.Printf("skipping existing cover for %s", serial)
			continue
		}

		wg.Add(1)
		go func(serial string) {
			defer wg.Done()

			log.Printf("downloading %s cover...\n", serial)
			err = downloadCover(coversFolder, serial)
			if err != nil {
				if err == downloader.ErrMissingFile {
					log.Printf("skipping missing cover for %s\n", serial)
					return
				}

				log.Fatalf("failed to download cover for %s: %v", serial, err)
			}
		}(serial)
	}

	wg.Wait()
}

func parseArgs() Args {
	var args Args
	const cwd = "current working directory"

	fs := flag.NewFlagSet(path.Base(os.Args[0]), flag.ExitOnError)

	installationPath := fs.String("path", cwd, "pcsx2 installation path")
	fs.StringVar(installationPath, "p", *installationPath, "alias for --path")

	fs.Parse(os.Args[1:])

	if *installationPath == cwd {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("failed to get current working directory: %v", err)
		}

		*installationPath = wd
	}

	args.Path = *installationPath

	return args
}

func getExistingCovers(coversFolder string) (map[string]bool, error) {
	coverFiles, err := filepath.Glob(path.Join(coversFolder, "*.jpg"))
	if err != nil {
		log.Fatalf("failed listing covers: %v", err)
	}

	existingCovers := make(map[string]bool)
	for _, file := range coverFiles {
		cover := strings.TrimSuffix(filepath.Base(file), ".jpg")
		existingCovers[cover] = true
	}

	return existingCovers, nil
}

func readGameSerials(cacheFilePath string) (map[string]bool, error) {
	file, err := os.Open(cacheFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache file: %v", err)
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %v", err)
	}

	gameSerials := parser.GetGameSerialSet(string(fileContent))

	return gameSerials, nil
}

func downloadCover(coversFolder string, serial string) error {
	coverURL := fmt.Sprintf(coverURLTemplate, serial)
	path := filepath.Join(coversFolder, strings.ToUpper(serial)+".jpg")

	content, err := downloader.DownloadFile(coverURL)
	if err != nil {
		return err
	}

	err = saveFile(content, path)
	if err != nil {
		return err
	}

	return nil
}

func saveFile(content []byte, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return err
	}

	return nil
}
