package watchlist

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-ozzo/ozzo-validation"
	regources "gitlab.com/tokend/regources/generated"
)

//AssetDetails contain details about asset that can be deposited using service
type AssetDetails struct {
	ExternalSystemType int32 `json:"external_system_type,string"`
	ERC20              struct {
		Deposit bool           `json:"deposit"`
		Address common.Address `json:"address"`
	} `json:"erc20"`
}

//Validate validates asset details
func (s AssetDetails) Validate() error {
	errs := validation.Errors{
		"ExternalSystemType": validation.Validate(&s.ExternalSystemType, validation.Required, validation.Min(1)),
		"Deposit":            validation.Validate(&s.ERC20.Deposit, validation.Required),
		"Address":            validation.Validate(&s.ERC20.Address, validation.Required),
	}

	return errs.Filter()
}

// Details is a composition structure which contain asset resource and it's details
type Details struct {
	regources.Asset
	AssetDetails
}
