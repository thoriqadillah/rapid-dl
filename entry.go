package rapid

import (
	"math"
	"math/rand"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type Entry interface {
	ID() string
	Name() string
	Location() string
	Size() int64
	Type() string    // document, compressed, audio, video, image, other, etc
	URL() string     // url which the entry downloaded from
	Date() time.Time // date created
	ChunkLen() int   // total chunks splitted into
}

type entry struct {
	id       string
	name     string
	location string
	size     int64
	filetype string
	url      string
	date     time.Time
	logger   Logger
	chunkLen int
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randID(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)

	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func filename(r *http.Response) string {
	disposition := r.Header.Get("Content-Disposition")
	_, params, _ := mime.ParseMediaType(disposition)

	filename, ok := params["filename"]
	if ok {
		return filename
	}

	urlPath := r.Request.URL.Path
	if i := strings.LastIndex(urlPath, "/"); i != -1 {
		return urlPath[i+1:]
	}

	// If the filename cannot be determined from the header or URL, return an empty string
	return "empty"
}

// calculatePartition calculates how many chunks will be for certain size
func calculatePartition(size int64, setting Setting) int {
	if size < setting.MinChunkSize() {
		return 1
	}

	total := math.Log10(float64(size / (1024 * 1024)))
	partsize := setting.MinChunkSize()

	// dampening the total partition based on digit figures, e.g 100 -> 3 digits
	for i := 0; i < int(total); i++ {
		partsize = int64(total) + 1
	}

	return int(size / partsize)

}

func Fetch(url string, setting Setting) (Entry, error) {
	logger := NewLogger(setting.LoggerProvider(), setting)
	logger.Print("Fetching url...")

	req, err := http.Get(url)
	if err != nil {
		logger.Print("Error while fetching url", err.Error())
		return nil, err
	}

	filename := filename(req)
	location := filepath.Join(setting.DownloadLocation(), filename)
	filetype := filetype(filename)
	chunklen := calculatePartition(req.ContentLength, setting)

	return &entry{
		id:       randID(5),
		name:     filename,
		location: location,
		filetype: filetype,
		url:      req.Request.URL.String(),
		size:     req.ContentLength,
		date:     time.Now(),
		logger:   logger,
		chunkLen: chunklen,
	}, nil
}

func (e *entry) ID() string {
	return e.id
}

func (e *entry) Name() string {
	return e.name
}

func (e *entry) Location() string {
	return e.location
}

func (e *entry) Size() int64 {
	return e.size
}

func (e *entry) Type() string {
	return e.filetype
}

func (e *entry) URL() string {
	return e.url
}

func (e *entry) Date() time.Time {
	return e.date
}

func (e *entry) ChunkLen() int {
	return e.chunkLen
}
