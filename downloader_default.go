package rapid

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// downloader that save the result into local file
type localDownloader struct {
	Setting
	logger     Logger
	onprogress OnProgress
}

var DownloaderDefault = "default"

func newLocalDownloader(setting Setting) Downloader {
	return &localDownloader{
		Setting: setting,
		logger:  NewLogger(setting),
	}
}

func (dl *localDownloader) Download(entry Entry) error {
	start := time.Now()

	worker, err := NewWorker(entry.Context(), entry.ChunkLen(), entry.ChunkLen(), dl.Setting)
	if err != nil {
		dl.logger.Print("Error while creating worker", err.Error())
		return err
	}

	var wg sync.WaitGroup
	worker.Start()
	defer worker.Stop()

	// TODO: handle do not download in chunk in ChunkLen() is 1. Just direct download
	// This is the possibility if the entry is unchunkable, less than min chunk size, or unresumable
	chunks := make([]*chunk, entry.ChunkLen())
	for i := 0; i < entry.ChunkLen(); i++ {
		chunks[i] = newChunk(entry, i, dl.Setting, &wg)

		if dl.onprogress != nil {
			chunks[i].onProgress(dl.onprogress)
		}
	}

	for _, chunk := range chunks {
		wg.Add(1)
		worker.Add(chunk)
	}

	wg.Wait()

	// combining file
	if err := dl.createFile(entry); err != nil {
		dl.logger.Print("Error while combining chunks:", err.Error())
		return err
	}

	elapsed := time.Since(start)
	dl.logger.Print(entry.Name(), "downloaded in", elapsed.Seconds(), "s")

	return nil
}

func (dl *localDownloader) Resume(entry Entry) error {
	dl.logger.Print("Resuming download", entry.Name(), "...")

	//TODO: check if link expired
	//TODO: check if context is canceled

	if !entry.Resumable() {
		dl.logger.Print(entry.Name(), "is not resumable. Restarting...")
		return dl.Restart(entry)
	}

	//TODO: implement resume
	return nil
}

func (dl *localDownloader) Restart(entry Entry) error {
	dl.logger.Print("Restarting download", entry.Name(), "...")
	//TODO: implement restart
	//TODO: check if link expired
	return nil
}

func (dl *localDownloader) Stop(entry Entry) error {
	dl.logger.Print("Stopping download", entry.Name(), "...")

	entry.Cancel()
	return nil
}

// Watch will update the id, index, downloaded bytes, and progress in percent of chunks. Watch must be called before Download
func (dl *localDownloader) Watch(update OnProgress) {
	dl.onprogress = update
}

// createFile will combine chunks into single actual file
func (dl *localDownloader) createFile(entry Entry) error {
	filename := filepath.Join(dl.DownloadLocation(), entry.Name())

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		dl.logger.Print("Error while creating downloaded file:", err.Error())
		return err
	}

	// TODO: if chunk len is 1, then just rename the chunk into filename

	for i := 0; i < entry.ChunkLen(); i++ {
		tempFilename := filepath.Join(dl.DownloadLocation(), fmt.Sprintf("%s-%d", entry.ID(), i))
		tmpFile, err := os.Open(tempFilename)
		if err != nil {
			dl.logger.Print("Error while opening downloaded chunk file:", err.Error())
			return err
		}

		if _, err := io.Copy(file, tmpFile); err != nil {
			dl.logger.Print("Error while copying chunk file into actual file:", err.Error())
			return err
		}

		if err := os.Remove(tempFilename); err != nil {
			dl.logger.Print("Error while removing temp file:", err.Error())
			return err
		}
	}

	return nil
}

func init() {
	RegisterDownloader(DownloaderDefault, newLocalDownloader)
}
