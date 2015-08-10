package vivoupdater

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

type VivoIndexer struct {
	Url      string
	Username string
	Password string
}

func (vi VivoIndexer) Index(batch map[string]bool) (bool, error) {

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	err := w.WriteField("email", vi.Username)
	if err != nil {
		return false, err
	}
	err = w.WriteField("password", vi.Password)
	if err != nil {
		return false, err
	}

	mh := textproto.MIMEHeader{}
	mh.Add("Content-Disposition", "form-data; name=\"uris\"; filename=\"uriList.txt\"")
	mh.Add("Content-Type", "text/plain")

	fw, fwerr := w.CreatePart(mh)
	if fwerr != nil {
		return false, fwerr
	}

	for u := range batch {
		fw.Write([]byte(u))
		fw.Write([]byte(","))
	}
	w.Close()

	req, err := http.NewRequest("POST", vi.Url, &buf)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	resp.Body.Close()
	if err != nil {
		return false, err
	}
	buf.Reset()
	return true, nil
}
