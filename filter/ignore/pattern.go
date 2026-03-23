package ignore

import (
	"path"
	"strings"
)

type Pattern struct {
	segments []string
	negated  bool
	dirOnly  bool
	rooted   bool
}

func Parse(line string) *Pattern {
	line = strings.TrimRight(line, " \t")
	if line == "" || line[0] == '#' {
		return nil
	}

	var p Pattern

	switch {
	case line[0] == '!':
		p.negated = true
		line = line[1:]
	case line[0] == '\\' && len(line) > 1 && (line[1] == '#' || line[1] == '!'):
		line = line[1:]
	}
	if line == "" {
		return nil
	}

	if line[len(line)-1] == '/' {
		p.dirOnly = true
		line = line[:len(line)-1]
	}
	if line == "" {
		return nil
	}

	if strings.Contains(line, "/") {
		p.rooted = true
	}

	line = strings.TrimPrefix(line, "/")
	p.segments = strings.Split(line, "/")

	return &p
}

func (p *Pattern) Match(name string) bool {
	if p == nil {
		return false
	}
	segs := strings.Split(name, "/")
	if p.rooted {
		return match(p.segments, segs)
	}
	// Non-rooted: match against the base name at any depth.
	base := segs[len(segs)-1]
	matched, _ := path.Match(p.segments[0], base)
	return matched
}

// match recursively matches pattern segments against path segments,
// treating "**" as a wildcard for zero or more path components.
func match(isegs, psegs []string) bool {
	if len(isegs) == 0 {
		return len(psegs) == 0
	}

	if isegs[0] == "**" {
		for i := 0; i <= len(psegs); i++ {
			if match(isegs[1:], psegs[i:]) {
				return true
			}
		}
		return false
	}
	if len(psegs) == 0 {
		return false
	}

	matched, _ := path.Match(isegs[0], psegs[0])
	if !matched {
		return false
	}
	return match(isegs[1:], psegs[1:])
}

// Ignored returns true if filePath should be blocked by the patterns.
// Patterns are evaluated in order; a negation pattern overrides prior matches.
func Ignored(patterns []Pattern, filePath string) bool {
	return File(patterns).Ignored(filePath)
}
