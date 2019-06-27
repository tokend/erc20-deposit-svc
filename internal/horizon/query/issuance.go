package query

import (
	"fmt"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/page"
	"net/url"
)

type CreateIssuanceRequestFilters struct {
	ReviewableRequestFilters
	Receiver *string
	Asset    *string
}

type CreateIssuanceRequestIncludes struct {
	ReviewableRequestIncludes
	Balance bool
	Asset   bool
}

type CreateIssuanceRequestParams struct {
	Includes   CreateIssuanceRequestIncludes
	Filters    CreateIssuanceRequestFilters
	PageParams page.Params
}

func (p CreateIssuanceRequestParams) Prepare() url.Values {
	result := url.Values{}
	p.Filters.prepare(&result)
	p.PageParams.Prepare(&result)
	p.Includes.prepare(&result)
	return result
}

func (p CreateIssuanceRequestFilters) prepare(result *url.Values) {
	p.ReviewableRequestFilters.prepare(result)
	if p.Receiver != nil {
		result.Add("filter[request_details.receiver]", fmt.Sprintf("%s", *p.Receiver))
	}
	if p.Asset != nil {
		result.Add("filter[request_details.asset]", fmt.Sprintf("%s", *p.Asset))
	}
}

func (p CreateIssuanceRequestIncludes) prepare(result *url.Values) {
	p.ReviewableRequestIncludes.prepare(result)

	if p.Asset {
		result.Add("include", "request_details.asset")
	}

	if p.Balance {
		result.Add("include", "request_details.balance")
	}
}

func (p CreateIssuanceRequestIncludes) Prepare() url.Values {
	result := url.Values{}
	p.prepare(&result)
	return result
}

func CreateIssuanceRequestByID(code string) string {
	return fmt.Sprintf("/v3/create_issuance_requests/%s", code)
}

func CreateIssuanceRequestList() string {
	return "/v3/create_issuance_requests"
}
