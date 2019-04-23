package filehub

import "strings"

type errors []error

func (e errors) Error() string {
	s := make([]string, 0, len(e))
	for _, v := range e {
		s = append(s, v.Error())
	}
	return strings.Join(s, "\n")
}
