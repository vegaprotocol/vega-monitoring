package sqlstore

import (
	"fmt"
	"strings"
)

func PrepareListForInCondition[T string | []byte](input []T) string {
	result := ""

	for _, val := range input {
		result = result + fmt.Sprintf("'%s',", strings.ReplaceAll(string(val), "'", `\'`))
	}

	return strings.TrimRight(result, ",")
}
