package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
)

type JuzPagesDetailResponse struct {
	Juz           entities.QuranJuzCatalog              `json:"juz"`
	TotalPages    int                                   `json:"total_pages"`
	ActivePages   int                                   `json:"active_pages"`
	MasteredPages int                                   `json:"mastered_pages"`
	DueToday      int                                   `json:"due_today"`
	Pages         []repositories.PageCatalogWithUserStatus `json:"pages"`
}

type ActivatePageResult struct {
	Item          *entities.Item `json:"item"`
	MushafPage    int            `json:"mushaf_page"`
	JuzNumber     int            `json:"juz_number"`
	Status        string         `json:"status"`
	AlreadyActive bool           `json:"already_active"`
	Message       string         `json:"message"`
}

type QuranCatalogService struct {
	catalogRepo repositories.QuranCatalogRepository
	itemRepo    *repositories.ItemRepository
	juzRepo     *repositories.JuzRepository
	juzItemRepo *repositories.JuzItemRepository
	db          *gorm.DB
}

func NewQuranCatalogService(
	catalogRepo repositories.QuranCatalogRepository,
	itemRepo *repositories.ItemRepository,
	juzRepo *repositories.JuzRepository,
	juzItemRepo *repositories.JuzItemRepository,
	db *gorm.DB,
) *QuranCatalogService {
	return &QuranCatalogService{
		catalogRepo: catalogRepo,
		itemRepo:    itemRepo,
		juzRepo:     juzRepo,
		juzItemRepo: juzItemRepo,
		db:          db,
	}
}

// GetAllJuzs returns static Juz 1-30 with aggregated active, mastered, and due counts for the user
func (s *QuranCatalogService) GetAllJuzs(ctx context.Context, userID uuid.UUID) ([]repositories.JuzCatalogWithUserStats, error) {
	now := time.Now().In(config.AppLocation)
	return s.catalogRepo.FindAllJuzWithUserStats(ctx, userID, now)
}

// GetJuzPages returns all pages within a Juz along with activation and review statuses for the user
func (s *QuranCatalogService) GetJuzPages(ctx context.Context, juzNumber int, userID uuid.UUID) (*JuzPagesDetailResponse, error) {
	if juzNumber < 1 || juzNumber > 30 {
		return nil, errors.New("invalid juz number: must be between 1 and 30")
	}

	juz, err := s.catalogRepo.FindJuzByNumber(ctx, juzNumber)
	if err != nil {
		return nil, fmt.Errorf("juz %d not found in catalog: %w", juzNumber, err)
	}

	now := time.Now().In(config.AppLocation)
	pages, err := s.catalogRepo.FindJuzPagesWithUserStatus(ctx, juzNumber, userID, now)
	if err != nil {
		return nil, err
	}

	activePages := 0
	masteredPages := 0
	dueToday := 0

	for _, p := range pages {
		if p.IsActivated {
			activePages++
			if p.ActivationStatus == "mastered" {
				masteredPages++
			}
			if p.IsDue {
				dueToday++
			}
		}
	}

	return &JuzPagesDetailResponse{
		Juz:           *juz,
		TotalPages:    juz.TotalPages,
		ActivePages:   activePages,
		MasteredPages: masteredPages,
		DueToday:      dueToday,
		Pages:         pages,
	}, nil
}

// ActivatePage creates a user Item and personal JuzItem for a specific Quran catalog page (idempotently & atomically)
func (s *QuranCatalogService) ActivatePage(ctx context.Context, userID uuid.UUID, mushafPage int) (*ActivatePageResult, error) {
	if mushafPage < 1 || mushafPage > 604 {
		return nil, errors.New("invalid mushaf page number: must be between 1 and 604")
	}

	// 1. Fetch catalog page info
	catalogPage, err := s.catalogRepo.FindPageByMushafPage(ctx, mushafPage)
	if err != nil {
		return nil, fmt.Errorf("catalog page %d not found: %w", mushafPage, err)
	}

	contentRef := fmt.Sprintf("page:%d", mushafPage)
	var activatedItem *entities.Item
	alreadyActive := false

	// 2. Transaction for atomic activation
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// A. Check if user already has an item for this page
		var existingItem entities.Item
		err := tx.Where("owner_id = ? AND source_type = ? AND content_ref = ?", userID, "quran", contentRef).
			First(&existingItem).Error

		if err == nil {
			// Item already exists - idempotent return
			alreadyActive = true
			activatedItem = &existingItem
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// B. Find or create personal Juz for user (ClassID IS NULL)
		var personalJuz entities.Juz
		err = tx.Where("user_id = ? AND index = ? AND class_id IS NULL", userID, catalogPage.JuzNumber).
			First(&personalJuz).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			personalJuz = entities.Juz{
				UserID:   userID,
				ClassID:  nil,
				Index:    catalogPage.JuzNumber,
				IsActive: true,
				IsDone:   false,
			}
			if err := tx.Create(&personalJuz).Error; err != nil {
				return fmt.Errorf("failed to create personal juz: %w", err)
			}
		} else if err != nil {
			return err
		}

		// C. Create new Item in initial status "menghafal"
		newItem := entities.Item{
			OwnerID:    userID,
			SourceType: "quran",
			ContentRef: contentRef,
			Status:     entities.ItemStatusMenghafal, // Standard initial status
			Stability:  0.0,
			Difficulty: 5.0,
		}

		if err := tx.Create(&newItem).Error; err != nil {
			return fmt.Errorf("failed to create user item: %w", err)
		}

		// D. Create JuzItem link
		juzItem := entities.JuzItem{
			ID:     uuid.New(),
			JuzID:  personalJuz.ID,
			ItemID: newItem.ID,
		}
		if err := tx.Create(&juzItem).Error; err != nil {
			return fmt.Errorf("failed to link item to juz: %w", err)
		}

		activatedItem = &newItem
		return nil
	})

	if err != nil {
		return nil, err
	}

	msg := "Halaman berhasil diaktivasi ke hafalan"
	if alreadyActive {
		msg = "Halaman sudah aktif sebelumnya"
	}

	return &ActivatePageResult{
		Item:          activatedItem,
		MushafPage:    mushafPage,
		JuzNumber:     catalogPage.JuzNumber,
		Status:        activatedItem.Status,
		AlreadyActive: alreadyActive,
		Message:       msg,
	}, nil
}
