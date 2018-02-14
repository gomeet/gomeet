package remotecli

import (
	"regexp"
	"strings"
)

func mapToArray(m map[int32]string) []string {
	res := make([]string, len(m))
	i := 0
	for _, v := range m {
		res[i] = v
		i++
	}
	return res
}

func parseParams(s string) []string {
	r := regexp.MustCompile(`'.*?'|".*?"|\S+`)
	res := r.FindAllString(s, -1)
	for k, v := range res {
		mod := strings.Trim(v, " ")
		mod = strings.Trim(mod, "'")
		mod = strings.Trim(mod, `"`)
		mod = strings.Trim(mod, " ")

		res[k] = mod
	}
	return res
}
