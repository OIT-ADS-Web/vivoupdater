package vivoupdater

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type WidgetsIndexer struct {
	Url      string
	Username string
	Password string
}

type WidgetsUpdateMessage struct {
	Uris []string `json:"uris"`
}

func (wi WidgetsIndexer) Name() string {
	return "WidgetsIndexer"
}

func (wi WidgetsIndexer) Index(batch map[string]bool) (map[string]bool, error) {
	i := 0
	uris := make([]string, len(batch))
	for u := range batch {
		uris[i] = u
		i++
	}
	m := WidgetsUpdateMessage{uris}
	j, err := json.Marshal(m)
	if err != nil {
		return batch, err
	}
	var data = url.Values{}
	data.Set("message", string(j))

	client := &http.Client{}
	req, err := http.NewRequest("POST", wi.Url, strings.NewReader(data.Encode()))
	if err != nil {
		return batch, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(wi.Username, wi.Password)
	resp, err := client.Do(req)
	resp.Body.Close()
	if err != nil {
		return batch, err
	}
	return batch, nil
}
