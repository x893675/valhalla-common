package policy

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

func IAMMatcher(arguments ...interface{}) (interface{}, error) {
	name1 := arguments[0].(string)
	name2 := arguments[1].(string)
	return DefaultMatcher.Matches(name1, name2)
}

var DefaultMatcher = NewRegexpMatcher(512)

func NewRegexpMatcher(size int) *RegexpMatcher {
	if size <= 0 {
		size = 512
	}

	// golang-lru only returns an error if the cache's size is 0. This, we can safely ignore this error.
	cache, _ := lru.New(size)
	return &RegexpMatcher{
		Cache: cache,
	}
}

type RegexpMatcher struct {
	*lru.Cache

	C map[string]*regexp2.Regexp
}

func (m *RegexpMatcher) get(pattern string) *regexp2.Regexp {
	var reg *regexp2.Regexp
	if val, ok := m.Cache.Get(pattern); !ok {
		return reg
	} else if reg, ok = val.(*regexp2.Regexp); !ok {
		return reg
	}
	return reg
}

func (m *RegexpMatcher) set(pattern string, reg *regexp2.Regexp) {
	m.Cache.Add(pattern, reg)
}

func (m *RegexpMatcher) MustMatch(key1 string, key2 string) bool {
	ok, err := m.Matches(key1, key2)
	if err != nil {
		return false
	}
	return ok
}

// Matches a key1 with pattern key2
// key1 form request
// key2 from policy
func (m *RegexpMatcher) Matches(key1 string, key2 string) (bool, error) {
	haystack := strings.Split(key2, ",")
	return m.matches(key1, haystack)
}

// matches a needle with an array of regular expressions and returns true if a match was found.
func (m *RegexpMatcher) matches(needle string, haystack []string) (bool, error) {
	var reg *regexp2.Regexp
	var err error
	for _, h := range haystack {

		// This means that the current haystack item does not contain a wildcard
		if !strings.Contains(h, "*") {
			// If we have a simple string match, we've got a match!
			if h == needle {
				return true, nil
			}

			// Not string match, but also no wildcard, continue with next haystack item
			continue
		}

		if reg = m.get(h); reg != nil {
			if matched, err := reg.MatchString(needle); err != nil {
				// according to regexp2 documentation: https://github.com/dlclark/regexp2#usage
				// The only error that the *Match* methods should return is a Timeout if you set the
				// re.MatchTimeout field. Any other error is a bug in the regexp2 package.
				return false, errors.WithStack(err)
			} else if matched {
				return true, nil
			}
			continue
		}

		reg, err = CompileWildcardRegex(h)
		if err != nil {
			return false, errors.WithStack(err)
		}

		m.set(h, reg)
		if matched, err := reg.MatchString(needle); err != nil {
			// according to regexp2 documentation: https://github.com/dlclark/regexp2#usage
			// The only error that the *Match* methods should return is a Timeout if you set the
			// re.MatchTimeout field. Any other error is a bug in the regexp2 package.
			return false, errors.WithStack(err)
		} else if matched {
			return true, nil
		}
	}
	return false, nil
}

const regexp2MatchTimeout = time.Millisecond * 250

// delimiterIndices returns the first level delimiter indices from a string.
// It returns an error in case of unbalanced delimiters.
func delimiterIndices(s string, delimiterStart, delimiterEnd byte) ([]int, error) {
	var level, idx int
	idxs := make([]int, 0)
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case delimiterStart:
			if level++; level == 1 {
				idx = i
			}
		case delimiterEnd:
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf(`Unbalanced braces in "%q"`, s)
			}
		}
	}

	if level != 0 {
		return nil, fmt.Errorf(`Unbalanced braces in "%q"`, s)
	}

	return idxs, nil
}

// CompileRegex parses a template and returns a Regexp.
//
// You can define your own delimiters. It is e.g. common to use curly braces {} but I recommend using characters
// which have no special meaning in Regex, e.g.: <, >
//
//	reg, err := compiler.CompileRegex("foo:bar.baz:<[0-9]{2,10}>", '<', '>')
//	// if err != nil ...
//	reg.MatchString("foo:bar.baz:123")
func CompileRegex(tpl string, delimiterStart, delimiterEnd byte) (*regexp2.Regexp, error) {
	// Check if it is well-formed.
	idxs, errBraces := delimiterIndices(tpl, delimiterStart, delimiterEnd)
	if errBraces != nil {
		return nil, errBraces
	}
	varsR := make([]*regexp2.Regexp, len(idxs)/2)
	pattern := bytes.NewBufferString("")
	pattern.WriteByte('^')

	var end int
	for i := 0; i < len(idxs); i += 2 {
		// Set all values we are interested in.
		raw := tpl[end:idxs[i]]
		end = idxs[i+1]
		patt := tpl[idxs[i]+1 : end-1]
		// Build the regexp pattern.
		varIdx := i / 2
		fmt.Fprintf(pattern, "%s(%s)", regexp.QuoteMeta(raw), patt)
		reg, err := regexp2.Compile(fmt.Sprintf("^%s$", patt), regexp2.RE2)
		if err != nil {
			return nil, err
		}
		reg.MatchTimeout = regexp2MatchTimeout
		varsR[varIdx] = reg
	}

	// Add the remaining.
	raw := tpl[end:]
	pattern.WriteString(regexp.QuoteMeta(raw))
	pattern.WriteByte('$')

	// Compile full regexp.
	reg, errCompile := regexp2.Compile(pattern.String(), regexp2.RE2)
	if errCompile != nil {
		return nil, errCompile
	}
	reg.MatchTimeout = regexp2MatchTimeout

	return reg, nil
}

// CompileWildcardRegex converts a wildcard pattern (using *) to a regexp2.Regexp
// Examples:
//   - "ecs:Describe*" matches "ecs:DescribeInstances", "ecs:Describe", etc.
//   - "*" matches any string
//   - "ecs:*:instance/*" matches "ecs:cn-hangzhou:instance/i-001", etc.
func CompileWildcardRegex(pattern string) (*regexp2.Regexp, error) {
	// Escape all regex special characters except *
	var buf bytes.Buffer
	buf.WriteByte('^')

	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		if c == '*' {
			// Convert * to .*
			buf.WriteString(".*")
		} else {
			// Quote special regex characters
			buf.WriteString(regexp.QuoteMeta(string(c)))
		}
	}

	buf.WriteByte('$')

	reg, err := regexp2.Compile(buf.String(), regexp2.RE2)
	if err != nil {
		return nil, err
	}
	reg.MatchTimeout = regexp2MatchTimeout

	return reg, nil
}
