package rapid

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type chunk struct {
	Entry
	Setting
	wg     *sync.WaitGroup
	index  int
	start  int64
	end    int64
	size   int64
	logger Logger
}

func newChunk(entry Entry, index int, setting Setting, wg *sync.WaitGroup) *chunk {
	chunkSize := entry.Size() / int64(entry.ChunkLen())
	start := int64(index * int(chunkSize))
	end := start + (chunkSize - 1)

	if index == int(entry.ChunkLen())-1 {
		end = entry.Size()
	}

	logger := NewLogger(setting.LoggerProvider(), setting)

	if chunkSize == -1 {
		logger.Print("Downloading chunk", index+1, "with unknown size")
	} else {
		logger.Print("Downloading chunk", index+1, "from", start, "to", end, fmt.Sprintf("(~%d MB)", (end-start)/(1024*1024)))
	}

	return &chunk{
		Entry:   entry,
		Setting: setting,
		wg:      wg,
		index:   index,
		start:   start,
		end:     end,
		size:    chunkSize,
		logger:  logger,
	}
}

func (c *chunk) download(ctx context.Context) error {
	start := time.Now()

	srcFile, err := c.getDownloadFile(ctx)
	if err != nil {
		c.logger.Print("Error fetching chunk file:", err.Error())
		return err
	}

	dstFile, err := c.getSaveFile()
	if err != nil {
		c.logger.Print("Error creating temp file for chunk:", err.Error())
		return err
	}

	// TODO: implement watch or progress bar

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		c.logger.Print("Error while downloading chunk", err.Error())
		return err
	}

	elapsed := time.Since(start)
	c.logger.Print("Chunk", c.index+1, "downloaded in", elapsed.Seconds())

	c.wg.Done()
	return nil
}

func (c *chunk) Execute(ctx context.Context) error {
	return c.download(ctx)
}

func (c *chunk) OnError(ctx context.Context, err error) {
	defer c.wg.Done()
	c.logger.Print("Error while downloading file:", err.Error(), ". Retrying...")

	var e error
	for i := 0; i < c.MaxRetry(); i++ {
		// TODO: retry and resume download from the last byte download
		if e = c.download(ctx); e != nil {
			continue
		}

		e = nil
		break
	}

	if e != nil {
		c.logger.Print("Failed downloading file:", err.Error())
	}
}

func (c *chunk) getDownloadFile(ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.URL(), nil)
	if err != nil {
		c.logger.Print("Error while creating chunk request:", err.Error())
		return nil, err
	}

	bytesRange := fmt.Sprintf("bytes=%d-%d", c.start, c.end)
	req.Header.Add("Range", bytesRange)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.Print("Error while fething chunk body:", err.Error())
		return nil, err
	}

	return res.Body, nil
}

func (c *chunk) getSaveFile() (io.WriteCloser, error) {
	tmpFilename := filepath.Join(c.DownloadLocation(), fmt.Sprintf("%s-%d", c.ID(), c.index))
	file, err := os.OpenFile(tmpFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		c.logger.Print("Error creating or appending file:", err.Error())
		return nil, err
	}

	return file, nil
}
