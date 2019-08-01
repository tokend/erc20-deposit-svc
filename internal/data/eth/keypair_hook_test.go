package eth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/distributed_lab/figure"
)

func TestKeypairHook(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			"89eaf47f681233c48a0be1f4d5da3fadd19b1a5e53d2896e78034cf6421249c6",
			"0x36A679C4C7c3B2D556a4D78caefAaaB2D25D273E",
		},
		{
			"0x4ea5b1977e31df777dcb244c237fe7f956e0569cb0468e96f16e8004edf1fecd",
			"0x3E9713D2fe14e6a34cD5192240e982Bdd7c249c7",
		},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			var config struct {
				Keypair Keypair
			}
			err := figure.Out(&config).From(map[string]interface{}{
				"keypair": tc.input,
			}).With(KeypairHook).Please()
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, config.Keypair.Address().Hex())
		})
	}
}
