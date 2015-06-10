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

func (vi VivoIndexer) Index(ctx Context, batches chan map[string]bool) {
	for b := range batches {

		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		err := w.WriteField("email", vi.Username)
		if err != nil {
			ctx.handleError("VivoIndexer: Could not write email field", err, true)
			break
		}
		err = w.WriteField("password", vi.Password)
		if err != nil {
			ctx.handleError("VivoIndexer: Could not write password field", err, true)
			break
		}

		mh := textproto.MIMEHeader{}
		mh.Add("Content-Disposition", "form-data; name=\"uris\"; filename=\"uriList.txt\"")
		mh.Add("Content-Type", "text/plain")

		fw, fwerr := w.CreatePart(mh)
		if fwerr != nil {
			ctx.handleError("VivoIndexer: Could not create form part for uris", err, true)
		}

		for u := range b {
			fw.Write([]byte(u))
			fw.Write([]byte(","))
		}
		w.Close()

		req, err := http.NewRequest("POST", vi.Url, &buf)
		if err != nil {
			ctx.handleError("VivoIndexer: Could not create index request", err, true)
			break
		}
		req.Header.Set("Content-Type", w.FormDataContentType())
		client := &http.Client{}
		resp, err := client.Do(req)
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			ctx.handleError("VivoIndexer: Could not post index request to Vivo", err, true)
			break
		}
		ctx.Logger.Printf("%v uris sent to Vivo for indexing\n", len(b))
		buf.Reset()
	}
}
