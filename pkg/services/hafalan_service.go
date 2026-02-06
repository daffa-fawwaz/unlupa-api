package services

import (
	"errors"

	"github.com/google/uuid"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
)

type HafalanService struct {
	juzRepo        *repositories.JuzRepository
	itemRepo       *repositories.ItemRepository
	juzItemRepo    *repositories.JuzItemRepository
	quranValidator *QuranValidator
}

func NewHafalanService(
	juzRepo *repositories.JuzRepository,
	itemRepo *repositories.ItemRepository,
	juzItemRepo *repositories.JuzItemRepository,
	quranValidator *QuranValidator,
) *HafalanService {
	return &HafalanService{juzRepo, itemRepo, juzItemRepo, quranValidator}
}

func (s *HafalanService) CreateJuz(userID uuid.UUID, index int) (*entities.Juz, error) {
	if index < 1 || index > 30 {
		return nil, errors.New("invalid juz index")
	}

	juz := &entities.Juz{
		UserID: userID,
		Index:  index,
	}

	err := s.juzRepo.Create(juz)
	return juz, err
}

func (s *HafalanService) AddItemToJuz(
	userID uuid.UUID,
	juzID uuid.UUID,
	mode string,       // surah | page
	contentRef string, // surah:78:1-5 | page:582
) error {
	// Validate content_ref against Quran data
	if err := s.quranValidator.ValidateContentRef(mode, contentRef); err != nil {
		return err
	}

	item := &entities.Item{
		OwnerID:    userID,
		SourceType: "quran",
		ContentRef: contentRef,
	}

	if err := s.itemRepo.Create(item); err != nil {
		return err
	}

	rel := &entities.JuzItem{
		ID:     uuid.New(),
		JuzID:  juzID,
		ItemID: item.ID,
	}

	return s.juzItemRepo.Create(rel)
}


