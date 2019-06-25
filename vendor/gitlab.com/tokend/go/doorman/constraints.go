package doorman

import (
	"gitlab.com/tokend/go/resources"
	"net/http"

	"gitlab.com/tokend/go/signcontrol"
)

type SignerOfExt interface {
	Check(signer resources.Signer) bool
}

func SignerOf(address string, ext ...SignerOfExt) SignerConstraint {
	return func(r *http.Request, doorman Doorman) error {
		signer, err := signcontrol.CheckSignature(r)
		if err != nil {
			return err
		}

		signers, err := doorman.AccountSigners(address)
		if err != nil {
			return err
		}

		for _, accountSigner := range signers {
			if accountSigner.AccountID == signer && accountSigner.Weight > 0 {
				return nil
			}

			if len(ext) == 0 {
				ext = doorman.DefaultSignerOfConstraints()
			}

			if checkConstraints(accountSigner, ext) {
				return nil
			}
		}
		return ErrNotAllowed
	}
}

func SignatureOf(address string) SignerConstraint {
	return func(r *http.Request, doorman Doorman) error {
		signer, err := signcontrol.CheckSignature(r)
		if err != nil {
			return err
		}

		if signer == address {
			return nil
		}

		return ErrNotAllowed
	}
}

func checkConstraints(accountSigner resources.Signer, constraints []SignerOfExt) bool {
	for _, c := range constraints {
		if !c.Check(accountSigner) {
			return false
		}
	}
	return true
}
