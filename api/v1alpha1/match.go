package v1alpha1

import (
	"regexp"
	"strings"
)

func (m *IstioFilter_StringMatch) MatchValue(value string) bool {
	switch m.GetMatch().(type) {
	case *IstioFilter_StringMatch_Exact:
		return m.GetExact() == value
	case *IstioFilter_StringMatch_Prefix:
		return strings.HasPrefix(value, m.GetPrefix())

	case *IstioFilter_StringMatch_Suffix:
		return strings.HasSuffix(value, m.GetSuffix())

	case *IstioFilter_StringMatch_Regex:
		compile, err := regexp.Compile(m.GetRegex())
		if err != nil {
			return false
		}
		return compile.MatchString(value)
	}
	return true
}
