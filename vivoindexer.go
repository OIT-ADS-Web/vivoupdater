package vivoupdater

import (
	"bytes"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"
)

type VivoIndexer struct {
	Url      string
	Username string
	Password string
	Metrics  bool
}

func (wi VivoIndexer) Name() string {
	return "VivoIndexer"
}

func PostToVivo(idx VivoIndexer, batch map[string]bool) error {
	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)
	err := w.WriteField("email", idx.Username)
	if err != nil {
		return err
	}
	err = w.WriteField("password", idx.Password)
	if err != nil {
		return err
	}

	mh := textproto.MIMEHeader{}
	mh.Add("Content-Disposition", "form-data; name=\"uris\"; filename=\"uriList.txt\"")
	mh.Add("Content-Type", "text/plain")

	fw, err := w.CreatePart(mh)
	if err != nil {
		return err
	}

	// if there is a low-level Write error, just return?
	for u := range batch {
		_, err = fw.Write([]byte(u))
		if err != nil {
			return err
		}
		_, err = fw.Write([]byte(","))
		if err != nil {
			return err
		}
	}
	w.Close()

	req, err := http.NewRequest("POST", idx.Url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)

	//stackoverflow.com/questions/16280176/go-panic-runtime-error-invalid-memory-address-or-nil-pointer-dereference
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf.Reset()
	return nil
}

func (vi VivoIndexer) Index(batch map[string]bool, logger *log.Logger) (map[string]bool, error) {
	start := time.Now()
	err := PostToVivo(vi, batch)

	logger.Printf("vivo-indexer-url:%#v", vi.Url)

	var uris []string
	for k, _ := range batch {
		uris = append(uris, k)
	}

	for _, uri := range uris {
		logger.Printf("->vivo-index:%#v\n", uri)
	}

	end := time.Now()
	metrics := IndexMetrics{Start: start,
		End:     end,
		Uris:    uris,
		Name:    "vivo",
		Success: (err == nil)}

	if vi.Metrics != false {
		SendMetrics(metrics)
	}

	// TODO: if fail, add to a topic or even back to same topic
	// with attempt #numbers tracked
	// if err != nil {
	//
	//}
	return batch, err
}
