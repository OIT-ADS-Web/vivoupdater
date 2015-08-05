package vivoupdater_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"vivoupdater"
)

func TestWidgetsPost(t *testing.T) {

	b := make(map[string]bool)
	b["http://a_uri"] = true
	b["http://b_uri"] = true

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		m := r.Form.Get("message")
		if m == "" {
			t.Fatal("message parameter not found")
		}

		var wum vivoupdater.WidgetsUpdateMessage
		json.Unmarshal([]byte(m), &wum)
		var sortUris sort.StringSlice = wum.Uris
		sortUris.Sort()
		if len(sortUris) != 2 || sortUris[0] != "http://a_uri" || sortUris[1] != "http://b_uri" {
			t.Errorf("expected 2 uris: http://a_uri, http://b_uri\ngot: %s", sortUris)
		}

		a := r.Header["Authorization"][0]
		if a == "" {
			t.Error("Authorization header not found")
		}
		exp_a := "Basic " + base64.StdEncoding.EncodeToString([]byte("testuser:testpassword"))
		if a != exp_a {
			t.Errorf("expected Authorization header to be: %s\ngot: %s", exp_a, a)
		}

	}))
	defer ts.Close()

	wi := vivoupdater.WidgetsIndexer{
		Url:      ts.URL,
		Username: "testuser",
		Password: "testpassword"}

	wi.Index(b)
}
