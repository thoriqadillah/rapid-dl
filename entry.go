package rapid

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Entry interface {
	ID() string
	Name() string
	Location() string
	Size() int64
	Type() string  // document, compressed, audio, video, image, other, etc
	URL() string   // url which the entry downloaded from
	ChunkLen() int // total chunks splitted into
	Resumable() bool
	Context() context.Context
	Cancel() context.CancelFunc
	Expired() bool
	Refresh() error
}

type entry struct {
	id        string
	name      string
	location  string
	size      int64
	filetype  string
	url       string
	resumable bool
	chunkLen  int
	logger    Logger
	ctx       context.Context
	cancel    context.CancelFunc
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

func handleDuplicate(filename string) string {
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
		// add number before ext of a file if there is none
		split := strings.Split(name, ".")
		if len(split) > 2 {
			split[len(split)-2] += " (1)"
		} else {
			split[0] += " (1)"
		}

		// re-check if the current name has duplication
		name = strings.Join(split, ".")
		name = handleDuplicate(name)
		return name
	}

	// if it's still has, add the number
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

	// re-check if the current name has duplication
	name = strings.Join(split, ".")
	name = handleDuplicate(name)

	return name
}

func resumable(r *http.Response) bool {
	acceptRanges := r.Header.Get("Accept-Ranges")
	return acceptRanges != "" || acceptRanges == "bytes"
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

	return "file"
}

// calculatePartition calculates how many chunks will be for certain size
func calculatePartition(size int64, setting Setting) int {
	if size == -1 {
		return 1
	}

	if size < setting.MinChunkSize() {
		return 1
	}

	total := math.Log10(float64(size / (1024 * 1024)))
	partsize := setting.MinChunkSize()

	// dampening the total partition based on digit figures, e.g 100 -> 3 digits
	for i := 0; i < int(total); i++ {
		partsize *= int64(total)
	}

	return int(size / partsize)

}

func Fetch(url string, setting Setting) (Entry, error) {
	logger := NewLogger(setting)
	logger.Print("Fetching url...")

	req, err := http.Get(url)
	if err != nil {
		logger.Print("Error while fetching url", err.Error())
		return nil, err
	}

	resumable := resumable(req)
	filename := handleDuplicate(filename(req))
	location := filepath.Join(setting.DownloadLocation(), filename)
	filetype := filetype(filename)
	chunklen := calculatePartition(req.ContentLength, setting)
	ctx, cancel := context.WithCancel(context.Background())

	return &entry{
		id:        randID(10),
		name:      filename,
		location:  location,
		filetype:  filetype,
		url:       req.Request.URL.String(),
		size:      req.ContentLength,
		logger:    logger,
		chunkLen:  chunklen,
		ctx:       ctx,
		cancel:    cancel,
		resumable: resumable,
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

func (e *entry) ChunkLen() int {
	return e.chunkLen
}

func (e *entry) Resumable() bool {
	return e.resumable
}

func (e *entry) Context() context.Context {
	return e.ctx
}

func (e *entry) Cancel() context.CancelFunc {
	return e.cancel
}

func (e *entry) Expired() bool {
	resp, err := http.Head(e.url)
	if err != nil {
		e.logger.Print("Error while checking url expiration:", err.Error())
		return true
	}

	return resp.StatusCode != http.StatusOK && resp.ContentLength <= 0
}

func (e *entry) Refresh() error {
	if e.ctx.Err() != nil {
		return nil
	}

	e.ctx, e.cancel = context.WithCancel(context.Background())
	// TODO: do something else, such as refresh the link (future feature if browser extenstion is present)

	return nil
}

func (e *entry) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("ID: %v\n", e.id))
	buffer.WriteString(fmt.Sprintf("Name: %v\n", e.name))
	buffer.WriteString(fmt.Sprintf("Location: %v\n", e.location))
	buffer.WriteString(fmt.Sprintf("Size: %v\n", e.size))
	buffer.WriteString(fmt.Sprintf("Filetype: %v\n", e.filetype))
	buffer.WriteString(fmt.Sprintf("URL: %v\n", e.url))
	buffer.WriteString(fmt.Sprintf("Resumable: %v\n", e.resumable))
	buffer.WriteString(fmt.Sprintf("ChunkLen: %v\n", e.chunkLen))
	buffer.WriteString(fmt.Sprintf("Expired: %v\n", e.Expired()))

	return buffer.String()
}
