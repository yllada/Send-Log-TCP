package tui

import (
	"strings"
)

type Option struct {
	Value  string
	Label  string
	Helper string
}

type OptionSelector struct {
	options []Option
	index   int
}

func NewOptionSelector(options []Option) *OptionSelector {
	return &OptionSelector{options: options, index: 0}
}

func (s *OptionSelector) SetByValue(value string) {
	for i, o := range s.options {
		if o.Value == value {
			s.index = i
			return
		}
	}
}

func (s *OptionSelector) Selected() Option {
	return s.options[s.index]
}

func (s *OptionSelector) Value() string {
	return s.Selected().Value
}

func (s *OptionSelector) Up() {
	if s.index > 0 {
		s.index--
	}
}

func (s *OptionSelector) Down() {
	if s.index < len(s.options)-1 {
		s.index++
	}
}

func (s *OptionSelector) Toggle() {
	if len(s.options) != 2 {
		return
	}
	if s.index == 0 {
		s.index = 1
	} else {
		s.index = 0
	}
}

func (s *OptionSelector) View() string {
	parts := make([]string, 0, len(s.options))
	for i, o := range s.options {
		text := o.Label
		if o.Helper != "" {
			text += " (" + o.Helper + ")"
		}
		if i == s.index {
			parts = append(parts, accent.Render("\u25b8 "+text))
		} else {
			parts = append(parts, "  "+text)
		}
	}
	return strings.Join(parts, "  ")
}
