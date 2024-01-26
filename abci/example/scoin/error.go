package scoin

import "fmt"

var (
	errUnMarshalableTx = fmt.Errorf("can not marshal scoin tx! bad tx")
	errInsertNoKey     = fmt.Errorf("sabci insert tx no key")
)
