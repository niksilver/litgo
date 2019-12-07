package main

import (
	"strconv"
	"strings"
	"testing"
)

func TestNewSection(t *testing.T) {
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
		sec := toSection(d.curr)
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

func toSection(line string) section {
	if line == "" {
		return section{}
	}

	a := strings.Index(line, " ")
	strNums := strings.Split(line[0:a], ".")
	nums := make([]int, len(strNums))
	for i, s := range strNums {
		nums[i], _ = strconv.Atoi(s)
	}

	return section{nums, line[a+1:]}
}

func TestProcForSectionTracking(t *testing.T) {
	s := newState()
	data := []struct {
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

	for i, d := range data {
		proc(&s, d.line)
		strSec := s.sec.toString()
		if strSec != d.exp {
			t.Errorf("Line %d: Expected sec=%q but got %q",
				i+1, d.exp, strSec)
		}
	}
}

func TestProcForNumsInSectionHeadings(t *testing.T) {
	s := newState()
	data := []struct {
		line string // Next line
		exp  string // Expected markdown line
	}{
		{"Aaa", "Aaa"},
		{"# Title", "# 1 Title"},
		{"", ""},
		{"## Subtitle", "## 1.1 Subtitle"},
		{"Content", "Content"},
		{"#Not a heading", "#Not a heading"},
		{"", ""},
		{"## Subheading", "## 1.2 Subheading"},
	}

	for i, d := range data {
		s.markdown = strings.Builder{}
		proc(&s, d.line)
		line := s.markdown.String()
		md := line[0 : len(line)-1]
		if md != d.exp {
			t.Errorf("Line %d: Expected markdown %q but got %q",
				i+1, d.exp, md)
		}
	}
}
