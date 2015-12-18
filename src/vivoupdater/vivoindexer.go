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

func (wi VivoIndexer) Name() string {
	return "VivoIndexer"
}

// FIXME: is it really necessasary to implement, since we don't use?
//func (wi VivoIndexer) Filter (batch map[string]bool) (map[string]bool, error) {
//	return batch, nil
//}


func (vi VivoIndexer) Index(batch map[string]bool) (map[string]bool, error) {

	var buf bytes.Buffer
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

	for u := range batch {
		fw.Write([]byte(u))
		fw.Write([]byte(","))
	}
	w.Close()

	req, err := http.NewRequest("POST", vi.Url, &buf)
	if err != nil {
		return batch, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	resp.Body.Close()
	if err != nil {
		return batch, err
	}
	buf.Reset()
	return batch, nil
}
