package condition

import (
	"github.com/crud-bird/bfe/bfe_basic"
)

func Build(str string) (Condition, error) {
	// todo
	return &Tmp{}, nil
}

type Tmp struct{}

func (t *Tmp) Match(req *bfe_basic.Request) bool {
	return false
}
