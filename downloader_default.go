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
	setting    Setting
	logger     Logger
	onprogress OnProgress
}

var DownloaderDefault = "default"

func newLocalDownloader(opt *downloaderOption) Downloader {
	return &localDownloader{
		setting: opt.setting,
		logger:  NewLogger(opt.setting),
	}
}

func (dl *localDownloader) Download(entry Entry) error {
	start := time.Now()

	if entry.Expired() {
		return errUrlExpired
	}

	worker, err := NewWorker(entry.Context(), entry.ChunkLen(), entry.ChunkLen(), dl.setting)
	if err != nil {
		dl.logger.Print("Error creating worker", err.Error())
		return err
	}

	var wg sync.WaitGroup
	worker.Start()
	defer worker.Stop()

	chunks := make([]*chunk, entry.ChunkLen())
	for i := 0; i < entry.ChunkLen(); i++ {
		chunks[i] = newChunk(entry, i, dl.setting, &wg)

		if dl.onprogress != nil {
			chunks[i].onProgress(dl.onprogress)
		}
	}

	for _, chunk := range chunks {
		wg.Add(1)
		worker.Add(chunk)
	}

	wg.Wait()

	if entry.Context().Err() != nil {
		return nil
	}

	// combining file
	if err := dl.createFile(entry); err != nil {
		dl.logger.Print("Error combining chunks:", err.Error())
		return err
	}

	elapsed := time.Since(start)
	dl.logger.Print(entry.Name(), "downloaded  in", elapsed.Seconds(), "s")

	return nil
}

var errUrlExpired = fmt.Errorf("link is expired")

func (dl *localDownloader) Resume(entry Entry) error {
	start := time.Now()

	if entry.Expired() {
		return errUrlExpired
	}

	// check if context is canceled (download stoppped by user)
	if err := entry.Refresh(); err != nil {
		return err
	}

	dl.logger.Print("Resuming download", entry.Name(), "...")

	if !entry.Resumable() {
		dl.logger.Print(entry.Name(), "does not support resume download. Restarting...")
		return dl.Download(entry)
	}

	worker, err := NewWorker(entry.Context(), entry.ChunkLen(), entry.ChunkLen(), dl.setting)
	if err != nil {
		dl.logger.Print("Error creating worker", err.Error())
		return err
	}

	var wg sync.WaitGroup
	worker.Start()
	defer worker.Stop()

	chunks := make([]*chunk, 0)
	for i := 0; i < entry.ChunkLen(); i++ {
		chunk := newChunk(entry, i, dl.setting, &wg)
		if file, err := os.Stat(chunk.path); err == nil && file.Size() == chunk.size {
			continue
		}

		chunk.start += resumePosition(chunk.path)
		chunks = append(chunks, chunk)
	}

	for _, chunk := range chunks {
		wg.Add(1)
		worker.Add(chunk)
	}

	wg.Wait()

	// // combining file
	if err := dl.createFile(entry); err != nil {
		dl.logger.Print("Error combining chunks:", err.Error())
		return err
	}

	elapsed := time.Since(start)
	dl.logger.Print(entry.Name(), "resumed in", elapsed.Seconds(), "s")

	return nil
}

func (dl *localDownloader) Restart(entry Entry) error {
	dl.logger.Print("Restarting download", entry.Name(), "...")

	if entry.Expired() {
		return errUrlExpired
	}

	// check if context is canceled (download stoppped by user)
	if err := entry.Refresh(); err != nil {
		return err
	}

	// remove the downloaded chunk if any
	for i := 0; i < entry.ChunkLen(); i++ {
		chunkFile := filepath.Join(dl.setting.DownloadLocation(), fmt.Sprintf("%s-%d", entry.ID(), i))
		if err := os.Remove(chunkFile); err != nil {
			return err
		}
	}

	return dl.Download(entry)
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
	file, err := os.Create(entry.Location())
	if err != nil {
		dl.logger.Print("Error creating downloaded file:", err.Error())
		return err
	}

	defer file.Close()

	// if chunk len is 1, then just rename the chunk into entry filename
	// we assume if the chunk len is 1, then it is not chunkable and unresumable
	if entry.ChunkLen() == 1 {
		chunkname := filepath.Join(dl.setting.DownloadLocation(), fmt.Sprintf("%s-%d", entry.ID(), 0))
		return os.Rename(chunkname, entry.Location())
	}

	for i := 0; i < entry.ChunkLen(); i++ {
		tmpFilename := filepath.Join(dl.setting.DownloadLocation(), fmt.Sprintf("%s-%d", entry.ID(), i))
		tmpFile, err := os.Open(tmpFilename)
		if err != nil {
			dl.logger.Print("Error opening downloaded chunk file:", err.Error())
			return err
		}

		if _, err := io.Copy(file, tmpFile); err != nil {
			dl.logger.Print("Error copying chunk file into actual file:", err.Error())
			return err
		}

		if err := os.Remove(tmpFilename); err != nil {
			dl.logger.Print("Error removing temp file:", err.Error())
			return err
		}
	}

	return nil
}

func init() {
	RegisterDownloader(DownloaderDefault, newLocalDownloader)
}
