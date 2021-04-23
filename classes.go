package aho_corasick

import "math"

type byteClassRepresentatives struct {
	classes   *byteClasses
	bbyte     int
	lastClass *byte
}

func (b *byteClassRepresentatives) next() *byte {
	for b.bbyte < 256 {
		bbyte := byte(b.bbyte)
		class := b.classes[bbyte]
		b.bbyte += 1

		if b.lastClass != nil && *b.lastClass != class {
			b.lastClass = &class
			return &bbyte
		}
	}
	return nil
}

type byteClassBuilder []bool

func (b byteClassBuilder) setRange(start, end byte) {
	if start > 0 {
		b[int(start)-1] = true
	}
	b[int(end)] = true
}

func (b byteClassBuilder) build() byteClasses {
	var classes byteClasses
	var class byte
	i := 0
	for {
		classes[byte(i)] = class
		if i >= 255 {
			break
		}
		if b[i] {
			if class+1 > math.MaxUint8 {
				panic("shit happens")
			}
			class += 1
		}
		i += 1
	}
	return classes
}

func newByteClassBuilder() byteClassBuilder {
	return make([]bool, 256)
}

type byteClasses [256]byte

func singletons() byteClasses {
	var bc byteClasses
	for i := range bc {
		bc[i] = byte(i)
	}
	return bc
}

func (b byteClasses) alphabetLen() int {
	return int(b[255]) + 1
}

func (b byteClasses) isSingleton() bool {
	return b.alphabetLen() == 256
}
