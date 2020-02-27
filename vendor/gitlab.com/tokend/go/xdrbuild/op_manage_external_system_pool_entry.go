package xdrbuild

import (
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/tokend/go/xdr"
)

type ManageExternalPoolEntryOp struct {
	Action xdr.ManageExternalSystemAccountIdPoolEntryAction
	Create *createExternalPoolEntryInput
	Remove *uint64
}

func (op ManageExternalPoolEntryOp) XDR() (*xdr.Operation, error) {
	mop := xdr.ManageExternalSystemAccountIdPoolEntryOp{
		ActionInput: xdr.ManageExternalSystemAccountIdPoolEntryOpActionInput{
			Action: op.Action,
		},
	}

	switch op.Action {
	case xdr.ManageExternalSystemAccountIdPoolEntryActionCreate:
		mop.ActionInput.CreateExternalSystemAccountIdPoolEntryActionInput = &xdr.CreateExternalSystemAccountIdPoolEntryActionInput{
			ExternalSystemType: xdr.Int32(op.Create.ExternalSystemType),
			Data:               xdr.Longstring(op.Create.Data),
			Parent:             xdr.Uint64(op.Create.Parent),
		}
	case xdr.ManageExternalSystemAccountIdPoolEntryActionRemove:
		mop.ActionInput.DeleteExternalSystemAccountIdPoolEntryActionInput = &xdr.DeleteExternalSystemAccountIdPoolEntryActionInput{
			PoolEntryId: xdr.Uint64(*op.Remove),
		}
	default:
		return nil, errors.New("unexpected action")
	}

	return &xdr.Operation{
		Body: xdr.OperationBody{
			Type:                                     xdr.OperationTypeManageExternalSystemAccountIdPoolEntry,
			ManageExternalSystemAccountIdPoolEntryOp: &mop,
		},
	}, nil
}

type createExternalPoolEntryInput struct {
	ExternalSystemType int32
	Data               string
	Parent             uint64
}

func CreateExternalPoolEntry(externalSystemType int32, data string, parent uint64) ManageExternalPoolEntryOp {
	return ManageExternalPoolEntryOp{
		Action: xdr.ManageExternalSystemAccountIdPoolEntryActionCreate,
		Create: &createExternalPoolEntryInput{
			ExternalSystemType: externalSystemType,
			Data:               data,
			Parent:             parent,
		},
	}
}

func RemoveExternalPoolEntry(id uint64) ManageExternalPoolEntryOp {
	return ManageExternalPoolEntryOp{
		Action: xdr.ManageExternalSystemAccountIdPoolEntryActionRemove,
		Remove: &id,
	}
}
