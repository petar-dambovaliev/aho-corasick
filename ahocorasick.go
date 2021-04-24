package aho_corasick

import (
	"strings"
	"unicode"
)

type findIter struct {
	fsm                 imp
	prestate            *prefilterState
	haystack            []byte
	pos                 int
	matchOnlyWholeWords bool
}

type Iter interface {
	Next() *Match
}

func (f *findIter) Next() *Match {
	if f.pos > len(f.haystack) {
		return nil
	}

	result := f.fsm.FindAtNoState(f.prestate, f.haystack, f.pos)

	if result == nil {
		return nil
	}

	if result.end == f.pos {
		f.pos += 1
	} else {
		f.pos = result.end
	}

	if f.matchOnlyWholeWords {
		if result.Start()-1 >= 0 && (unicode.IsLetter(rune(f.haystack[result.Start()-1])) || unicode.IsDigit(rune(f.haystack[result.Start()-1]))) {
			return f.Next()
		}
		if result.end < len(f.haystack) && (unicode.IsLetter(rune(f.haystack[result.end])) || unicode.IsDigit(rune(f.haystack[result.end]))) {
			return f.Next()
		}
	}

	return result
}

type AhoCorasick struct {
	i                   imp
	matchKind           matchKind
	matchOnlyWholeWords bool
}

func (ac AhoCorasick) Iter(haystack string) Iter {
	prestate := &prefilterState{
		skips:       0,
		skipped:     0,
		maxMatchLen: ac.i.MaxPatternLen(),
		inert:       false,
		lastScanAt:  0,
	}

	return &findIter{
		fsm:                 ac.i,
		prestate:            prestate,
		haystack:            []byte(haystack),
		pos:                 0,
		matchOnlyWholeWords: ac.matchOnlyWholeWords,
	}
}

func (ac AhoCorasick) ReplaceAllFunc(haystack string, f func(match Match) (string, bool)) string {
	matches := ac.FindAll(haystack)

	if len(matches) == 0 {
		return haystack
	}

	replaceWith := make([]string, 0)

	for _, match := range matches {
		rw, ok := f(match)
		if !ok {
			break
		}
		replaceWith = append(replaceWith, rw)
	}

	diff := calcDiffSize(matches, replaceWith)

	var str strings.Builder
	str.Grow(len(haystack) + diff)

	start := 0

	for _, match := range matches {
		if match.Pattern() >= len(replaceWith) {
			str.WriteString(haystack[start:])
			return str.String()
		}
		str.WriteString(haystack[start:match.Start()])
		str.WriteString(replaceWith[match.Pattern()])
		start = match.Start() + match.len
	}

	if start-1 < len(haystack) {
		str.WriteString(haystack[start:])
	}

	return str.String()
}

func (ac AhoCorasick) ReplaceAll(haystack string, replaceWith []string) string {
	if len(replaceWith) != ac.i.PatternCount() {
		panic("replaceWith needs to have the same length as the pattern count")
	}

	return ac.ReplaceAllFunc(haystack, func(match Match) (string, bool) {
		return replaceWith[match.pattern], true
	})
}

func calcDiffSize(matches []Match, replaceWith []string) int {
	var diff int
	for _, match := range matches {
		if match.Pattern() >= len(replaceWith) {
			return diff
		}
		diff += len(replaceWith[match.pattern]) - match.len
	}
	return diff
}

func (ac AhoCorasick) FindAll(haystack string) []Match {
	iter := ac.Iter(haystack)
	matches := make([]Match, 0)

	for {
		next := iter.Next()
		if next == nil {
			break
		}

		matches = append(matches, *next)
	}

	return matches
}

type AhoCorasickBuilder struct {
	dfaBuilder          *iDFABuilder
	nfaBuilder          *iNFABuilder
	dfa                 bool
	matchOnlyWholeWords bool
}

type Opts struct {
	AsciiCaseInsensitive bool
	MatchOnlyWholeWords  bool
	MatchKind            matchKind
}

func NewAhoCorasickBuilder(o Opts) AhoCorasickBuilder {
	return AhoCorasickBuilder{
		dfaBuilder:          newDFABuilder(),
		nfaBuilder:          newNFABuilder(o.MatchKind, o.AsciiCaseInsensitive),
		dfa:                 false,
		matchOnlyWholeWords: o.MatchOnlyWholeWords,
	}
}

func (a *AhoCorasickBuilder) Build(patterns []string) AhoCorasick {
	nfa := a.nfaBuilder.build(patterns)
	match_kind := nfa.matchKind

	if a.dfa {
		dfa := a.dfaBuilder.build(nfa)
		return AhoCorasick{dfa, match_kind, a.matchOnlyWholeWords}
	}

	return AhoCorasick{nfa, match_kind, a.matchOnlyWholeWords}
}

type imp interface {
	MatchKind() *matchKind
	StartState() stateID
	MaxPatternLen() int
	PatternCount() int
	Prefilter() prefilter
	UsePrefilter() bool
	OverlappingFindAt(prestate *prefilterState, haystack []byte, at int, state_id *stateID, match_index *int) *Match
	EarliestFindAt(prestate *prefilterState, haystack []byte, at int, state_id *stateID) *Match
	FindAtNoState(prestate *prefilterState, haystack []byte, at int) *Match
}

type matchKind int

const (
	StandardMatch matchKind = iota
	LeftMostFirstMatch
	LeftMostLongestMatch
)

func (m matchKind) supportsOverlapping() bool {
	return m.isStandard()
}

func (m matchKind) supportsStream() bool {
	return m.isStandard()
}

func (m matchKind) isStandard() bool {
	return m == StandardMatch
}

func (m matchKind) isLeftmost() bool {
	return m == LeftMostFirstMatch || m == LeftMostLongestMatch
}

func (m matchKind) isLeftmostFirst() bool {
	return m == LeftMostFirstMatch
}

type Match struct {
	pattern int
	len     int
	end     int
}

func (m *Match) Pattern() int {
	return m.pattern
}

func (m *Match) End() int {
	return m.end
}

func (m *Match) Start() int {
	return m.end - m.len
}

type stateID uint

const (
	failedStateID stateID = 0
	deadStateID   stateID = 1
)
