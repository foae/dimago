package cacoo

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	apiKey := "foo"
	baseURL := "bar"
	folderID := "someFolder"

	c := NewClient(apiKey, baseURL, folderID)
	require.NotEmpty(t, c)
	require.NotEmpty(t, c.httpClient)
	require.Equal(t, apiKey, c.apiKey)
	require.Equal(t, baseURL, c.baseURL)
	require.Equal(t, folderID, c.folderID)
}

func TestCreateDiagram(t *testing.T) {
	apiKey := "foo"
	folderID := "someFolder"

	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = fmt.Fprintln(w, "{}")

			require.Equal(t, apiKey, r.URL.Query().Get("apiKey"))
			require.Equal(t, folderID, r.URL.Query().Get("folderId"))
		}))
	defer ts.Close()

	c := NewClient(apiKey, ts.URL, folderID)
	c.httpClient = ts.Client()

	err := c.CreateDiagram("New diagram", "some description")
	require.NoError(t, err)

}
