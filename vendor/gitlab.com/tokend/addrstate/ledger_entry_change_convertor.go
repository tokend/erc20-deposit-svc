package addrstate

import (
	"gitlab.com/tokend/regources/generated"
	"gitlab.com/tokend/go/xdr"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

var ErrUnexpectedEffect = errors.New("unexpected change effect")

func convertLedgerEntryChange(change regources.LedgerEntryChange) (xdr.LedgerEntryChange, error) {
	var ledgerEntryChange xdr.LedgerEntryChange
	err := xdr.SafeUnmarshalBase64(change.Attributes.Payload, &ledgerEntryChange)
	if err != nil {
		return xdr.LedgerEntryChange{}, errors.Wrap(err, "failed to unmarshal ledger entry", logan.F{
			"xdr" : change.Attributes.Payload,
		})
	}
	return ledgerEntryChange, nil
}
