package submit

import (
	"errors"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestTxFailure_GetLoganFields(t *testing.T) {
	txFailure := TxFailure{
		error:                 errors.New("tx malformed"),
		ResultXDR:             "AAAAAAAAAAKEK==",
		TransactionResultCode: "tx_malformed",
		OperationResultCodes:  []string{"op_kek_is_malformed", "op_lol_is_malformed"},
	}
	fileds := txFailure.GetLoganFields()
	assert.Equal(t, fileds["operation_result_codes"], "[op_kek_is_malformed op_lol_is_malformed]")
}
