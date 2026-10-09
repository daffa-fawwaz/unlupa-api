package services_test

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
)

func TestActivateToFSRSOwnershipAndSharedBookAccess(t *testing.T) {
	db := setupTestPostgresDB(t)
	previousDB := config.DB
	config.DB = db
	t.Cleanup(func() { config.DB = previousDB })
	t.Run("shared book member can activate teacher item", func(t *testing.T) {
		teacherID, studentID := uuid.New(), uuid.New()
		_, _, _, teacherItem := addSharedBookFixture(t, db, teacherID, studentID, true, true)
		statusService := newActivationTestService(db)

		activated, err := statusService.ActivateToFSRS(teacherItem.ID, studentID)
		if err != nil {
			t.Fatalf("expected activation to succeed: %v", err)
		}
		if activated.OwnerID != studentID || activated.Status != entities.ItemStatusFSRSActive {
			t.Fatalf("expected student's active item, got owner=%s status=%s", activated.OwnerID, activated.Status)
		}
		// Simulate the status GET performed after refresh: load the student's
		// active rows again from the repository and match by the book content_ref.
		activeItems, err := statusService.GetItemsByStatus(studentID, entities.ItemStatusFSRSActive)
		if err != nil {
			t.Fatalf("fetch active items after activation: %v", err)
		}
		persisted := false
		for _, row := range activeItems {
			if row.ContentRef == teacherItem.ContentRef && row.ID == activated.ID && row.OwnerID == studentID {
				persisted = row.Status == entities.ItemStatusFSRSActive
				break
			}
		}
		if !persisted {
			t.Fatalf("student's activated item was not returned by status fetch: %+v", activeItems)
		}
		var unchanged entities.Item
		if err := db.First(&unchanged, "id = ?", teacherItem.ID).Error; err != nil {
			t.Fatalf("load teacher item: %v", err)
		}
		if unchanged.OwnerID != teacherID || unchanged.Status != entities.ItemStatusStart {
			t.Fatalf("teacher item was modified: owner=%s status=%s", unchanged.OwnerID, unchanged.Status)
		}
	})

	t.Run("nonmember cannot activate shared book item", func(t *testing.T) {
		teacherID, studentID := uuid.New(), uuid.New()
		_, _, _, teacherItem := addSharedBookFixture(t, db, teacherID, studentID, true, false)
		if _, err := newActivationTestService(db).ActivateToFSRS(teacherItem.ID, studentID); err == nil {
			t.Fatal("expected nonmember activation to be denied")
		}
	})

	t.Run("unshared book item cannot be activated", func(t *testing.T) {
		teacherID, studentID := uuid.New(), uuid.New()
		_, _, _, teacherItem := addSharedBookFixture(t, db, teacherID, studentID, false, true)
		if _, err := newActivationTestService(db).ActivateToFSRS(teacherItem.ID, studentID); err == nil {
			t.Fatal("expected activation of unshared book item to be denied")
		}
	})

	t.Run("other user's personal item remains denied", func(t *testing.T) {
		ownerID, callerID := uuid.New(), uuid.New()
		item := entities.Item{ID: uuid.New(), OwnerID: ownerID, SourceType: "personal", ContentRef: "personal:" + uuid.NewString(), Status: entities.ItemStatusInterval}
		if err := db.Create(&item).Error; err != nil {
			t.Fatal(err)
		}
		defer db.Delete(&item)
		if _, err := newActivationTestService(db).ActivateToFSRS(item.ID, callerID); err == nil {
			t.Fatal("expected another user's personal item to be denied")
		}
	})

	t.Run("own personal item still activates", func(t *testing.T) {
		ownerID := uuid.New()
		item := entities.Item{ID: uuid.New(), OwnerID: ownerID, SourceType: "personal", ContentRef: "personal:" + uuid.NewString(), Status: entities.ItemStatusInterval}
		if err := db.Create(&item).Error; err != nil {
			t.Fatal(err)
		}
		defer db.Delete(&item)
		activated, err := newActivationTestService(db).ActivateToFSRS(item.ID, ownerID)
		if err != nil {
			t.Fatalf("expected own item activation to succeed: %v", err)
		}
		if activated.ID != item.ID || activated.OwnerID != ownerID || activated.Status != entities.ItemStatusFSRSActive {
			t.Fatalf("unexpected activated personal item: %+v", activated)
		}
	})
}

func newActivationTestService(db *gorm.DB) *services.ItemStatusService {
	return services.NewItemStatusService(repositories.NewItemRepository(db), nil, repositories.NewClassBookRepository(db), nil)
}

func addSharedBookFixture(t *testing.T, db *gorm.DB, teacherID, studentID uuid.UUID, share, addMember bool) (entities.Book, entities.BookItem, entities.Class, entities.Item) {
	t.Helper()
	book := entities.Book{ID: uuid.New(), OwnerID: teacherID, Title: "test book", Status: entities.BookStatusDraft}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("create fixture book: %v", err)
	}
	bookItem := entities.BookItem{ID: uuid.New(), BookID: book.ID, Title: "test item", Content: "question", Answer: "answer"}
	if err := db.Create(&bookItem).Error; err != nil {
		t.Fatalf("create fixture book item: %v", err)
	}
	class := entities.Class{ID: uuid.New(), GuruID: teacherID, Name: "test class", ClassCode: uuid.NewString()[:8], Type: entities.ClassTypeBook}
	if err := db.Create(&class).Error; err != nil {
		t.Fatalf("create fixture class: %v", err)
	}
	teacherItem := entities.Item{ID: uuid.New(), OwnerID: teacherID, SourceType: "book", ContentRef: "book:" + book.ID.String() + ":item:" + bookItem.ID.String(), Status: entities.ItemStatusStart}
	if err := db.Create(&teacherItem).Error; err != nil {
		t.Fatalf("create fixture teacher item: %v", err)
	}
	if share {
		classBook := entities.ClassBook{ID: uuid.New(), ClassID: class.ID, BookID: book.ID}
		if err := db.Create(&classBook).Error; err != nil {
			t.Fatalf("share book to class: %v", err)
		}
	}
	if addMember {
		member := entities.ClassMember{ID: uuid.New(), ClassID: class.ID, UserID: studentID}
		if err := db.Create(&member).Error; err != nil {
			t.Fatalf("add student to class: %v", err)
		}
	}
	t.Cleanup(func() { cleanupSharedBookFixture(t, db, book, bookItem, class, teacherItem, studentID) })
	return book, bookItem, class, teacherItem
}

func cleanupSharedBookFixture(t *testing.T, db *gorm.DB, book entities.Book, bookItem entities.BookItem, class entities.Class, teacherItem entities.Item, studentID uuid.UUID) {
	t.Helper()
	db.Where("owner_id = ? AND content_ref = ?", studentID, teacherItem.ContentRef).Delete(&entities.Item{})
	db.Delete(&teacherItem)
	db.Where("book_id = ?", book.ID).Delete(&entities.ClassBook{})
	db.Where("class_id = ?", class.ID).Delete(&entities.ClassMember{})
	db.Delete(&class)
	db.Delete(&bookItem)
	db.Delete(&book)
}
