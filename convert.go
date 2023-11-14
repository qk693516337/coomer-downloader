package main

import (
	"github.com/pterm/pterm"
	"github.com/spf13/afero"
	"github.com/thoas/go-funk"
	"path/filepath"
	"strings"
	"sync"
)

var fs = afero.NewOsFs()

func ConvertImages(downloads []Download) {
	var wg sync.WaitGroup // wait group to wait for all goroutines to complete
	sem := make(chan struct{}, 5)

	filteredDownloads := funk.Filter(downloads, func(download Download) bool {
		extension := filepath.Ext(download.FilePath)
		return extension == ".jpg" || extension == ".jpeg" || extension == ".png"
	}).([]Download)

	if len(filteredDownloads) > 0 {
		pterm.Println("\n⚙️ Converting images to AVIF...")
	}

	for _, image := range filteredDownloads {
		wg.Add(1)

		go func(download Download) {
			defer wg.Done()
			sem <- struct{}{} // acquire a semaphore token

			outputFile := replaceExtension(download.FilePath, ".avif")
			fileName := filepath.Base(download.FilePath)

			pterm.Printf("["+pterm.Green("C")+"] Converting %s to %s...\n", pterm.Bold.Sprintf(fileName),
				pterm.Magenta("AVIF"))
			ConvertToAvif(download.FilePath, outputFile)

			// If the file was converted successfully, then we delete the original file
			if fileExists(outputFile) && fileSize(outputFile) > 0 {
				_ = fs.Remove(download.FilePath)
			}

			<-sem
		}(image)
	}

	wg.Wait()
	close(sem)
}

func ConvertVideos(downloads []Download) {
	var wg sync.WaitGroup // wait group to wait for all goroutines to complete
	sem := make(chan struct{}, 5)

	filteredDownloads := funk.Filter(downloads, func(download Download) bool {
		extension := filepath.Ext(download.FilePath)
		return extension == ".gif" || extension == ".mp4" || extension == ".m4v"
	}).([]Download)

	if len(filteredDownloads) > 0 {
		pterm.Println("\n⚙️ Converting videos to AV1...")
	}

	for _, video := range filteredDownloads {
		wg.Add(1)

		go func(download Download) {
			defer wg.Done()
			sem <- struct{}{} // acquire a semaphore token

			outputFile := replaceExtension(download.FilePath, ".mkv")
			fileName := filepath.Base(download.FilePath)

			pterm.Printf("["+pterm.Green("C")+"] Converting %s to %s...\n", pterm.Bold.Sprintf(fileName),
				pterm.Yellow("AV1"))
			ConvertToAv1(download.FilePath, outputFile)

			// If the file was converted successfully, then we delete the original file
			if fileExists(outputFile) && fileSize(outputFile) > 0 {
				_ = fs.Remove(download.FilePath)
			}

			<-sem
		}(video)
	}

	wg.Wait()
	close(sem)
}

func RemoveDuplicates(downloads []Download) int {
	numDeleted := 0

	duplicates := make(map[string][]Download)
	for _, download := range downloads {
		duplicates[download.Hash] = append(duplicates[download.Hash], download)
	}

	pterm.Println("\n🚮 Removing duplicated downloads...")

	for _, value := range duplicates {
		deleteList := value[1:]

		for _, deleteFile := range deleteList {
			pterm.Printf("["+pterm.LightRed("D")+"] %s...\n", pterm.Bold.Sprintf(deleteFile.FilePath))
			numDeleted++
			_ = fs.Remove(deleteFile.FilePath)
		}
	}

	return numDeleted
}

// region - Private functions

func replaceExtension(filePath string, newExtension string) string {
	extension := filepath.Ext(filePath)
	return strings.Replace(filePath, extension, newExtension, 1)
}

func fileExists(filePath string) bool {
	exists, err := afero.Exists(fs, filePath)
	if err != nil {
		return false
	}

	return exists
}

func fileSize(filePath string) int64 {
	file, err := fs.Open(filePath)
	if err != nil {
		return -1
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return -1
	}

	return fileInfo.Size()
}

// endregion
