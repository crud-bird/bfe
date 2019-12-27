package condition

import (
	"github.com/crud-bird/bfe/bfe_basic"
)

type Condition interface {
	Match(req *bfe_basic.Request) bool
}
