package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
)

func TestQuranCatalogService(t *testing.T) {
	db := setupTestPostgresDB(t)

	// Ensure catalog is migrated and seeded
	_ = db.AutoMigrate(&entities.QuranJuzCatalog{}, &entities.QuranPageCatalog{}, &entities.Item{}, &entities.Juz{}, &entities.JuzItem{})
	if err := config.SeedQuranCatalog(db); err != nil {
		t.Fatalf("Failed to seed quran catalog: %v", err)
	}

	catalogRepo := repositories.NewQuranCatalogRepository(db)
	itemRepo := repositories.NewItemRepository(db)
	juzRepo := repositories.NewJuzRepository(db)
	juzItemRepo := repositories.NewJuzItemRepository(db)
	catalogSvc := services.NewQuranCatalogService(catalogRepo, itemRepo, juzRepo, juzItemRepo, db)

	ctx := context.Background()
	testUserID := uuid.New()

	// Clean up after test
	defer func() {
		_ = db.Where("owner_id = ?", testUserID).Delete(&entities.Item{}).Error
		_ = db.Where("user_id = ?", testUserID).Delete(&entities.Juz{}).Error
	}()

	// -------------------------------------------------------------------------
	// 1. Get 30 Juz
	// -------------------------------------------------------------------------
	t.Run("1. GetAllJuzs returns all 30 Juz", func(t *testing.T) {
		juzs, err := catalogSvc.GetAllJuzs(ctx, testUserID)
		if err != nil {
			t.Fatalf("GetAllJuzs failed: %v", err)
		}
		if len(juzs) != 30 {
			t.Errorf("Expected 30 Juz, got %d", len(juzs))
		}
		for i, j := range juzs {
			if j.JuzNumber != i+1 {
				t.Errorf("Expected Juz %d at index %d, got %d", i+1, i, j.JuzNumber)
			}
		}
	})

	// -------------------------------------------------------------------------
	// 2. Get Juz 1 pages (21 pages, Page 1 to 21)
	// -------------------------------------------------------------------------
	t.Run("2. GetJuzPages Juz 1 has 21 pages", func(t *testing.T) {
		res, err := catalogSvc.GetJuzPages(ctx, 1, testUserID)
		if err != nil {
			t.Fatalf("GetJuzPages(1) failed: %v", err)
		}
		if res.TotalPages != 21 {
			t.Errorf("Expected 21 pages for Juz 1, got %d", res.TotalPages)
		}
		if len(res.Pages) != 21 {
			t.Errorf("Expected 21 page records for Juz 1, got %d", len(res.Pages))
		}
		if res.Pages[0].MushafPage != 1 || res.Pages[0].PageNumberInJuz != 1 {
			t.Errorf("First page of Juz 1 mismatch: got mushaf_page=%d, page_in_juz=%d", res.Pages[0].MushafPage, res.Pages[0].PageNumberInJuz)
		}
		if res.Pages[20].MushafPage != 21 || res.Pages[20].PageNumberInJuz != 21 {
			t.Errorf("Last page of Juz 1 mismatch: got mushaf_page=%d, page_in_juz=%d", res.Pages[20].MushafPage, res.Pages[20].PageNumberInJuz)
		}
	})

	// -------------------------------------------------------------------------
	// 3. Get Juz 30 pages (23 pages, Page 582 to 604)
	// -------------------------------------------------------------------------
	t.Run("3. GetJuzPages Juz 30 has 23 pages", func(t *testing.T) {
		res, err := catalogSvc.GetJuzPages(ctx, 30, testUserID)
		if err != nil {
			t.Fatalf("GetJuzPages(30) failed: %v", err)
		}
		if res.TotalPages != 23 {
			t.Errorf("Expected 23 pages for Juz 30, got %d", res.TotalPages)
		}
		if len(res.Pages) != 23 {
			t.Errorf("Expected 23 page records for Juz 30, got %d", len(res.Pages))
		}
		if res.Pages[0].MushafPage != 582 || res.Pages[0].PageNumberInJuz != 1 {
			t.Errorf("First page of Juz 30 mismatch: got mushaf_page=%d, page_in_juz=%d", res.Pages[0].MushafPage, res.Pages[0].PageNumberInJuz)
		}
		if res.Pages[22].MushafPage != 604 || res.Pages[22].PageNumberInJuz != 23 {
			t.Errorf("Last page of Juz 30 mismatch: got mushaf_page=%d, page_in_juz=%d", res.Pages[22].MushafPage, res.Pages[22].PageNumberInJuz)
		}
	})

	// -------------------------------------------------------------------------
	// 4 & 5. New user sees 30 catalog Juz and 0 active pages
	// -------------------------------------------------------------------------
	newUser := uuid.New()
	t.Run("4 & 5. New user sees 30 catalog Juz and has 0 active pages", func(t *testing.T) {
		juzs, err := catalogSvc.GetAllJuzs(ctx, newUser)
		if err != nil {
			t.Fatalf("GetAllJuzs failed: %v", err)
		}
		if len(juzs) != 30 {
			t.Errorf("Expected 30 Juz, got %d", len(juzs))
		}
		totalActive := 0
		for _, j := range juzs {
			totalActive += j.ActivePages
		}
		if totalActive != 0 {
			t.Errorf("Expected 0 active pages for new user, got %d", totalActive)
		}
	})

	// -------------------------------------------------------------------------
	// 6. Activate page 1
	// -------------------------------------------------------------------------
	t.Run("6. Activate page 1 for user", func(t *testing.T) {
		res, err := catalogSvc.ActivatePage(ctx, testUserID, 1)
		if err != nil {
			t.Fatalf("ActivatePage(1) failed: %v", err)
		}
		if res.Item == nil {
			t.Fatal("Expected created item, got nil")
		}
		if res.Item.Status != entities.ItemStatusMenghafal {
			t.Errorf("Expected initial status 'menghafal', got '%s'", res.Item.Status)
		}
		if res.Item.ContentRef != "page:1" {
			t.Errorf("Expected content_ref 'page:1', got '%s'", res.Item.ContentRef)
		}
		if res.AlreadyActive {
			t.Errorf("Expected already_active = false on first activation")
		}
	})

	// -------------------------------------------------------------------------
	// 7. Activate page 1 twice -> idempotent, no duplicate
	// -------------------------------------------------------------------------
	t.Run("7. Activate page 1 twice is idempotent", func(t *testing.T) {
		res, err := catalogSvc.ActivatePage(ctx, testUserID, 1)
		if err != nil {
			t.Fatalf("ActivatePage(1) second call failed: %v", err)
		}
		if !res.AlreadyActive {
			t.Errorf("Expected already_active = true on second activation")
		}

		var count int64
		db.Model(&entities.Item{}).Where("owner_id = ? AND content_ref = ?", testUserID, "page:1").Count(&count)
		if count != 1 {
			t.Errorf("Expected exactly 1 item in database for page 1, got %d", count)
		}
	})

	// -------------------------------------------------------------------------
	// 8. Activate page 604
	// -------------------------------------------------------------------------
	t.Run("8. Activate page 604 in Juz 30", func(t *testing.T) {
		res, err := catalogSvc.ActivatePage(ctx, testUserID, 604)
		if err != nil {
			t.Fatalf("ActivatePage(604) failed: %v", err)
		}
		if res.JuzNumber != 30 {
			t.Errorf("Expected Juz 30 for page 604, got %d", res.JuzNumber)
		}
		if res.Item.ContentRef != "page:604" {
			t.Errorf("Expected content_ref 'page:604', got '%s'", res.Item.ContentRef)
		}
	})

	// -------------------------------------------------------------------------
	// 9. Invalid page numbers return error
	// -------------------------------------------------------------------------
	t.Run("9. Invalid page numbers rejected", func(t *testing.T) {
		if _, err := catalogSvc.ActivatePage(ctx, testUserID, 0); err == nil {
			t.Error("Expected error for page 0, got nil")
		}
		if _, err := catalogSvc.ActivatePage(ctx, testUserID, 605); err == nil {
			t.Error("Expected error for page 605, got nil")
		}
	})

	// -------------------------------------------------------------------------
	// 10. Invalid Juz number returns error
	// -------------------------------------------------------------------------
	t.Run("10. Invalid Juz numbers rejected", func(t *testing.T) {
		if _, err := catalogSvc.GetJuzPages(ctx, 0, testUserID); err == nil {
			t.Error("Expected error for Juz 0, got nil")
		}
		if _, err := catalogSvc.GetJuzPages(ctx, 31, testUserID); err == nil {
			t.Error("Expected error for Juz 31, got nil")
		}
	})

	// -------------------------------------------------------------------------
	// 11 & 12. Personal Juz has ClassID == nil and JuzItem is linked properly
	// -------------------------------------------------------------------------
	t.Run("11 & 12. Personal Juz has ClassID nil and JuzItem is linked", func(t *testing.T) {
		var personalJuz entities.Juz
		err := db.Where("user_id = ? AND index = 1", testUserID).First(&personalJuz).Error
		if err != nil {
			t.Fatalf("Personal Juz 1 not found: %v", err)
		}
		if personalJuz.ClassID != nil {
			t.Errorf("Expected personal Juz ClassID to be nil, got %v", personalJuz.ClassID)
		}

		var juzItem entities.JuzItem
		err = db.Where("juz_id = ?", personalJuz.ID).First(&juzItem).Error
		if err != nil {
			t.Fatalf("JuzItem link not found: %v", err)
		}

		var item entities.Item
		err = db.Where("id = ?", juzItem.ItemID).First(&item).Error
		if err != nil {
			t.Fatalf("Linked item not found: %v", err)
		}
		if item.ContentRef != "page:1" {
			t.Errorf("Expected linked item content_ref 'page:1', got '%s'", item.ContentRef)
		}
	})

	// -------------------------------------------------------------------------
	// 13 & 14. Existing Quran Item and legacy surah content_ref remain unchanged
	// -------------------------------------------------------------------------
	t.Run("13 & 14. Legacy surah content_ref is preserved and compatible", func(t *testing.T) {
		legacyItem := entities.Item{
			ID:         uuid.New(),
			OwnerID:    testUserID,
			SourceType: "quran",
			ContentRef: "surah:78:1-5",
			Status:     entities.ItemStatusMenghafal,
		}
		if err := db.Create(&legacyItem).Error; err != nil {
			t.Fatalf("Failed to create legacy item: %v", err)
		}
		defer db.Delete(&legacyItem)

		var fetched entities.Item
		if err := db.Where("id = ?", legacyItem.ID).First(&fetched).Error; err != nil {
			t.Fatalf("Failed to fetch legacy item: %v", err)
		}
		if fetched.ContentRef != "surah:78:1-5" {
			t.Errorf("Legacy content_ref modified unexpectedly: %s", fetched.ContentRef)
		}
	})

	// -------------------------------------------------------------------------
	// 15. Classroom item behavior remains isolated (ClassID != nil)
	// -------------------------------------------------------------------------
	t.Run("15. Classroom item behavior isolated with ClassID != nil", func(t *testing.T) {
		classID := uuid.New()
		classJuz := entities.Juz{
			ID:       uuid.New(),
			UserID:   testUserID,
			ClassID:  &classID,
			Index:    1,
			IsActive: true,
		}
		if err := db.Create(&classJuz).Error; err != nil {
			t.Fatalf("Failed to create class juz: %v", err)
		}
		defer db.Delete(&classJuz)

		classItem := entities.Item{
			ID:         uuid.New(),
			OwnerID:    testUserID,
			SourceType: "quran",
			ContentRef: "page:1",
			Status:     entities.ItemStatusMenghafal,
		}
		if err := db.Create(&classItem).Error; err != nil {
			t.Fatalf("Failed to create class item: %v", err)
		}
		defer db.Delete(&classItem)

		classJuzItem := entities.JuzItem{
			ID:     uuid.New(),
			JuzID:  classJuz.ID,
			ItemID: classItem.ID,
		}
		if err := db.Create(&classJuzItem).Error; err != nil {
			t.Fatalf("Failed to create class juz_item: %v", err)
		}
		defer db.Delete(&classJuzItem)

		// Verify that classroom juz retains ClassID != nil
		var fetched entities.Juz
		if err := db.Where("id = ?", classJuz.ID).First(&fetched).Error; err != nil {
			t.Fatalf("Failed to fetch class juz: %v", err)
		}
		if fetched.ClassID == nil || *fetched.ClassID != classID {
			t.Errorf("ClassID lost or modified: got %v, expected %v", fetched.ClassID, classID)
		}
	})
}
