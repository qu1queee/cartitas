package repo

import "testing"

func TestTopicDirName(t *testing.T) {
	if got := TopicDirName("geography"); got != "Geography" {
		t.Fatalf("got %q", got)
	}
	if got := TopicDirName("solar_system"); got != "SolarSystem" {
		t.Fatalf("got %q", got)
	}
}

func TestLoadLanguagesOrder(t *testing.T) {
	cfg, err := LoadLanguagesFile("../../languages.yaml")
	if err != nil {
		t.Fatal(err)
	}
	got := LanguageCodes(cfg)
	want := []string{"en", "es", "de"}
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("got %v want %v", got, want)
	}
}
