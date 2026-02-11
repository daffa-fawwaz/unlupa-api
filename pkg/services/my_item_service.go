package services

import (
	"strings"

	"github.com/google/uuid"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
)

// ==================== Response Types ====================

type MyItemDetail struct {
	ItemID      uuid.UUID `json:"item_id"`
	ContentRef  string    `json:"content_ref"`
	Status      string    `json:"status"`
	ReviewCount int       `json:"review_count"`
	CreatedAt   string    `json:"created_at"`
}

type QuranItemDetail struct {
	MyItemDetail
}

type BookItemDetail struct {
	MyItemDetail
	BookItemTitle string `json:"book_item_title,omitempty"`
}

type QuranGroup struct {
	JuzIndex  int               `json:"juz_index"`
	JuzID     string            `json:"juz_id"`
	ItemCount int               `json:"item_count"`
	Items     []QuranItemDetail `json:"items"`
}

type BookGroup struct {
	BookID     string           `json:"book_id"`
	BookTitle  string           `json:"book_title"`
	CoverImage string          `json:"cover_image,omitempty"`
	ItemCount  int              `json:"item_count"`
	Items      []BookItemDetail `json:"items"`
}

type MyItemsQuranResponse struct {
	Type   string       `json:"type"`
	Groups []QuranGroup `json:"groups"`
}

type MyItemsBookResponse struct {
	Type   string      `json:"type"`
	Groups []BookGroup `json:"groups"`
}

// ==================== Service ====================

type MyItemService struct {
	itemRepo     *repositories.ItemRepository
	juzItemRepo  *repositories.JuzItemRepository
	bookRepo     repositories.BookRepository
	bookItemRepo repositories.BookItemRepository
}

func NewMyItemService(
	itemRepo *repositories.ItemRepository,
	juzItemRepo *repositories.JuzItemRepository,
	bookRepo repositories.BookRepository,
	bookItemRepo repositories.BookItemRepository,
) *MyItemService {
	return &MyItemService{
		itemRepo:     itemRepo,
		juzItemRepo:  juzItemRepo,
		bookRepo:     bookRepo,
		bookItemRepo: bookItemRepo,
	}
}

func (s *MyItemService) GetMyQuranItems(userID uuid.UUID) (*MyItemsQuranResponse, error) {
	// Fetch all quran items
	items, err := s.itemRepo.FindByOwnerAndSourceType(userID, "quran")
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return &MyItemsQuranResponse{Type: "quran", Groups: []QuranGroup{}}, nil
	}

	// Collect item IDs for juz lookup
	itemIDs := make([]string, len(items))
	for i, item := range items {
		itemIDs[i] = item.ID.String()
	}

	// Batch fetch juz info
	juzInfoMap, err := s.juzItemRepo.FindJuzInfoByItemIDs(itemIDs)
	if err != nil {
		juzInfoMap = make(map[string]repositories.JuzInfo)
	}

	// Group items by juz
	juzGroupMap := make(map[int]*QuranGroup)
	var juzOrder []int

	for _, item := range items {
		info := juzInfoMap[item.ID.String()]
		juzIdx := info.JuzIndex

		if _, exists := juzGroupMap[juzIdx]; !exists {
			juzGroupMap[juzIdx] = &QuranGroup{
				JuzIndex: juzIdx,
				JuzID:    info.JuzID,
				Items:    []QuranItemDetail{},
			}
			juzOrder = append(juzOrder, juzIdx)
		}

		juzGroupMap[juzIdx].Items = append(juzGroupMap[juzIdx].Items, QuranItemDetail{
			MyItemDetail: MyItemDetail{
				ItemID:      item.ID,
				ContentRef:  item.ContentRef,
				Status:      item.Status,
				ReviewCount: item.ReviewCount,
				CreatedAt:   item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			},
		})
	}

	// Build ordered groups
	groups := make([]QuranGroup, 0, len(juzOrder))
	for _, idx := range juzOrder {
		g := juzGroupMap[idx]
		g.ItemCount = len(g.Items)
		groups = append(groups, *g)
	}

	return &MyItemsQuranResponse{Type: "quran", Groups: groups}, nil
}

func (s *MyItemService) GetMyBookItems(userID uuid.UUID) (*MyItemsBookResponse, error) {
	// Fetch all book items
	items, err := s.itemRepo.FindByOwnerAndSourceType(userID, "book")
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return &MyItemsBookResponse{Type: "book", Groups: []BookGroup{}}, nil
	}

	// Parse content_ref to extract book IDs and book_item IDs
	// Format: book:BOOK_ID:item:BOOK_ITEM_ID
	type parsedRef struct {
		BookID     string
		BookItemID string
	}

	parsedRefs := make(map[uuid.UUID]parsedRef)
	bookIDSet := make(map[string]bool)
	bookItemIDSet := make(map[string]bool)

	for _, item := range items {
		parts := strings.Split(item.ContentRef, ":")
		if len(parts) == 4 && parts[0] == "book" && parts[2] == "item" {
			ref := parsedRef{BookID: parts[1], BookItemID: parts[3]}
			parsedRefs[item.ID] = ref
			bookIDSet[ref.BookID] = true
			bookItemIDSet[ref.BookItemID] = true
		}
	}

	// Batch fetch books
	bookMap := make(map[string]*entities.Book)
	for bookID := range bookIDSet {
		book, err := s.bookRepo.FindByID(bookID)
		if err == nil {
			bookMap[bookID] = book
		}
	}

	// Batch fetch book items for titles
	bookItemTitleMap := make(map[string]string)
	for bookItemID := range bookItemIDSet {
		bookItem, err := s.bookItemRepo.FindByID(bookItemID)
		if err == nil {
			bookItemTitleMap[bookItemID] = bookItem.Title
		}
	}

	// Group items by book
	bookGroupMap := make(map[string]*BookGroup)
	var bookOrder []string

	for _, item := range items {
		ref, ok := parsedRefs[item.ID]
		if !ok {
			continue
		}

		if _, exists := bookGroupMap[ref.BookID]; !exists {
			bookTitle := ""
			coverImage := ""
			if b, ok := bookMap[ref.BookID]; ok {
				bookTitle = b.Title
				coverImage = b.CoverImage
			}
			bookGroupMap[ref.BookID] = &BookGroup{
				BookID:     ref.BookID,
				BookTitle:  bookTitle,
				CoverImage: coverImage,
				Items:      []BookItemDetail{},
			}
			bookOrder = append(bookOrder, ref.BookID)
		}

		bookGroupMap[ref.BookID].Items = append(bookGroupMap[ref.BookID].Items, BookItemDetail{
			MyItemDetail: MyItemDetail{
				ItemID:      item.ID,
				ContentRef:  item.ContentRef,
				Status:      item.Status,
				ReviewCount: item.ReviewCount,
				CreatedAt:   item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			},
			BookItemTitle: bookItemTitleMap[ref.BookItemID],
		})
	}

	// Build ordered groups
	groups := make([]BookGroup, 0, len(bookOrder))
	for _, id := range bookOrder {
		g := bookGroupMap[id]
		g.ItemCount = len(g.Items)
		groups = append(groups, *g)
	}

	return &MyItemsBookResponse{Type: "book", Groups: groups}, nil
}
