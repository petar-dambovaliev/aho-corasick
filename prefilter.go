package aho_corasick

import (
	"math"
)

type startBytesThree struct {
	byte1 byte
	byte2 byte
	byte3 byte
}

func (s startBytesThree) NextCandidate(_ *PrefilterState, haystack []byte, at int) (interface{}, CandidateType) {
	for i, b := range haystack[at:] {
		if s.byte1 == b || s.byte2 == b || s.byte3 == b {
			return at + i, PossibleStartOfMatchCandidate
		}
	}
	return nil, NoneCandidate
}

func (s startBytesThree) HeapBytes() int {
	return 0
}

func (s startBytesThree) ReportsFalsePositives() bool {
	return true
}

func (s startBytesThree) LooksForNonStartOfMatch() bool {
	return false
}

func (s *startBytesThree) clone() prefilter {
	if s == nil {
		return nil
	}
	u := *s
	return &u
}

type startBytesTwo struct {
	byte1 byte
	byte2 byte
}

func (s startBytesTwo) NextCandidate(_ *PrefilterState, haystack []byte, at int) (interface{}, CandidateType) {
	for i, b := range haystack[at:] {
		if s.byte1 == b || s.byte2 == b {
			return at + i, PossibleStartOfMatchCandidate
		}
	}
	return nil, NoneCandidate
}

func (s startBytesTwo) HeapBytes() int {
	return 0
}

func (s startBytesTwo) ReportsFalsePositives() bool {
	return true
}

func (s startBytesTwo) LooksForNonStartOfMatch() bool {
	return false
}

func (s *startBytesTwo) clone() prefilter {
	if s == nil {
		return nil
	}
	u := *s
	return &u
}

type startBytesOne struct {
	byte1 byte
}

func (s startBytesOne) NextCandidate(_ *PrefilterState, haystack []byte, at int) (interface{}, CandidateType) {
	for i, b := range haystack[at:] {
		if s.byte1 == b {
			return at + i, PossibleStartOfMatchCandidate
		}
	}
	return nil, NoneCandidate
}

func (s startBytesOne) HeapBytes() int {
	return 0
}

func (s startBytesOne) ReportsFalsePositives() bool {
	return true
}

func (s startBytesOne) LooksForNonStartOfMatch() bool {
	return false
}

func (s *startBytesOne) clone() prefilter {
	if s == nil {
		return nil
	}
	u := *s
	return &u
}

type byteSet [256]bool

func (b *byteSet) contains(bb byte) bool {
	return b[int(bb)]
}

func (b *byteSet) insert(bb byte) bool {
	n := !b.contains(bb)
	b[int(bb)] = true
	return n
}

type rareByteOffset struct {
	max byte
}

type rareByteOffsets struct {
	rbo [256]rareByteOffset
}

func (r *rareByteOffsets) set(b byte, off rareByteOffset) {
	m := byte(max(int(r.rbo[int(b)].max), int(off.max)))
	r.rbo[int(b)].max = m
}

type prefilterBuilder struct {
	count                int
	asciiCaseInsensitive bool
	startBytes           StartBytesBuilder
	rareBytes            rareBytesBuilder
}

func (p *prefilterBuilder) build() prefilter {
	startBytes := p.startBytes.build()
	rareBytes := p.rareBytes.build()

	switch true {
	case startBytes != nil && rareBytes != nil:
		hasFewerBytes := p.startBytes.count < p.rareBytes.count

		hasRarerBytes := p.startBytes.rankSum <= p.rareBytes.rankSum+50
		if hasFewerBytes || hasRarerBytes {
			return startBytes
		} else {
			return rareBytes
		}
	case startBytes != nil:
		return startBytes
	case rareBytes != nil:
		return rareBytes
	case p.asciiCaseInsensitive:
		return nil
	default:
		return nil
	}
}

func (p *prefilterBuilder) add(bytes []byte) {
	p.count += 1
	p.startBytes.add(bytes)
	p.rareBytes.add(bytes)
}

func newPrefilterBuilder(asciiCaseInsensitive bool) prefilterBuilder {
	return prefilterBuilder{
		count:                0,
		asciiCaseInsensitive: asciiCaseInsensitive,
		startBytes:           newStartBytesBuilder(asciiCaseInsensitive),
		rareBytes:            newRareBytesBuilder(asciiCaseInsensitive),
	}
}

type rareBytesBuilder struct {
	asciiCaseInsensitive bool
	rareSet              byteSet
	byteOffsets          rareByteOffsets
	available            bool
	count                int
	rankSum              uint16
}

type rareBytesOne struct {
	byte1  byte
	offset rareByteOffset
}

