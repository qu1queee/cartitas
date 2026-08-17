package cardsplit

import "testing"

func TestSplitAndJoin(t *testing.T) {
	text := "# Title\n\nQ: One?\nA: Yes.\n\n---\n\nQ: Two?\nA: No.\n"
	parts := Split(text)
	if len(parts) != 2 {
		t.Fatalf("got %d parts: %#v", len(parts), parts)
	}
	joined := Join(parts)
	again := Split(joined)
	if len(again) != 2 {
		t.Fatalf("round-trip got %d", len(again))
	}
}

func TestSplitEmpty(t *testing.T) {
	if Split("   ") != nil {
		t.Fatal("expected nil")
	}
}
