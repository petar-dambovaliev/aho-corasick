package aho_corasick

type automaton interface {
	Repr() *iRepr
	MatchKind() *matchKind
	Anchored() bool
	Prefilter() prefilter
	StartState() stateID
	IsValid(stateID) bool
	IsMatchState(stateID) bool
	IsMatchOrDeadState(stateID) bool
	GetMatch(stateID, int, int) *Match
	MatchCount(stateID) int
	NextState(stateID, byte) stateID
	NextStateNoFail(stateID, byte) stateID
	StandardFindAt(*PrefilterState, []byte, int, *stateID) *Match
	StandardFindAtImp(*PrefilterState, prefilter, []byte, int, *stateID) *Match
	LeftmostFindAt(*PrefilterState, []byte, int, *stateID) *Match
	LeftmostFindAtImp(*PrefilterState, prefilter, []byte, int, *stateID) *Match
	LeftmostFindAtNoState(*PrefilterState, []byte, int) *Match
	LeftmostFindAtNoStateImp(*PrefilterState, prefilter, []byte, int) *Match
	OverlappingFindAt(*PrefilterState, []byte, int, *stateID, *int) *Match
	EarliestFindAt(*PrefilterState, []byte, int, *stateID) *Match
	FindAt(*PrefilterState, []byte, int, *stateID) *Match
	FindAtNoState(*PrefilterState, []byte, int) *Match
}

func isMatchOrDeadState(a automaton, si stateID) bool {
	return si == deadStateID || a.IsMatchState(si)
}

func nextStateNoFail(a automaton, cur stateID, input byte) stateID {
	next := a.NextState(cur, input)
	if next == failedStateID {
		panic("automaton should never return fail_id for next state")
	}
	return next
}

func standardFindAt(a automaton, prestate *PrefilterState, haystack []byte, at int, sID *stateID) *Match {
	pre := a.Prefilter()
	return a.StandardFindAtImp(prestate, pre, haystack, at, sID)
}

func standardFindAtImp(a automaton, prestate *PrefilterState, prefilter prefilter, haystack []byte, at int, sID *stateID) *Match {
	for at < len(haystack) {
		if prefilter != nil {
			startState := a.StartState()
			if prestate.IsEffective(at) && sID == &startState {
				c, ttype := nextPrefilter(prestate, prefilter, haystack, at)
				switch ttype {
				case NoneCandidate:
					return nil
				case PossibleStartOfMatchCandidate:
					i := c.(int)
					at = i
				}
			}
		}
		*sID = a.NextStateNoFail(*sID, haystack[at])
		at += 1

		if a.IsMatchOrDeadState(*sID) {
			if *sID == deadStateID {
				return nil
			} else {
				return a.GetMatch(*sID, 0, at)
			}
		}
	}
	return nil
}

func leftmostFindAt(a automaton, prestate *PrefilterState, haystack []byte, at int, sID *stateID) *Match {
	prefilter := a.Prefilter()
	return a.LeftmostFindAtImp(prestate, prefilter, haystack, at, sID)
}

func leftmostFindAtImp(a automaton, prestate *PrefilterState, prefilter prefilter, haystack []byte, at int, sID *stateID) *Match {
	if a.Anchored() && at > 0 && *sID == a.StartState() {
		return nil
	}
	lastMatch := a.GetMatch(*sID, 0, at)

	for at < len(haystack) {
		if prefilter != nil {
			startState := a.StartState()
			if prestate.IsEffective(at) && sID == &startState {
				c, ttype := nextPrefilter(prestate, prefilter, haystack, at)
				switch ttype {
				case NoneCandidate:
					return nil
				case PossibleStartOfMatchCandidate:
					i := c.(int)
					at = i
				}
			}
		}

		*sID = a.NextStateNoFail(*sID, haystack[at])
		at += 1

		if a.IsMatchOrDeadState(*sID) {
			if *sID == deadStateID {
				return lastMatch
			} else {
				a.GetMatch(*sID, 0, at)
			}
		}
	}

	return lastMatch
}

func leftmostFindAtNoState(a automaton, prestate *PrefilterState, haystack []byte, at int) *Match {
	return leftmostFindAtNoStateImp(a, prestate, a.Prefilter(), haystack, at)
}

func leftmostFindAtNoStateImp(a automaton, prestate *PrefilterState, prefilter prefilter, haystack []byte, at int) *Match {
	if a.Anchored() && at > 0 {
		return nil
	}
	if prefilter != nil && !prefilter.ReportsFalsePositives() {
		c, ttype := prefilter.NextCandidate(prestate, haystack, at)
		switch ttype {
		case NoneCandidate:
			return nil
		case MatchCandidate:
			m := c.(*Match)
			return m
		}
	}

	stateID := a.StartState()
	lastMatch := a.GetMatch(stateID, 0, at)

	for at < len(haystack) {
		if prefilter != nil && prestate.IsEffective(at) && stateID == a.StartState() {
			c, ttype := prefilter.NextCandidate(prestate, haystack, at)
			switch ttype {
			case NoneCandidate:
				return nil
			case MatchCandidate:
				m := c.(*Match)
				return m
			case PossibleStartOfMatchCandidate:
				i := c.(int)
				at = i
			}
		}

		stateID = a.NextStateNoFail(stateID, haystack[at])
		at += 1

		if a.IsMatchOrDeadState(stateID) {
			if stateID == deadStateID {
				return lastMatch
			}
			lastMatch = a.GetMatch(stateID, 0, at)
		}
	}

	return lastMatch
}

func overlappingFindAt(a automaton, prestate *PrefilterState, haystack []byte, at int, id *stateID, matchIndex *int) *Match {
	if a.Anchored() && at > 0 && *id == a.StartState() {
		return nil
	}

	matchCount := a.MatchCount(*id)

	if *matchIndex < matchCount {
		result := a.GetMatch(*id, *matchIndex, at)
		*matchIndex += 1
		return result
	}

	*matchIndex = 0
	match := a.StandardFindAt(prestate, haystack, at, id)

	if match == nil {
		return nil
	}

	*matchIndex = 1
	return match
}

func earliestFindAt(a automaton, prestate *PrefilterState, haystack []byte, at int, id *stateID) *Match {
	if *id == a.StartState() {
		if a.Anchored() && at > 0 {
			return nil
		}
		match := a.GetMatch(*id, 0, at)
		if match != nil {
			return match
		}
	}
	return a.StandardFindAt(prestate, haystack, at, id)
}

func findAt(a automaton, prestate *PrefilterState, haystack []byte, at int, id *stateID) *Match {
	kind := a.MatchKind()
	if kind == nil {
		return nil
	}
	switch *kind {
	case StandardMatch:
		return a.EarliestFindAt(prestate, haystack, at, id)
	case LeftMostFirstMatch, LeftMostLongestMatch:
		return a.LeftmostFindAt(prestate, haystack, at, id)
	}
	return nil
}

func findAtNoState(a automaton, prestate *PrefilterState, haystack []byte, at int) *Match {
	kind := a.MatchKind()
	if kind == nil {
		return nil
	}
	switch *kind {
	case StandardMatch:
		state := a.StartState()
		return a.EarliestFindAt(prestate, haystack, at, &state)
	case LeftMostFirstMatch, LeftMostLongestMatch:
		return a.LeftmostFindAtNoState(prestate, haystack, at)
	}
	return nil
}
