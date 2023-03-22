package keeper

import (
	"github.com/bernardstanislas/checkers/x/checkers/types"
)

var _ types.QueryServer = Keeper{}
