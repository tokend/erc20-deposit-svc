package getters

import (
	"bytes"
	"encoding/json"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/client"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/query"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

//go:generate genny -in=getter.tmpl -out=asset_getter.go gen "Template=Asset Resource=Asset"
//go:generate genny -in=getter.tmpl -out=transaction_getter.go gen "Template=Transaction Resource=Transaction"
//go:generate genny -in=getter.tmpl -out=issuance_getter.go gen "Template=ReviewableRequest Resource=CreateIssuanceRequest"

type Getter interface {
	GetPage(endpoint string, params query.Params, result interface{}) error
	PageFromLink(link string, v interface{}) error
}

type getter struct {
	*client.Client
}

func New(client *client.Client) *getter {
	return &getter{Client: client}
}

func (g *getter) PageFromLink(link string, v interface{}) error {
	resp, err := g.Get(link)
	if err != nil {
		return errors.Wrap(err, "failed to get page")
	}

	response := bytes.NewReader(resp)
	decoder := json.NewDecoder(response)

	err = decoder.Decode(v)
	if err != nil {
		return errors.Wrap(err, "failed to parse response")
	}

	return nil
}

func (g *getter) GetPage(endpoint string, params query.Params, result interface{}) error {
	q := params.Prepare()
	uri, err := g.Resolve().URI(endpoint, q)
	if err != nil {
		return errors.Wrap(err, "failed to resolve request URI", logan.F{
			"endpoint": endpoint,
			"query":    params,
		})
	}
	resp, err := g.Get(uri)
	if err != nil {
		return errors.Wrap(err, "failed to perform request")
	}
	response := bytes.NewReader(resp)
	decoder := json.NewDecoder(response)
	err = decoder.Decode(result)
	if err != nil {
		return errors.Wrap(err, "failed to parse response")
	}
	return nil
}
