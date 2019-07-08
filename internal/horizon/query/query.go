package query

import "net/url"

type Params interface {
	Prepare() url.Values
}