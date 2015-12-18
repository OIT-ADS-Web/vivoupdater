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
/*

real world example uris:

https://scholars.duke.edu/individual/per3336942
https://scholars.duke.edu/individual/org50344699
*/

func TestWidgetsPost(t *testing.T) {

	b := make(map[string]bool)

	b["http://scholars/individual/per0000000"] = true
	b["http://scholars/individual/org0000000"] = true
	// something it doesn't expect - should ignore, right?
	b["http://domain.com/individual/gra000000"] = true

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
		
		if !((sortUris[0] == "http://scholars/individual/per0000000" && r.URL.Path == "/people/uris")  || 
		    (sortUris[0] == "http://scholars/individual/org0000000" && r.URL.Path == "/organizations/uris")) {
		    t.Errorf("expected http://scholars/individual/per0000000 to POST to /people/uris")
		    t.Errorf("OR expected http://scholars/individual/org0000000 to POST to /organizations/uris")
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

	wi.Index(b)
}
