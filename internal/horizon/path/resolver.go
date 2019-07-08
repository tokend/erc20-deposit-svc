package path

import (
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/url"
)

type Resolver interface {
	URL (string) (string, error)
	URI(string, url.Values) (string, error)
}

type resolver struct {
	base *url.URL
}
//URL returns full path
func (r resolver) URL(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse endpoint into URL")
	}

	return r.base.ResolveReference(u).String(), nil
}
//URI returns path with query
func (r resolver) URI(endpoint string, values url.Values) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse endpoint into URL")
	}
	u.RawQuery = values.Encode()
	return u.RequestURI(), nil
}

func NewResolver(base *url.URL) Resolver {
	return &resolver{
		base: base,
	}
}

