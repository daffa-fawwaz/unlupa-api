package entities

// QuranJuzCatalog represents static catalog metadata for Juz 1 to 30
type QuranJuzCatalog struct {
	JuzNumber  int    `gorm:"primaryKey;type:int" json:"juz_number"`
	NameAr     string `gorm:"type:varchar(50);not null" json:"name_ar"`
	NameEn     string `gorm:"type:varchar(50);not null" json:"name_en"`
	StartPage  int    `gorm:"not null" json:"start_page"`
	EndPage    int    `gorm:"not null" json:"end_page"`
	TotalPages int    `gorm:"not null" json:"total_pages"`
	StartSurah string `gorm:"type:varchar(100);not null" json:"start_surah"`
	EndSurah   string `gorm:"type:varchar(100);not null" json:"end_surah"`
	SurahSpan  string `gorm:"type:varchar(150);not null" json:"surah_span"`
	AyahSpan   string `gorm:"type:varchar(150);not null" json:"ayah_span"`
}

func (QuranJuzCatalog) TableName() string {
	return "quran_juz_catalogs"
}

// QuranPageCatalog represents static catalog metadata for standard Madinah Mushaf Pages 1 to 604
type QuranPageCatalog struct {
	MushafPage       int    `gorm:"primaryKey;type:int" json:"mushaf_page"` // 1 - 604
	JuzNumber        int    `gorm:"type:int;not null;index:idx_page_juz" json:"juz_number"` // 1 - 30
	PageNumberInJuz  int    `gorm:"type:int;not null" json:"page_number_in_juz"` // 1 - 21/20/23
	SurahNameEn      string `gorm:"type:varchar(100);not null" json:"surah_name_en"`
	SurahNameAr      string `gorm:"type:varchar(100);not null" json:"surah_name_ar"`
	SurahNumber      int    `gorm:"type:int;not null" json:"surah_number"`
	AyahRange        string `gorm:"type:varchar(50);not null" json:"ayah_range"`
	StartVerseKey    string `gorm:"type:varchar(20);not null" json:"start_verse_key"`
	EndVerseKey      string `gorm:"type:varchar(20);not null" json:"end_verse_key"`
	ContentRef       string `gorm:"type:varchar(50);not null;index:idx_page_content_ref" json:"content_ref"` // e.g. "page:1", "page:582"
}

func (QuranPageCatalog) TableName() string {
	return "quran_page_catalogs"
}
