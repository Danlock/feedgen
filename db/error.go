package db

import (
	"fmt"

	"github.com/danlock/structs"

	"github.com/lib/pq"
)

func ErrDetails(err error) string {
	if err, ok := err.(*pq.Error); ok && err != nil {
		return fmt.Sprintf("pqerr:%s:%v", err.Code.Name(), structs.Map(err))
	}
	return err.Error()
}
