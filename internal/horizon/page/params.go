package page

import (
	"net/url"

	. "github.com/go-ozzo/ozzo-validation"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Params struct {
	Number *string `page:"number"`
	Limit  *string `page:"limit"`
	Cursor *string `page:"cursor"`
	Order  *string `page:"order"`
}

func isNil(i interface{}) error {
	_, isNil := Indirect(i)
	if !isNil {
		return errors.New("must be nil")
	}
	return nil
}

func (p Params) Validate() error {
	errs := Errors{
		"Limit": Validate(&p.Limit, NilOrNotEmpty),
		"Order": Validate(&p.Order, NilOrNotEmpty),
	}

	if p.Cursor != nil {
		errs["Cursor"] = Validate(p.Cursor, NotNil)
		errs["Number"] = Validate(p.Number, By(isNil))
	}

	if p.Number != nil {
		errs["Cursor"] = Validate(p.Cursor, By(isNil))
		errs["Number"] = Validate(&p.Number, NilOrNotEmpty)
	}
	return errs.Filter()
}

func (p Params) Prepare(result *url.Values) {

	if p.Number != nil {
		result.Add("page[number]", *p.Number)
	}
	if p.Limit != nil {
		result.Add("page[limit]", *p.Limit)
	}
	if p.Order != nil {
		result.Add("page[order]", *p.Order)
	}

	if p.Cursor != nil {
		result.Add("page[cursor]", *p.Cursor)
	}

}
