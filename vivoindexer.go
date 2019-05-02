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
}

func (wi VivoIndexer) Name() string {
	return "VivoIndexer"
}

func (vi VivoIndexer) Index(batch map[string]bool, logger *log.Logger) (map[string]bool, error) {
	var buf bytes.Buffer
	start := time.Now()

	w := multipart.NewWriter(&buf)
	err := w.WriteField("email", vi.Username)
	if err != nil {
		return batch, err
	}
	err = w.WriteField("password", vi.Password)
	if err != nil {
		return batch, err
	}

	mh := textproto.MIMEHeader{}
	mh.Add("Content-Disposition", "form-data; name=\"uris\"; filename=\"uriList.txt\"")
	mh.Add("Content-Type", "text/plain")

	fw, fwerr := w.CreatePart(mh)
	if fwerr != nil {
		return batch, fwerr
	}

	// if there is a low-level Write error, just return?
	for u := range batch {
		_, err = fw.Write([]byte(u))
		if err != nil {
			return batch, err
		}
		_, err = fw.Write([]byte(","))
		if err != nil {
			return batch, err
		}
	}
	w.Close()

	req, err := http.NewRequest("POST", vi.Url, &buf)
	if err != nil {
		return batch, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)

	//stackoverflow.com/questions/16280176/go-panic-runtime-error-invalid-memory-address-or-nil-pointer-dereference
	if err != nil {
		return batch, err
	}

	uris := make([]string, len(batch))
	for k, _ := range batch {
		uris = append(uris, k)
	}

	logger.Printf("vivo-indexer-url:%#v", vi.Url)

	// filter out "" here? how are they making it in?
	for _, uri := range uris {
		logger.Printf("->vivo-index:%#v\n", uri)
	}

	end := time.Now()
	metrics := IndexMetrics{Start: start, End: end, Uris: uris, Name: "vivo"}
	SendMetrics(metrics, logger)

	defer resp.Body.Close()
	buf.Reset()
	return batch, nil
}
