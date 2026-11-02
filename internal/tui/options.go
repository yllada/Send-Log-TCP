package tui

import (
	"strings"
)

// Option is one selectable entry in an OptionSelector with its value, label, and optional helper text.
type Option struct {
	Value  string
	Label  string
	Helper string
}

// OptionSelector is a single-choice picker over a fixed list of options.
type OptionSelector struct {
	options []Option
	index   int
}

// NewOptionSelector creates a selector starting on the first option.
func NewOptionSelector(options []Option) *OptionSelector {
	return &OptionSelector{options: options, index: 0}
}

// SetByValue selects the option whose value matches, if any.
func (s *OptionSelector) SetByValue(value string) {
	for i, o := range s.options {
		if o.Value == value {
			s.index = i
			return
		}
	}
}

// Selected returns the currently selected option.
func (s *OptionSelector) Selected() Option {
	return s.options[s.index]
}

// Value returns the selected option's value.
func (s *OptionSelector) Value() string {
	return s.Selected().Value
}

// Up moves the selection one entry toward the start, clamped.
func (s *OptionSelector) Up() {
	if s.index > 0 {
		s.index--
	}
}

// Down moves the selection one entry toward the end, clamped.
func (s *OptionSelector) Down() {
	if s.index < len(s.options)-1 {
		s.index++
	}
}

// Toggle flips the selection between two options, or does nothing unless the selector holds exactly two.
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

// View renders the options on one line, highlighting the selected one.
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
