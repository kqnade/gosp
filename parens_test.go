package gosp

import "testing"

func TestParensBalanced(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want bool
	}{
		{"empty", "", true},
		{"single atom", "x\n", true},
		{"balanced flat", "(a b c)\n", true},
		{"balanced nested", "(a (b (c d)) e)\n", true},
		{"open without close", "(a b c\n", false},
		{"more closes than opens still balanced", "))\n", true},
		{"close in line comment", "(a ; )\n  b)\n", true},
		{"open in line comment ignored", "(a ; (\n  b)\n", true},
		{"unbalanced after comment line", "(a ; comment\n  b\n", false},
		{"comment only", "; just a comment\n", true},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			if got := parensBalanced(c.src); got != c.want {
				t.Fatalf("parensBalanced(%q) = %v, want %v", c.src, got, c.want)
			}
		})
	}
}
