package main

import (
	"strconv"
	"strings"
	"testing"
)

func TestSectionNext_SameFile(t *testing.T) {
	data := []struct {
		curr       string
		line       string
		expSec     string
		expChanged bool
	}{
		// Starting from nothing
		{"", "", "0", false},
		{"", "Aaa", "0", false},
		{"", "# Title", "1 Title", true},
		{"", "#  \t  Title", "1 Title", true},

		// When we're at level 1, moving to level 1
		{"1 One", "", "1 One", false},
		{"1 One", "Aaa", "1 One", false},
		{"1 One", "# Next", "2 Next", true},
		{"1 One", "#  Next", "2 Next", true},
		{"2 Two", "# Next", "3 Next", true},

		// When we're at level 1, moving to level 2
		{"1 One", "## Twext", "1.1 Twext", true},
		{"1 One", "##  Twext", "1.1 Twext", true},
		{"2 Two", "## Twext", "2.1 Twext", true},

		// When we're at level 2, moving to level 2
		{"2.1 Twone", "", "2.1 Twone", false},
		{"2.1 Twone", "Aaa bbb", "2.1 Twone", false},
		{"2.1 Twone", "## Thext", "2.2 Thext", true},
		{"5.9 Fine", "## Twext", "5.10 Twext", true},

		// When we're at level 2, moving to level 1
		{"2.1 Twone", "# One", "3 One", true},
		{"5.9 Fine", "# One", "6 One", true},

		// When we're at level 3, moving to level 3
		{"2.1.3 Twone", "", "2.1.3 Twone", false},
		{"2.1.3 Twone", "Aaa bbb", "2.1.3 Twone", false},
		{"2.1.3 Twone", "### Thext", "2.1.4 Thext", true},
		{"5.9.1 Fine", "### Twext", "5.9.2 Twext", true},

		// When we're at level 3, moving to level 2
		{"2.1.3 Twone", "## Thext", "2.2 Thext", true},
		{"5.9.1 Fine", "## Twext", "5.10 Twext", true},

		// When we're at level 3, moving to level 1
		{"2.1.3 Twone", "# Thext", "3 Thext", true},
		{"5.9.1 Fine", "# Twext", "6 Twext", true},

		// When we're at level 1, moving to level 3
		{"1 One", "### Twext", "1.0.1 Twext", true},
		{"1 One", "###  Twext", "1.0.1 Twext", true},
		{"2 Two", "### Twext", "2.0.1 Twext", true},

		// When we're at level 2, moving to level 3
		{"2.1 Twone", "### Thext", "2.1.1 Thext", true},
		{"5.9 Fine", "### Twext", "5.9.1 Twext", true},
	}

	for _, d := range data {
		sec := toSection("dummy.md", d.curr)
		actSec, actChanged := sec.next(d.line)
		if actSec.toString() != d.expSec {
			t.Errorf("From %q with line %q got %q but expected %q",
				d.curr, d.line, actSec.toString(), d.expSec)
		}
		if actChanged != d.expChanged {
			t.Errorf("From %q with line %q got changed %t but expected %t",
				d.curr, d.line, actChanged, d.expChanged)
		}
	}
}

func toSection(inName string, line string) section {
	if line == "" {
		return section{}
	}

	a := strings.Index(line, " ")
	strNums := strings.Split(line[0:a], ".")
	nums := make([]int, len(strNums))
	for i, s := range strNums {
		nums[i], _ = strconv.Atoi(s)
	}

	return section{inName, nums, line[a+1:]}
}

func TestProcForSectionTrackingHeadings(t *testing.T) {
	s := newState()
	s.setFirstInName("headings.md")
	d := newDoc()
	tData := []struct {
		line string // Next line
		exp  string // Expected section as a string
	}{
		{"Aaa", "0"},
		{"# Title", "1 Title"},
		{"", "1 Title"},
		{"## Subtitle", "1.1 Subtitle"},
		{"Content", "1.1 Subtitle"},
		{"# Next", "2 Next"},
		{"More content", "2 Next"},
		{"```", "2 Next"},
		{"# Code comment", "2 Next"},
		{"```", "2 Next"},
		{"", "2 Next"},
		{"## After code", "2.1 After code"},
	}

	for i, p := range tData {
		s.proc(&s, &d, p.line)
		strSec := s.sec.toString()
		if strSec != p.exp {
			t.Errorf("Line %d: Expected sec=%q but got %q",
				i+1, p.exp, strSec)
		}
	}
}

func TestProcForSectionTrackingStartLines(t *testing.T) {
	s := newState()
	s.setFirstInName("startlines.md")
	d := newDoc()
	tData := []struct {
		line  string // Next line
		start bool   // True if it's supposed to be a section start
	}{
		{"Aaa", false},
		{"# Title", true},
		{"", false},
		{"## Subtitle", true},
		{"Content", false},
		{"# Next", true},
		{"More content", false},
		{"```", false},
		{"# Code comment", false},
		{"```", false},
		{"", false},
		{"## After code", true},
	}

	// Process all the lines
	for _, p := range tData {
		s.proc(&s, &d, p.line)
	}

	for i, p := range tData {
		if _, okay := d.secStarts[s.inName][i+1]; okay != p.start {
			t.Errorf("Line %d: Expected section start %t but got %t",
				i+1, p.start, okay)
		}
	}
}

func TestSectionLess(t *testing.T) {
	s0 := section{"n/a", []int(nil), ""}
	s1 := section{"n/a", []int{1}, ""}
	s2 := section{"n/a", []int{2}, ""}
	s2_4 := section{"n/a", []int{2, 4}, ""}
	s2_5 := section{"n/a", []int{2, 5}, ""}
	s3 := section{"n/a", []int{3}, ""}
	s3_4 := section{"n/a", []int{3, 4}, ""}
	s3_4_2 := section{"n/a", []int{3, 4, 2}, ""}
	s3_4_3 := section{"n/a", []int{3, 4, 3}, ""}
	data := []struct {
		a   section
		b   section
		exp bool
	}{
		// Zero comparitors
		{s0, s0, false},
		{s0, s1, true},
		{s1, s0, false},
		{s0, s3_4_2, true},
		{s3_4_2, s0, false},
		// 1 comparitors
		{s1, s1, false},
		{s2, s2, false},
		{s1, s2, true},
		{s2, s1, false},
		{s2, s2_4, true},
		{s2_4, s2, false},
		{s2, s3_4_3, true},
		{s3_4_3, s2, false},
		{s3, s3_4_3, true},
		{s3_4_3, s3, false},
		// 2 comparitors
		{s2_4, s2_4, false},
		{s2_4, s2_5, true},
		{s2_5, s2_4, false},
		{s3_4, s3_4_2, true},
		{s3_4_2, s3_4, false},
		// 3 comparitors
		{s3_4_2, s3_4_2, false},
		{s3_4_2, s3_4_3, true},
		{s3_4_3, s3_4_2, false},
	}

	for _, d := range data {
		act := d.a.less(d.b)
		if act != d.exp {
			t.Errorf("%#v < %#v? Expected %t but got %t",
				d.a.nums, d.b.nums, d.exp, act)
		}
	}
}
