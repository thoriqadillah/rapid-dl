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

type progressBar struct {
	Entry
	onprogress OnProgress
	reader     io.ReadCloser
	index      int
	downloaded int64
	progress   float64
	chunkSize  int64
}

func (r *progressBar) Read(payload []byte) (n int, err error) {
	n, err = r.reader.Read(payload)
	if err != nil {
		return n, err
	}

	r.downloaded += int64(n)
	r.progress = float64(100 * r.downloaded / r.chunkSize)

	if r.onprogress != nil {
		r.onprogress(
			r.ID(),
			r.index,
			r.downloaded,
			r.progress,
		)
	}

	return n, err
}

func (r *progressBar) Close() error {
	return r.reader.Close()
}

type chunk struct {
	Entry
	Setting
	wg         *sync.WaitGroup
	index      int
	start      int64
	end        int64
	size       int64
	logger     Logger
	onprogress OnProgress
}

func calculatePosition(entry Entry, chunkSize int64, index int) (int64, int64) {
	start := int64(index * int(chunkSize))
	end := start + (chunkSize - 1)

	if index == int(entry.ChunkLen())-1 {
		end = entry.Size()
	}

	return start, end
}

func newChunk(entry Entry, index int, setting Setting, wg *sync.WaitGroup) *chunk {
	chunkSize := entry.Size() / int64(entry.ChunkLen()) // TODO: make this absolute
	start, end := calculatePosition(entry, chunkSize, index)

	logger := NewLogger(setting)
	logger.Print("Downloading chunk", index+1, "from", start, "to", end, fmt.Sprintf("(~%d MB)", (end-start)/(1024*1024)))

	return &chunk{
		Entry:      entry,
		Setting:    setting,
		wg:         wg,
		index:      index,
		start:      start,
		end:        end,
		size:       chunkSize,
		logger:     logger,
		onprogress: nil,
	}
}

func (c *chunk) download(ctx context.Context) error {
	defer c.wg.Done()
	start := time.Now()

	srcFile, err := c.getDownloadFile(ctx)
	if err != nil {
		c.logger.Print("Error fetching chunk file:", err.Error())
		return err
	}
	defer srcFile.Close()

	dstFile, err := c.getSaveFile()
	if err != nil {
		c.logger.Print("Error creating temp file for chunk:", err.Error())
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		c.logger.Print("Error while downloading chunk:", err.Error())
		return err
	}

	elapsed := time.Since(start)
	c.logger.Print("Chunk", c.index+1, "downloaded in", elapsed.Seconds(), "s")

	return nil
}

func (c *chunk) Execute(ctx context.Context) error {
	return c.download(ctx)
}

func (c *chunk) OnError(ctx context.Context, err error) {
	var e error
	for i := 0; i < c.MaxRetry(); i++ {
		c.wg.Add(1)
		c.logger.Print("Error while downloading file:", err.Error(), ". Retrying...")

		// TODO: retry and resume download from the last byte download
		if e = c.download(ctx); e == nil {
			return
		}
	}

	c.logger.Print("Failed downloading file:", err.Error())
}

func (c *chunk) onProgress(onprogress OnProgress) {
	c.onprogress = onprogress
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

	progressBar := &progressBar{
		onprogress: c.onprogress,
		reader:     res.Body,
		Entry:      c.Entry,
		index:      c.index,
		downloaded: 0,
		progress:   0,
		chunkSize:  c.size,
	}

	return progressBar, nil
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