func (r rareBytesOne) NextCandidate(state *PrefilterState, haystack []byte, at int) (interface{}, CandidateType) {
	for i, b := range haystack[at:] {
		if r.byte1 == b {
			pos := at + i
			state.lastScanAt = pos
			r := pos - int(r.offset.max)
			if r < 0 {
				r = 0
			}

			if at > r {
				r = at
			}
			return r, PossibleStartOfMatchCandidate
		}
	}
	return nil, NoneCandidate
}

func (r rareBytesOne) HeapBytes() int {
	return 0
}

func (r rareBytesOne) ReportsFalsePositives() bool {
	return true
}

func (r rareBytesOne) LooksForNonStartOfMatch() bool {
	return true
}

func (r *rareBytesOne) clone() prefilter {
	if r == nil {
		return nil
	}
	u := *r
	return &u
}

type rareBytesTwo struct {
	offsets rareByteOffsets
	byte1   byte
	byte2   byte
}

func (r rareBytesTwo) NextCandidate(state *PrefilterState, haystack []byte, at int) (interface{}, CandidateType) {
	for i, b := range haystack[at:] {
		if r.byte1 == b || r.byte2 == b {
			pos := at + i
			state.updateAt(pos)
			r := pos - int(r.offsets.rbo[haystack[pos]].max)
			if r < 0 {
				r = 0
			}

			if at > r {
				r = at
			}
			return r, PossibleStartOfMatchCandidate
		}
	}
	return nil, NoneCandidate
}

func (r rareBytesTwo) HeapBytes() int {
	return 0
}

func (r rareBytesTwo) ReportsFalsePositives() bool {
	return true
}

func (r rareBytesTwo) LooksForNonStartOfMatch() bool {
	return true
}

func (r *rareBytesTwo) clone() prefilter {
	if r == nil {
		return nil
	}
	u := *r
	return &u
}

type rareBytesThree struct {
	offsets rareByteOffsets
	byte1   byte
	byte2   byte
	byte3   byte
}

func (r rareBytesThree) NextCandidate(state *PrefilterState, haystack []byte, at int) (interface{}, CandidateType) {
	for i, b := range haystack[at:] {
		if r.byte1 == b || r.byte2 == b || r.byte3 == b {
			pos := at + i
			state.updateAt(pos)
			r := pos - int(r.offsets.rbo[haystack[pos]].max)
			if r < 0 {
				r = 0
			}

			if at > r {
				r = at
			}
			return r, PossibleStartOfMatchCandidate
		}
	}
	return nil, NoneCandidate
}

func (r rareBytesThree) HeapBytes() int {
	return 0
}

func (r rareBytesThree) ReportsFalsePositives() bool {
	return true
}

func (r rareBytesThree) LooksForNonStartOfMatch() bool {
	return true
}

func (r *rareBytesThree) clone() prefilter {
	if r == nil {
		return nil
	}
	u := *r
	return &u
}

func (r *rareBytesBuilder) build() prefilter {
	if !r.available || r.count > 3 {
		return nil
	}
	var length int
	bytes := [3]byte{}

	for b := 0; b <= 255; b++ {
		if r.rareSet.contains(byte(b)) {
			bytes[length] = byte(b)
			length += 1
		}
	}

	switch length {
	case 0:
		return nil
	case 1:
		return &rareBytesOne{
			byte1:  bytes[0],
			offset: r.byteOffsets.rbo[bytes[0]],
		}
	case 2:
		return &rareBytesTwo{
			offsets: r.byteOffsets,
			byte1:   bytes[0],
			byte2:   bytes[1],
		}
	case 3:
		return &rareBytesThree{
			offsets: r.byteOffsets,
			byte1:   bytes[0],
			byte2:   bytes[1],
			byte3:   bytes[2],
		}
	default:
		return nil
	}
}

func (r *rareBytesBuilder) add(bytes []byte) {
	if !r.available {
		return
	}

	if r.count > 3 {
		r.available = false
		return
	}

	if len(bytes) >= 256 {
		r.available = false
		return
	}

	if len(bytes) == 0 {
		return
	}

	rarest1, rarest2 := bytes[0], freqRank(bytes[0])
	found := false

	for pos, b := range bytes {
		r.setOffset(pos, b)
		if found {
			continue
		}
		if r.rareSet.contains(b) {
			found = true
		}
		rank := freqRank(b)
		if rank < rarest2 {
			rarest1 = b
			rarest2 = rank
		}

		if !found {
			r.addRareByte(rarest1)
		}
	}
}

func (r *rareBytesBuilder) addRareByte(b byte) {
	r.addOneRareByte(b)
	if r.asciiCaseInsensitive {
		r.addOneRareByte(oppositeAsciiCase(b))
	}
}

