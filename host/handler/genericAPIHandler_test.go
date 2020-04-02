package handler

import (
	"fmt"
	"testing"
)

func Test_filterTypePrefixInForm(t *testing.T) {
	inputform := map[string]interface{}{
		"time_1" : "2016-08-15T01:24:33Z",
	}
	err := filterTypePrefixInForm(inputform)
	fmt.Println(inputform, err)
}