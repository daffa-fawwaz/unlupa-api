package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SurahInfo holds data from surah.json
type SurahInfo struct {
	Index string `json:"index"`
	Title string `json:"title"`
	Count int    `json:"count"`
}

// QuranValidator validates content_ref against Quran data
type QuranValidator struct {
	surahs map[int]SurahInfo // key: surah number (1-114)
}

// NewQuranValidator loads surah data from JSON file
func NewQuranValidator(surahJSONPath string) (*QuranValidator, error) {
	data, err := os.ReadFile(surahJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read surah.json: %w", err)
	}

	var surahList []SurahInfo
	if err := json.Unmarshal(data, &surahList); err != nil {
		return nil, fmt.Errorf("failed to parse surah.json: %w", err)
	}

	surahs := make(map[int]SurahInfo)
	for _, s := range surahList {
		idx, _ := strconv.Atoi(s.Index)
		surahs[idx] = s
	}

	return &QuranValidator{surahs: surahs}, nil
}

// ValidateContentRef validates content_ref format: "surah:78:1-5"
func (v *QuranValidator) ValidateContentRef(mode, contentRef string) error {
	if mode != "surah" && mode != "page" {
		return errors.New("invalid mode: must be 'surah' or 'page'")
	}

	if mode == "surah" {
		return v.validateSurahRef(contentRef)
	}

	// TODO: implement page validation if needed
	return nil
}

// validateSurahRef validates format "surah:78:1-5"
func (v *QuranValidator) validateSurahRef(contentRef string) error {
	// Expected format: surah:SURAH_NUM:START-END
	parts := strings.Split(contentRef, ":")
	if len(parts) != 3 {
		return errors.New("invalid content_ref format, expected: surah:SURAH_NUM:START-END")
	}

	if parts[0] != "surah" {
		return errors.New("content_ref must start with 'surah:'")
	}

	// Parse surah number
	surahNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.New("invalid surah number")
	}

	// Check if surah exists
	surah, exists := v.surahs[surahNum]
	if !exists {
		return fmt.Errorf("surah %d not found (valid: 1-114)", surahNum)
	}

	// Parse verse range (e.g., "1-5")
	verseRange := strings.Split(parts[2], "-")
	if len(verseRange) != 2 {
		return errors.New("invalid verse range format, expected: START-END")
	}

	startVerse, err := strconv.Atoi(verseRange[0])
	if err != nil {
		return errors.New("invalid start verse number")
	}

	endVerse, err := strconv.Atoi(verseRange[1])
	if err != nil {
		return errors.New("invalid end verse number")
	}

	// Validate verse range
	if startVerse < 1 {
		return errors.New("start verse must be at least 1")
	}

	if startVerse > endVerse {
		return errors.New("start verse cannot be greater than end verse")
	}

	if endVerse > surah.Count {
		return fmt.Errorf("verse %d exceeds surah %s verse count (%d)", endVerse, surah.Title, surah.Count)
	}

	return nil
}
