package query

import (
	"fmt"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/page"
	"net/url"
	"strings"
)

func TransactionList() string {
	return "/v3/transactions"
}

func TransactionByID(id string) string {
	return fmt.Sprintf("/v3/transactions/%s", id)
}

type TransactionFilters struct {
	ChangeTypes []int
	EntryTypes  []int
}

type TransactionIncludes struct {
	LedgerEntryChanges bool
}

func (p TransactionIncludes) Prepare() url.Values {
	result := url.Values{}
	p.prepare(&result)
	return result
}

type TransactionParams struct {
	Includes   TransactionIncludes
	Filters    TransactionFilters
	PageParams page.Params
}

func (p TransactionParams) Prepare() url.Values {
	result := url.Values{}
	p.Filters.prepare(&result)
	p.PageParams.Prepare(&result)
	p.Includes.prepare(&result)
	return result
}

func (p TransactionFilters) prepare(result *url.Values) {
	if p.EntryTypes != nil {
		types := make([]string, len(p.EntryTypes))
		for _, et := range p.EntryTypes {
			types = append(types, fmt.Sprintf("%d", et))
		}
		result.Add("filter[ledger_entry_changes.entry_types]", strings.Join(types, ","))
	}

	if p.ChangeTypes != nil {
		types := make([]string, len(p.ChangeTypes))
		for _, ct := range p.ChangeTypes {
			types = append(types, fmt.Sprintf("%d", ct))
		}
		result.Add("filter[ledger_entry_changes.change_types]", strings.Join(types, ","))
	}
}

func (p TransactionIncludes) prepare(result *url.Values) {
	if p.LedgerEntryChanges {
		result.Add("include", "ledger_entry_changes")
	}
}
