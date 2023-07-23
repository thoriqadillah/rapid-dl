package rapid

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// downloader that save the result into local file
type localDownloader struct {
	Setting
	ctx    context.Context
	cancel context.CancelFunc
	logger Logger
}

var DownloaderDefault = "default"

func newLocalDownloader(setting Setting) Downloader {
	ctx, cancel := context.WithCancel(context.Background())

	return &localDownloader{
		Setting: setting,
		ctx:     ctx,
		cancel:  cancel,
		logger:  NewLogger(setting.LoggerProvider(), setting),
	}
}

func (dl *localDownloader) Download(entry Entry) error {
	start := time.Now()

	worker, err := NewWorker(dl.ctx, dl.Poolsize(), entry.ChunkLen(), dl.Setting)
	if err != nil {
		dl.logger.Print("Error while creating worker", err.Error())
		return err
	}

	var wg sync.WaitGroup
	worker.Start()
	defer worker.Stop()

	chunks := make([]*chunk, entry.ChunkLen())
	for i := 0; i < entry.ChunkLen(); i++ {
		chunks[i] = newChunk(entry, i, dl.Setting, &wg)
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
	dl.logger.Print(entry.Name(), "downloaded in", elapsed.Seconds())

	return nil
}

func (dl *localDownloader) Resume(entry Entry) error {
	//TODO: implement resume
	return nil
}

func (dl *localDownloader) Restart(entry Entry) error {
	//TODO: implement restart
	return nil
}

func (dl *localDownloader) Stop(entry Entry) error {
	dl.cancel()
	return nil
}

//TODO: implement watcher

func (dl *localDownloader) handleDuplicate(filename string) string {
	name := filename
	if _, err := os.Stat(filename); err != nil {
		return name
	}

	regex, err := regexp.Compile(`\((.*?)\)`)
	if err != nil { // if there is no number prefix
		return name
	}

	prefix := regex.FindStringSubmatch(name)
	if len(prefix) == 0 {
		split := strings.Split(name, ".")
		if len(split) > 2 {
			split[len(split)-2] += " (1)"
		} else {
			split[0] += " (1)"
		}
		name = strings.Join(split, ".")
		name = dl.handleDuplicate(name)
		return name
	}

	name = strings.ReplaceAll(name, " "+prefix[0], "")
	number, err := strconv.Atoi(prefix[1])
	if err != nil {
		return name
	}
	split := strings.Split(name, ".")
	if len(split) > 2 {
		split[len(split)-2] += " (" + strconv.Itoa(number+1) + ")"
	} else {
		split[0] += " (" + strconv.Itoa(number+1) + ")"
	}
	name = strings.Join(split, ".")
	name = dl.handleDuplicate(name)

	return name
}

func (dl *localDownloader) createFile(entry Entry) error {
	//TODO: handle filename duplication
	filename := filepath.Join(dl.DownloadLocation(), entry.Name())
	filename = dl.handleDuplicate(filename)

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		dl.logger.Print("Error while creating downloaded file:", err.Error())
		return err
	}

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
