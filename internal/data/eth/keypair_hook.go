package eth

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure"
)

var KeypairHook = figure.Hooks{
	"eth.Keypair": func(raw interface{}) (reflect.Value, error) {
		switch value := raw.(type) {
		case string:
			kp, err := NewKeypair(value)
			if err != nil {
				return reflect.Value{}, errors.Wrap(err, "failed to init keypair")
			}
			return reflect.ValueOf(*kp), nil
		default:
			return reflect.Value{}, fmt.Errorf("cant init keypair from type: %T", value)
		}
	},
}
