// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package getters

import (
	"github.com/tokend/erc20-deposit-svc/internal/horizon/client"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/page"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/query"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	regources "gitlab.com/tokend/regources/generated"
)

type CreateIssuanceRequestPager interface {
	Next() (*regources.ReviewableRequestListResponse, error)
	Prev() (*regources.ReviewableRequestListResponse, error)
	Self() (*regources.ReviewableRequestListResponse, error)
	First() (*regources.ReviewableRequestListResponse, error)
}

type CreateIssuanceRequestGetter interface {
	SetFilters(filters query.CreateIssuanceRequestFilters)
	SetIncludes(includes query.CreateIssuanceRequestIncludes)
	SetPageParams(pageParams page.Params)
	SetParams(params query.CreateIssuanceRequestParams)

	Filter() query.CreateIssuanceRequestFilters
	Include() query.CreateIssuanceRequestIncludes
	Page() page.Params

	ByID(ID string) (*regources.ReviewableRequestResponse, error)
	List() (*regources.ReviewableRequestListResponse, error)
}

type CreateIssuanceRequestHandler interface {
	CreateIssuanceRequestGetter
	CreateIssuanceRequestPager
}

type defaultCreateIssuanceRequestHandler struct {
	base   Getter
	params query.CreateIssuanceRequestParams

	currentPageLinks *regources.Links
}

func NewDefaultCreateIssuanceRequestHandler(c *client.Client) *defaultCreateIssuanceRequestHandler {
	return &defaultCreateIssuanceRequestHandler{
		base: New(c),
	}
}

func (g *defaultCreateIssuanceRequestHandler) SetFilters(filters query.CreateIssuanceRequestFilters) {
	g.params.Filters = filters
}

func (g *defaultCreateIssuanceRequestHandler) SetIncludes(includes query.CreateIssuanceRequestIncludes) {
	g.params.Includes = includes
}

func (g *defaultCreateIssuanceRequestHandler) SetPageParams(pageParams page.Params) {
	g.params.PageParams = pageParams
}

func (g *defaultCreateIssuanceRequestHandler) SetParams(params query.CreateIssuanceRequestParams) {
	g.params = params
}

func (g *defaultCreateIssuanceRequestHandler) Params() query.CreateIssuanceRequestParams {
	return g.params
}

func (g *defaultCreateIssuanceRequestHandler) Filter() query.CreateIssuanceRequestFilters {
	return g.params.Filters
}

func (g *defaultCreateIssuanceRequestHandler) Include() query.CreateIssuanceRequestIncludes {
	return g.params.Includes
}

func (g *defaultCreateIssuanceRequestHandler) Page() page.Params {
	return g.params.PageParams
}

func (g *defaultCreateIssuanceRequestHandler) ByID(ID string) (*regources.ReviewableRequestResponse, error) {
	result := &regources.ReviewableRequestResponse{}
	err := g.base.GetPage(query.CreateIssuanceRequestByID(ID), g.params.Includes, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get record by id", logan.F{
			"id": ID,
		})
	}
	return result, nil
}

func (g *defaultCreateIssuanceRequestHandler) List() (*regources.ReviewableRequestListResponse, error) {
	result := &regources.ReviewableRequestListResponse{}
	err := g.base.GetPage(query.CreateIssuanceRequestList(), g.params, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get records list", logan.F{
			"query_params": g.params,
		})
	}
	g.currentPageLinks = result.Links
	return result, nil
}

func (g *defaultCreateIssuanceRequestHandler) Next() (*regources.ReviewableRequestListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Next == "" {
		return nil, errors.From(errors.New("No link to page"), logan.F{
			"page": "next",
		})
	}
	result := &regources.ReviewableRequestListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.Next, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get next page", logan.F{
			"link": g.currentPageLinks.Next,
		})
	}
	g.currentPageLinks = result.Links

	return result, nil
}

func (g *defaultCreateIssuanceRequestHandler) Prev() (*regources.ReviewableRequestListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Prev == "" {
		return nil, errors.From(errors.New("No link to page"), logan.F{
			"page": "prev",
		})
	}

	result := &regources.ReviewableRequestListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.Prev, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get previous page", logan.F{
			"link": g.currentPageLinks.Prev,
		})
	}
	g.currentPageLinks = result.Links

	return result, nil
}

func (g *defaultCreateIssuanceRequestHandler) Self() (*regources.ReviewableRequestListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Self == "" {
		return nil, errors.From(errors.New("No link to page"), logan.F{
			"page": "self",
		})
	}
	result := &regources.ReviewableRequestListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.Self, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get same page", logan.F{
			"link": g.currentPageLinks.Self,
		})
	}
	g.currentPageLinks = result.Links

	return result, nil
}

func (g *defaultCreateIssuanceRequestHandler) First() (*regources.ReviewableRequestListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.First == "" {
		return nil, errors.From(errors.New("No link to page"), logan.F{
			"page": "first",
		})
	}
	result := &regources.ReviewableRequestListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.First, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get first page", logan.F{
			"link": g.currentPageLinks.First,
		})
	}
	g.currentPageLinks = result.Links

	return result, nil
}