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

	// Check if juz already exists for this user
	existing, err := s.juzRepo.FindByUserAndIndex(userID.String(), index)
	if err == nil && existing != nil {
		return nil, errors.New("juz already exists")
	}

	juz := &entities.Juz{
		UserID: userID,
		Index:  index,
	}

	err = s.juzRepo.Create(juz)
	return juz, err
}

// CreateHafalanResult represents the result of creating a hafalan item
type CreateHafalanResult struct {
	ItemID                 uuid.UUID `json:"item_id"`
	JuzID                  uuid.UUID `json:"juz_id"`
	SourceType             string    `json:"source_type"`
	ContentRef             string    `json:"content_ref"`
	Status                 string    `json:"status"`
	EstimatedReviewSeconds int       `json:"estimated_review_seconds"`
}

func (s *HafalanService) AddItemToJuz(
	userID uuid.UUID,
	juzID uuid.UUID,
	mode string, // surah | page
	contentRef string, // surah:78:1-5 | page:582
	estimateVal int, // optional
	estimateUnit string, // "seconds" | "minutes"
) (*CreateHafalanResult, error) {
	// Validate content_ref against Quran data
	if err := s.quranValidator.ValidateContentRef(mode, contentRef); err != nil {
		return nil, err
	}

	// Normalize estimation into seconds
	estSeconds := 0
	if estimateVal > 0 {
		switch estimateUnit {
		case "minutes", "minute", "min", "m":
			estSeconds = estimateVal * 60
		default:
			estSeconds = estimateVal
		}
		if estSeconds < 0 {
			estSeconds = 0
		}
	}

	item := &entities.Item{
		OwnerID:                userID,
		SourceType:             "quran",
		ContentRef:             contentRef,
		EstimatedReviewSeconds: estSeconds,
	}

	if err := s.itemRepo.Create(item); err != nil {
		return nil, err
	}

	rel := &entities.JuzItem{
		ID:     uuid.New(),
		JuzID:  juzID,
		ItemID: item.ID,
	}

	if err := s.juzItemRepo.Create(rel); err != nil {
		return nil, err
	}

	return &CreateHafalanResult{
		ItemID:                 item.ID,
		JuzID:                  juzID,
		SourceType:             item.SourceType,
		ContentRef:             item.ContentRef,
		Status:                 item.Status,
		EstimatedReviewSeconds: item.EstimatedReviewSeconds,
	}, nil
}
