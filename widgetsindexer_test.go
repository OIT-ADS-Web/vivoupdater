package vivoupdater_test

import (
	"encoding/base64"
	"encoding/json"
	"github.com/OIT-ADS-Web/vivoupdater"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"
)

/*

real world example uris:

https://scholars.duke.edu/individual/per3336942
https://scholars.duke.edu/individual/org50344699
*/

func TestWidgetsPost(t *testing.T) {
	b := make(map[string]bool)

	// NOTE: 7 digits for person, 8 digits for org
	b["http://scholars/individual/per0000001"] = true
	b["http://scholars/individual/org00000001"] = true

	b["http://scholars/individual/pernet1"] = true
	b["http://scholars/individual/pernet2"] = true

	b["http://domain.com/individual/gra0000001"] = true
	b["http://domain.com/individual/per_addr000000"] = true

	b["http://domain.com/individual/per1"] = true
	b["http://domain.com/individual/org1"] = true


	logger := log.New(os.Stdout, "", log.LstdFlags)

	// FIXME: the test should really accumulate all requests over time, and compare
	// against expected, but I'm not sure how to do that
	// all_requests := make(map[string]string)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		r.ParseForm()
		m := r.Form.Get("message")
		if m == "" {
			t.Fatal("message parameter not found")
		}

		a := r.Header["Authorization"][0]
		if a == "" {
			t.Error("Authorization header not found")
		}
		exp_a := "Basic " + base64.StdEncoding.EncodeToString([]byte("testuser:testpassword"))
		if a != exp_a {
			t.Errorf("expected Authorization header to be: %s\ngot: %s", exp_a, a)
		}

		var wum vivoupdater.WidgetsUpdateMessage
		json.Unmarshal([]byte(m), &wum)

		var sortUris sort.StringSlice = wum.Uris
		sortUris.Sort()

		if (sortUris[0] == "http://domain.com/individual/per1" || sortUris[0] == "http://domain.com/individual/org1" ||
		    sortUris[0] == "http://domain.com/indivudal/gra0000001" || sortUris[0] == "http://domain.com/individual/per_addr000000") {
			t.Errorf("Not supposed to allow %s", sortUris)
		}

		if !((sortUris[0] == "http://scholars/individual/per0000001" && r.URL.Path == "/people/uris") ||
			(sortUris[0] == "http://scholars/individual/org00000001" && r.URL.Path == "/organizations/uris")) ||
			((sortUris[0] == "http://scholars/individual/pernet1" && r.URL.Path == "/people/uris") ||
		         (sortUris[0] == "http://scholars/individual/pernet2" && r.URL.Path == "/people/uris")) {

			t.Errorf("expected http://scholars/individual/per0000001 to POST to /people/uris")
			t.Errorf("OR expected http://scholars/individual/org00000001 to POST to /organizations/uris")
			t.Errorf("not %s to %s", sortUris, r.URL.Path)

		} else {
			t.Logf("%s will route to %s", sortUris, r.URL.Path)
		}

	}))
	defer ts.Close()

	wi := vivoupdater.WidgetsIndexer{
		Url:      ts.URL,
		Username: "testuser",
		Password: "testpassword"}

	wi.Index(b, logger)
}
