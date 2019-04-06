package db

import (
	"fmt"

	"github.com/lib/pq"
)

func ErrDetails(err error) string {
	if err, ok := err.(*pq.Error); ok {
		return fmt.Sprintf("pqerr:%s:%+v", err.Code.Name(), err)
	}
	return err.Error()
}
