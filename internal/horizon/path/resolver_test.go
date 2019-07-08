package path

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestResolver_URL(t *testing.T) {
	baseURL, _ := url.Parse("http://google.com")
	r := NewResolver(baseURL)
	t.Run("path", func(t *testing.T) {
		resolved, err := r.URL("/search")
		assert.NoError(t, err)
		assert.Equal(t, "http://google.com/search", resolved)
	})

	t.Run("query", func(t *testing.T) {
		resolved, err := r.URL("?a=1&b=2")
		assert.NoError(t, err)
		assert.Equal(t, "http://google.com?a=1&b=2", resolved)
	})

	t.Run("path & query", func(t *testing.T) {
		resolved, err := r.URL("/search?a=1&b=2")
		assert.NoError(t, err)
		assert.Equal(t, "http://google.com/search?a=1&b=2", resolved)
	})

}

func TestResolver_URI(t *testing.T) {
	baseURL, _ := url.Parse("http://google.com")
	r := NewResolver(baseURL)
	values := url.Values{}
	values.Add("a", "1")
	values.Add("b", "2")
	values.Add("c", "3")
	t.Run("with path", func(t *testing.T) {
		resolved, err := r.URI("/search", values)
		assert.NoError(t, err)
		assert.Equal(t, "/search?a=1&b=2&c=3", resolved)
	})

	t.Run("bare query", func(t *testing.T) {
		resolved, err := r.URI("", values)
		assert.NoError(t, err)
		assert.Equal(t, "/?a=1&b=2&c=3", resolved)
	})

}
