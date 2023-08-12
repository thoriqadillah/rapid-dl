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
	entry      Entry
	setting    Setting
	wg         *sync.WaitGroup
	path       string
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

// TODO: test this
func resumePosition(location string) int64 {
	file, err := os.Stat(location)
	if err != nil {
		return 0
	}

	resumePos := file.Size()
	if err := os.Truncate(location, resumePos); err != nil {
		return 0
	}

	return resumePos
}

func newChunk(entry Entry, index int, setting Setting, wg *sync.WaitGroup) *chunk {
	chunkSize := entry.Size() / int64(entry.ChunkLen())
	start, end := calculatePosition(entry, chunkSize, index)

	logger := NewLogger(setting)

	return &chunk{
		path:       filepath.Join(setting.DownloadLocation(), fmt.Sprintf("%s-%d", entry.ID(), index)),
		entry:      entry,
		setting:    setting,
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
	c.logger.Print("Downloading chunk", c.index, "from", c.start, "to", c.end, fmt.Sprintf("(~%d MB)", (c.end-c.start)/(1024*1024)))

	defer c.wg.Done()
	start := time.Now()

	if c.start >= c.end {
		return nil
	}

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
		c.logger.Print("Error downloading chunk:", err.Error())
		return err
	}

	elapsed := time.Since(start)
	c.logger.Print("Chunk", c.index, "downloaded in", elapsed.Seconds(), "s")

	return nil
}

func (c *chunk) Execute(ctx context.Context) error {
	return c.download(ctx)
}

func (c *chunk) OnError(ctx context.Context, err error) {
	if c.entry.Context().Err() != nil {
		return
	}

	var e error
	for i := 0; i < c.setting.MaxRetry(); i++ {
		c.wg.Add(1)
		c.logger.Print("Error downloading file:", err.Error(), ". Retrying...")

		if c.entry.Resumable() {
			c.start += resumePosition(c.path)
		}

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
	req, err := http.NewRequestWithContext(ctx, "GET", c.entry.URL(), nil)
	if err != nil {
		c.logger.Print("Error creating chunk request:", err.Error())
		return nil, err
	}

	bytesRange := fmt.Sprintf("bytes=%d-%d", c.start, c.end)
	req.Header.Add("Range", bytesRange)

	if entryCookie, ok := c.entry.(EntryCookies); ok && len(entryCookie.Cookies()) > 0 {
		for _, cookie := range entryCookie.Cookies() {
			req.AddCookie(cookie)
		}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.Print("Error fething chunk body:", err.Error())
		return nil, err
	}

	progressBar := &progressBar{
		onprogress: c.onprogress,
		reader:     res.Body,
		Entry:      c.entry,
		index:      c.index,
		downloaded: 0,
		progress:   0,
		chunkSize:  c.size,
	}

	return progressBar, nil
}

func (c *chunk) getSaveFile() (io.WriteCloser, error) {
	tmpFilename := filepath.Join(c.setting.DownloadLocation(), fmt.Sprintf("%s-%d", c.entry.ID(), c.index))
	file, err := os.OpenFile(tmpFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		c.logger.Print("Error creating or appending file:", err.Error())
		return nil, err
	}

	return file, nil
}
