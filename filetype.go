package rapid

import "regexp"

type TypeExpression func() string

func imagetype() string {
	return `^.*.(jpg|jpeg|png|gif|svg|bmp)$`
}

func videotype() string {
	return `^.*\.(mp4|mov|avi|mkv|wmv|flv|webm|mpeg|mpg|3gp|m4v|m4a)$`
}

func audiotype() string {
	return `^.*.(mp3|wav|flac|aac|ogg|opus)$`
}

func documenttype() string {
	return `^.*.(doc|docx|pdf|txt|ppt|pptx|xls|xlsx|odt|ods|odp|odg|odf|rtf|tex|texi|texinfo|wpd|wps|wpg|wks|wqd|wqx|w)$`
}

func compressedtype() string {
	return `^.*.(zip|rar|7z|tar|gz|bz2|tgz|tbz2|xz|txz|zst|zstd)$`
}

var filetypeMap = make(map[string]TypeExpression)

func Filetype(filename string) string {
	for name, expr := range filetypeMap {
		if match, _ := regexp.MatchString(expr(), filename); match {
			return name
		}
	}

	return "Other"
}

func RegisterFiletype(name string, expr TypeExpression) {
	filetypeMap[name] = expr
}

func init() {
	RegisterFiletype("Image", imagetype)
	RegisterFiletype("Video", videotype)
	RegisterFiletype("Audio", audiotype)
	RegisterFiletype("Compressed", compressedtype)
	RegisterFiletype("Document", documenttype)
}
