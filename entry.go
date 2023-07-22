package rapid

import (
	"mime"
	"net/http"
	"path/filepath"
	"regexp"
	"time"
)

type Entry interface {
	Name() string
	Location() string
	Size() int64
	Type() string    // document, compressed, audio, video, image, other, etc
	URL() string     // url which the entry downloaded from
	Date() time.Time // date created
}

type entry struct {
	name     string
	location string
	size     int64
	filetype string
	url      string
	date     time.Time
	setting  Setting
	logger   Logger
}

var (
	imagetype      = `^.*.(jpg|jpeg|png|gif|svg|bmp)$`
	videotype      = `^.*\.(mp4|mov|avi|mkv|wmv|flv|webm|mpeg|mpg|3gp|m4v|m4a)$`
	audiotype      = `^.*.(mp3|wav|flac|aac|ogg|opus)$`
	documenttype   = `^.*.(doc|docx|pdf|txt|ppt|pptx|xls|xlsx|odt|ods|odp|odg|odf|rtf|tex|texi|texinfo|wpd|wps|wpg|wks|wqd|wqx|w)$`
	compressedtype = `^.*.(zip|rar|7z|tar|gz|bz2|tgz|tbz2|xz|txz|zst|zstd)$`
)

func filetype(name string) string {
	if match, _ := regexp.MatchString(imagetype, name); match {
		return "image"
	}
	if match, _ := regexp.MatchString(videotype, name); match {
		return "video"
	}
	if match, _ := regexp.MatchString(audiotype, name); match {
		return "audio"
	}
	if match, _ := regexp.MatchString(documenttype, name); match {
		return "document"
	}
	if match, _ := regexp.MatchString(compressedtype, name); match {
		return "compressed"
	}

	return "other"
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

	return &entry{
		name:     filename,
		location: location,
		filetype: filetype,
		url:      req.Request.URL.String(),
		size:     req.ContentLength,
		setting:  setting,
		date:     time.Now(),
		logger:   logger,
	}, nil
}

func filename(r *http.Response) string {
	disposition := r.Header.Get("Content-Disposition")
	_, params, _ := mime.ParseMediaType(disposition)

	return params["filename"]
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
