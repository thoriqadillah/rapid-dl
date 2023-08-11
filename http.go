package rapid

import "net/http"

var httpMap = make(map[string]*http.Client)

func HttpClient(name ...string) *http.Client {
	provider := "default"
	if len(name) > 0 {
		provider = name[0]
	}

	client, ok := httpMap[provider]
	if ok {
		return client
	}

	return &http.Client{}
}

func RegisterHttpClient(name string, impl *http.Client) {
	httpMap[name] = impl
}
