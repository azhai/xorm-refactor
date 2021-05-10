package setting

import (
	"github.com/gobwas/glob"
)

type Globs []glob.Glob

func NewGlobs(wildcards []string) Globs {
	globs := make([]glob.Glob, 0)
	for _, w := range wildcards {
		if g, err := glob.Compile(w); err == nil {
			globs = append(globs, g)
		}
	}
	return globs
}

func (gs Globs) MatchAll(word string, forEmpty bool) bool {
	if len(gs) == 0 {
		return forEmpty
	}
	for _, matcher := range gs {
		if !matcher.Match(word) {
			return false
		}
	}
	return true
}

func (gs Globs) MatchAny(word string, forEmpty bool) bool {
	if len(gs) == 0 {
		return forEmpty
	}
	for _, matcher := range gs {
		if matcher.Match(word) {
			return true
		}
	}
	return false
}
