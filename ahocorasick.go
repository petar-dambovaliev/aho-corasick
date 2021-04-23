package aho_corasick

import (
	"unicode"
)

type findIter struct {
	fsm                 imp
	prestate            *PrefilterState
	haystack            []byte
	pos                 int
	matchOnlyWholeWords bool
}

func (f *findIter) next() *Match {
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

	return result
}

type AhoCorasick struct {
	i                   imp
	matchKind           matchKind
	matchOnlyWholeWords bool
}

func (ac AhoCorasick) findIter(haystack string) findIter {
	prestate := &PrefilterState{
		skips:       0,
		skipped:     0,
		maxMatchLen: ac.i.MaxPatternLen(),
		inert:       false,
		lastScanAt:  0,
	}

	return findIter{
		fsm:                 ac.i,
		prestate:            prestate,
		haystack:            []byte(haystack),
		pos:                 0,
		matchOnlyWholeWords: ac.matchOnlyWholeWords,
	}
}

func (ac AhoCorasick) FindAll(haystack string) []Match {
	iter := ac.findIter(haystack)
	matches := make([]Match, 0)

	for {
		next := iter.next()
		if next == nil {
			break
		}

		if ac.matchOnlyWholeWords {
			if next.Start()-1 >= 0 && (unicode.IsLetter(rune(haystack[next.Start()-1])) || unicode.IsDigit(rune(haystack[next.Start()-1]))) {
				continue
			}
			if next.end < len(haystack) && (unicode.IsLetter(rune(haystack[next.end])) || unicode.IsDigit(rune(haystack[next.end]))) {
				continue
			}
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
	OverlappingFindAt(prestate *PrefilterState, haystack []byte, at int, state_id *stateID, match_index *int) *Match
	EarliestFindAt(prestate *PrefilterState, haystack []byte, at int, state_id *stateID) *Match
	FindAtNoState(prestate *PrefilterState, haystack []byte, at int) *Match
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
	return m.Pattern()
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
