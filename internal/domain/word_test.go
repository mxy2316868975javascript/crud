package domain

import "testing"

func stringPointer(value string) *string {
	return &value
}

func intPointer(value int) *int {
	return &value
}

func TestNewWord(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		meaning     string
		level       int
		wantWord    string
		wantMeaning string
		wantLevel   int
		wantError   string
	}{
		{
			name:        "合法词汇",
			text:        "apple",
			meaning:     "n. 苹果",
			level:       1,
			wantWord:    "apple",
			wantMeaning: "n. 苹果",
			wantLevel:   1,
		},
		{
			name:        "去掉首尾空格",
			text:        "  apple  ",
			meaning:     "  n. 苹果  ",
			level:       2,
			wantWord:    "apple",
			wantMeaning: "n. 苹果",
			wantLevel:   2,
		},
		{
			name:      "word 为空",
			text:      "",
			meaning:   "n. 苹果",
			level:     1,
			wantError: "word is required",
		},
		{
			name:      "word 只有空格",
			text:      "   ",
			meaning:   "n. 苹果",
			level:     1,
			wantError: "word is required",
		},
		{
			name:      "meaning 为空",
			text:      "apple",
			meaning:   "",
			level:     1,
			wantError: "meaning is required",
		},
		{
			name:      "level 太小",
			text:      "apple",
			meaning:   "n. 苹果",
			level:     0,
			wantError: "level must be between 1 and 5",
		},
		{
			name:      "level 太大",
			text:      "apple",
			meaning:   "n. 苹果",
			level:     6,
			wantError: "level must be between 1 and 5",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			word, err := NewWord(test.text, test.meaning, test.level)
			if test.wantError != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", test.wantError)
				}
				if err.Error() != test.wantError {
					t.Fatalf("error = %q, want %q", err.Error(), test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if word.ID != 0 {
				t.Errorf("ID = %d, want 0", word.ID)
			}
			if word.Word != test.wantWord {
				t.Errorf("Word = %q, want %q", word.Word, test.wantWord)
			}
			if word.Meaning != test.wantMeaning {
				t.Errorf("Meaning = %q, want %q", word.Meaning, test.wantMeaning)
			}
			if word.Level != test.wantLevel {
				t.Errorf("Level = %d, want %d", word.Level, test.wantLevel)
			}
		})
	}
}

func TestWordPatchNormalize(t *testing.T) {
	t.Run("空 patch", func(t *testing.T) {
		_, err := WordPatch{}.Normalize()
		if err == nil || err.Error() != "at least one field is required" {
			t.Fatalf("error = %v, want at least one field is required", err)
		}
	})

	t.Run("去掉首尾空格且不修改原指针", func(t *testing.T) {
		originalWord := "  apple  "
		originalMeaning := "  n. 苹果  "
		patch := WordPatch{
			Word:    &originalWord,
			Meaning: &originalMeaning,
			Level:   intPointer(3),
		}

		normalized, err := patch.Normalize()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if *normalized.Word != "apple" {
			t.Errorf("Word = %q, want apple", *normalized.Word)
		}
		if *normalized.Meaning != "n. 苹果" {
			t.Errorf("Meaning = %q, want n. 苹果", *normalized.Meaning)
		}
		if *normalized.Level != 3 {
			t.Errorf("Level = %d, want 3", *normalized.Level)
		}
		if *patch.Word != "  apple  " || *patch.Meaning != "  n. 苹果  " {
			t.Errorf("Normalize mutated the original patch: %+v", patch)
		}
	})

	t.Run("只传 level 时其他字段保持 nil", func(t *testing.T) {
		normalized, err := (WordPatch{Level: intPointer(3)}).Normalize()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if normalized.Word != nil || normalized.Meaning != nil {
			t.Errorf("unexpected non-nil fields: %+v", normalized)
		}
	})

	tests := []struct {
		name      string
		patch     WordPatch
		wantError string
	}{
		{
			name:      "word 为空",
			patch:     WordPatch{Word: stringPointer("")},
			wantError: "word cannot be empty",
		},
		{
			name:      "meaning 为空",
			patch:     WordPatch{Meaning: stringPointer("   ")},
			wantError: "meaning cannot be empty",
		},
		{
			name:      "level 太小",
			patch:     WordPatch{Level: intPointer(0)},
			wantError: "level must be between 1 and 5",
		},
		{
			name:      "level 太大",
			patch:     WordPatch{Level: intPointer(6)},
			wantError: "level must be between 1 and 5",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.patch.Normalize()
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("error = %v, want %q", err, test.wantError)
			}
		})
	}
}
