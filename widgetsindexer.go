package vivoupdater

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type WidgetsIndexer struct {
	Url      string
	Username string
	Password string
}

type WidgetsBatchIndexer struct {
	Indexer WidgetsIndexer
	Suffix  string
	Regex   *regexp.Regexp
	Uris    []string
}

func NewWidgetsBatchIndexer(wi WidgetsIndexer, suffix string, regex *regexp.Regexp) *WidgetsBatchIndexer {
	arr := make([]string, 0)
	return &WidgetsBatchIndexer{wi, suffix, regex, arr}
}

type WidgetsUpdateMessage struct {
	Uris []string `json:"uris"`
}

func (wi WidgetsIndexer) Name() string {
	return "WidgetsIndexer"

}

// This pulls out URLs the particular indexer might be interested in
// typically they would look like one of the following:
//
// https://scholars.duke.edu/individual/per3336942
// https://scholars.duke.edu/individual/org50344699
//
func (wbi *WidgetsBatchIndexer) Gather(u string) {
	if wbi.Regex.MatchString(u) {
		wbi.Uris = append(wbi.Uris, u)
	}
}

func (wbi WidgetsBatchIndexer) IndexUris(logger *log.Logger) error {
	size := len(wbi.Uris)

	if size <= 0 {
		return nil
	}

	m := WidgetsUpdateMessage{wbi.Uris}

	j, err := json.Marshal(m)
	if err != nil {
		return err
	}

	var data = url.Values{}
	data.Set("message", string(j))

	logger.Printf("%#v", wbi.Indexer.Url+wbi.Suffix)

	for _, uri := range wbi.Uris {
		logger.Printf("->%#v\n", uri)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", wbi.Indexer.Url+wbi.Suffix, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(wbi.Indexer.Username, wbi.Indexer.Password)

	resp, err := client.Do(req)

	resp.Body.Close()

	if err != nil {
		return err
	}

	return nil
}

/*
func (wi WidgetsIndexer) Index(batch map[string]bool) (map[string]bool, error) {

    logger := log.New(os.Stdout, "", log.LstdFlags)

    return wi.IndexLog(batch, logger)

}
*/

func (wi WidgetsIndexer) Index(batch map[string]bool, logger *log.Logger) (map[string]bool, error) {
	perRegx := regexp.MustCompile(`.*individual/per*`)
	orgRegx := regexp.MustCompile(`.*individual/org*`)

	widgetsPeopleIndexer := NewWidgetsBatchIndexer(wi, "/people/uris", perRegx)
	widgetsOrganizationIndexer := NewWidgetsBatchIndexer(wi, "/organizations/uris", orgRegx)

	for u := range batch {
		widgetsPeopleIndexer.Gather(u)
		widgetsOrganizationIndexer.Gather(u)
	}

	err := widgetsPeopleIndexer.IndexUris(logger)

	if err != nil {
		return batch, err
	}

	err = widgetsOrganizationIndexer.IndexUris(logger)

	if err != nil {
		return batch, err
	}

	return batch, nil
}