func (r *rareBytesBuilder) addOneRareByte(b byte) {
	if r.rareSet.insert(b) {
		r.count += 1
		r.rankSum += uint16(freqRank(b))
	}
}

func newRareByteOffset(i int) rareByteOffset {
	if i > math.MaxUint8 {
		return rareByteOffset{max: 0}
	}
	b := byte(i)
	return rareByteOffset{max: b}
}

func (r *rareBytesBuilder) setOffset(pos int, b byte) {
	offset := newRareByteOffset(pos)
	r.byteOffsets.set(b, offset)

	if r.asciiCaseInsensitive {
		r.byteOffsets.set(oppositeAsciiCase(b), offset)
	}
}

func newRareBytesBuilder(asciiCaseInsensitive bool) rareBytesBuilder {
	return rareBytesBuilder{
		asciiCaseInsensitive: asciiCaseInsensitive,
		rareSet:              byteSet{},
		byteOffsets:          rareByteOffsets{},
		available:            true,
		count:                0,
		rankSum:              0,
	}
}

type StartBytesBuilder struct {
	asciiCaseInsensitive bool
	byteset              []bool
	count                int
	rankSum              uint16
}

func (s *StartBytesBuilder) build() prefilter {
	if s.count > 3 {
		return nil
	}
	var length int
	bytes := [3]byte{}

	for b := 0; b < 256; b++ {
		//todo case insensitive is not set in byteset
		if !s.byteset[b] {
			continue
		}
		if b > 0x7F {
			return nil
		}
		bytes[length] = byte(b)
		length += 1
	}

	switch length {
	case 0:
		return nil
	case 1:
		return &startBytesOne{byte1: bytes[0]}
	case 2:
		return &startBytesTwo{
			byte1: bytes[0],
			byte2: bytes[1],
		}
	case 3:
		return &startBytesThree{
			byte1: bytes[0],
			byte2: bytes[1],
			byte3: bytes[2],
		}
	default:
		return nil
	}
}

func (s *StartBytesBuilder) add(bytes []byte) {
	if s.count > 3 || len(bytes) == 0 {
		return
	}

	b := bytes[0]

	s.addOneByte(b)
	if s.asciiCaseInsensitive {
		s.addOneByte(oppositeAsciiCase(b))
	}
}

func (s *StartBytesBuilder) addOneByte(b byte) {
	if !s.byteset[int(b)] {
		s.byteset[int(b)] = true
		s.count += 1
		s.rankSum += uint16(freqRank(b))
	}
}

func freqRank(b byte) byte {
	return byteFrequencies[int(b)]
}

func newStartBytesBuilder(asciiCaseInsensitive bool) StartBytesBuilder {
	return StartBytesBuilder{
		asciiCaseInsensitive: asciiCaseInsensitive,
		byteset:              make([]bool, 256),
		count:                0,
		rankSum:              0,
	}
}

const MinSkips int = 40
const MinAvgFactor int = 2

type PrefilterState struct {
	skips       int
	skipped     int
	maxMatchLen int
	inert       bool
	lastScanAt  int
}

func (p *PrefilterState) updateAt(at int) {
	if at > p.lastScanAt {
		p.lastScanAt = at
	}
}

func (p *PrefilterState) IsEffective(at int) bool {
	if p.inert || at < p.lastScanAt {
		return false
	}

	if p.skips < MinSkips {
		return true
	}

	minAvg := MinAvgFactor * p.maxMatchLen

	if p.skipped >= minAvg*p.skips {
		return true
	}

	p.inert = true
	return false
}

func (p *PrefilterState) updateSkippedBytes(skipped int) {
	p.skips += 1
	p.skipped += skipped
}

type CandidateType uint

const (
	NoneCandidate CandidateType = iota
	MatchCandidate
	PossibleStartOfMatchCandidate
)

type prefilter interface {
	NextCandidate(state *PrefilterState, haystack []byte, at int) (interface{}, CandidateType)
	HeapBytes() int
	ReportsFalsePositives() bool
	LooksForNonStartOfMatch() bool
	clone() prefilter
}

func nextPrefilter(state *PrefilterState, prefilter prefilter, haystack []byte, at int) (interface{}, CandidateType) {
	cand, ttype := prefilter.NextCandidate(state, haystack, at)

	switch ttype {
	case NoneCandidate:
		state.updateSkippedBytes(len(haystack) - at)
	case MatchCandidate:
		m := cand.(*Match)
		state.updateSkippedBytes(m.Start() - at)
	case PossibleStartOfMatchCandidate:
		i := cand.(int)
		state.updateSkippedBytes(i - at)
	}
	return cand, ttype
}
