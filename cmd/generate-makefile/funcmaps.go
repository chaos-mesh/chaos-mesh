package main

import "strings"

var defaultFuncMap = map[string]interface{}{"StringsJoin": StringsJoin}

func StringsJoin(s []string, sep string) string {
	return strings.Join(s, sep)
}
