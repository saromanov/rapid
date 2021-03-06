// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package rapid

import (
	"fmt"
	"math/bits"
	"os"
	"time"
)

const shrinkTimeLimit = 30 * time.Second

func shrink(tb limitedTB, rec recordedBits, err *panicError, prop func(*T)) ([]uint64, *panicError) {
	rec.prune()

	s := &shrinker{
		tb:   tb,
		rec:  rec,
		err:  err,
		prop: prop,
	}

	buf, err := s.shrink()

	if *debugvis {
		name := fmt.Sprintf("vis-%v.html", tb.Name())
		f, err := os.Create(name)
		if err != nil {
			tb.Logf("failed to create debugvis file %v: %v", name, err)
		} else {
			defer f.Close()

			if err = visWriteHTML(f, tb.Name(), s.visBits); err != nil {
				tb.Logf("failed to write debugvis file %v: %v", name, err)
			}
		}
	}

	return buf, err
}

type shrinker struct {
	tb      limitedTB
	rec     recordedBits
	err     *panicError
	prop    func(*T)
	visBits []recordedBits
	tries   int
}

func (s *shrinker) debugf(format string, args ...interface{}) {
	if *debug {
		s.tb.Helper()
		s.tb.Logf("[shrink] "+format, args...)
	}
}

func (s *shrinker) shrink() (buf []uint64, err *panicError) {
	defer func() {
		if r := recover(); r != nil {
			buf, err = s.rec.data, r.(*panicError)
		}
	}()

	i := 0
	shrunk := true
	start := time.Now()
	for ; shrunk && time.Since(start) < shrinkTimeLimit; i++ {
		data := append([]uint64(nil), s.rec.data...)

		s.debugf("round %v start", i)
		s.removeBlockGroups()
		s.minimizeBlocks()

		shrunk = compareData(s.rec.data, data) < 0
	}
	s.debugf("done, %v rounds total (%v tries)", i, s.tries)

	return s.rec.data, s.err
}

func (s *shrinker) removeBlockGroups() {
	for i := 0; i < len(s.rec.groups); {
		g := s.rec.groups[i]
		if !g.removable {
			s.debugf("skip non-removable group %q at %v: [%v, %v)", g.label, i, g.begin, g.end)
			i++
			continue
		}

		buf := append([]uint64(nil), s.rec.data...)
		if g.end >= 0 {
			buf = append(buf[:g.begin], buf[g.end:]...)
		} else {
			buf = buf[:g.begin]
		}
		if !s.accept(buf, "remove group %q at %v: [%v, %v)", g.label, i, g.begin, g.end) {
			for i++; i < len(s.rec.groups) && s.rec.groups[i].begin == g.begin && s.rec.groups[i].end == g.end; i++ {
				s.debugf("skip duplicate group %v: [%v, %v)", i, g.begin, g.end)
			}
		}
	}
}

func (s *shrinker) minimizeBlocks() {
	for i := 0; i < len(s.rec.data); i++ {
		minimize(s.rec.data[i], func(u uint64) bool {
			buf := append([]uint64(nil), s.rec.data...)
			buf[i] = u
			return s.accept(buf, "minimize block %v: %v to %v", i, s.rec.data[i], u)
		})
	}
}

func (s *shrinker) accept(buf []uint64, format string, args ...interface{}) bool {
	if compareData(buf, s.rec.data) >= 0 {
		return false
	}

	s.tries++
	s1 := newBufBitStream(buf, false)
	t1 := newT(s.tb, s1, *debug)
	t1.Logf("[shrink] trying to reproduce the failure with a smaller test case: "+format, args...)
	err1 := checkOnce(t1, s.prop)
	if traceback(err1) != traceback(s.err) {
		return false
	}

	s.err = err1
	s2 := newBufBitStream(buf, true)
	t2 := newT(s.tb, s2, *debug)
	t2.Logf("[shrink] trying to reproduce the failure")
	err2 := checkOnce(t2, s.prop)
	s.rec = s2.recordedBits
	s.rec.prune()
	assert(compareData(s.rec.data, buf) <= 0)
	if *debugvis {
		s.visBits = append(s.visBits, s.rec)
	}
	if !sameError(err1, err2) {
		panic(err2)
	}

	return true
}

func minimize(u uint64, cond func(uint64) bool) uint64 {
	if u == 0 {
		return 0
	}
	for i := uint64(0); i < u && i < small; i++ {
		if cond(i) {
			return i
		}
	}
	if u <= small {
		return u
	}

	m := &minimizer{best: u, cond: cond}

	m.rShift()
	m.unsetBits()
	m.sortBits()
	m.binSearch()

	return m.best
}

type minimizer struct {
	best uint64
	cond func(uint64) bool
}

func (m *minimizer) accept(u uint64) bool {
	if u >= m.best || !m.cond(u) {
		return false
	}
	m.best = u
	return true
}

func (m *minimizer) rShift() {
	for m.accept(m.best >> 1) {
	}
}

func (m *minimizer) unsetBits() {
	size := uint(bits.Len64(m.best))

	for i := uint(0); i < size; i++ {
		m.accept(m.best ^ 1<<i)
	}
}

func (m *minimizer) sortBits() {
	size := uint(bits.Len64(m.best))

	for i := uint(0); i < size; i++ {
		for j := uint(0); j < size-i-1; j++ {
			l := uint64(1 << j)
			h := uint64(1 << (j + 1))
			if m.best&l == 0 && m.best&h != 0 {
				m.accept(m.best ^ (l | h))
			}
		}
	}
}

func (m *minimizer) binSearch() {
	if !m.accept(m.best - 1) {
		return
	}

	i := uint64(0)
	j := m.best
	for i < j {
		h := i + (j-i)/2
		if m.accept(h) {
			j = h
		} else {
			i = h + 1
		}
	}
}

func compareData(a []uint64, b []uint64) int {
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}

	for i := range a {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}

	return 0
}
