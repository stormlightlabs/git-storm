package shared

import "testing"

func TestTitleCase(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		got := TitleCase("hello world")
		want := "Hello World"
		if got != want {
			t.Fatalf("TitleCase() = %q, want %q", got, want)
		}
	})

	t.Run("MixedCase", func(t *testing.T) {
		got := TitleCase("go is GREAT")
		want := "Go Is Great"
		if got != want {
			t.Fatalf("TitleCase() = %q, want %q", got, want)
		}
	})

	t.Run("WithPunctuation", func(t *testing.T) {
		got := TitleCase("don't stop believing")
		want := "Don't Stop Believing"
		if got != want {
			t.Fatalf("TitleCase() = %q, want %q", got, want)
		}
	})

	t.Run("ExtraSpaces", func(t *testing.T) {
		got := TitleCase("  leading and  internal   spaces ")
		if got != "  Leading And  Internal   Spaces " {
			t.Fatalf("TitleCase() = %q, spacing/words not as expected", got)
		}
	})
}
