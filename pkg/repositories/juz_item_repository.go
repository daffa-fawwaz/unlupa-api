package repositories

import (
	"hifzhun-api/pkg/entities"

	"gorm.io/gorm"
)

type JuzItemRepository struct {
	db *gorm.DB
}

func NewJuzItemRepository(db *gorm.DB) *JuzItemRepository {
	return &JuzItemRepository{db}
}

func (r *JuzItemRepository) Create(rel *entities.JuzItem) error {
	return r.db.Create(rel).Error
}

// FindJuzIndexByItemID returns the juz index for a given item ID
func (r *JuzItemRepository) FindJuzIndexByItemID(itemID string) (int, error) {
	var result struct {
		Index int
	}
	err := r.db.
		Table("juz_items").
		Select("juzs.index").
		Joins("JOIN juzs ON juzs.id = juz_items.juz_id").
		Where("juz_items.item_id = ?", itemID).
		Scan(&result).Error
	if err != nil {
		return 0, err
	}
	return result.Index, nil
}

// FindJuzIndexByItemIDs returns a map of item_id -> juz index for multiple items
func (r *JuzItemRepository) FindJuzIndexByItemIDs(itemIDs []string) (map[string]int, error) {
	type row struct {
		ItemID string `gorm:"column:item_id"`
		Index  int    `gorm:"column:index"`
	}
	var rows []row
	err := r.db.
		Table("juz_items").
		Select("juz_items.item_id, juzs.index").
		Joins("JOIN juzs ON juzs.id = juz_items.juz_id").
		Where("juz_items.item_id IN ?", itemIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]int)
	for _, r := range rows {
		result[r.ItemID] = r.Index
	}
	return result, nil
}

// JuzInfo holds juz_id and juz_index for an item
type JuzInfo struct {
	JuzID    string
	JuzIndex int
}

// FindJuzInfoByItemIDs returns a map of item_id -> JuzInfo for multiple items
func (r *JuzItemRepository) FindJuzInfoByItemIDs(itemIDs []string) (map[string]JuzInfo, error) {
	type row struct {
		ItemID string `gorm:"column:item_id"`
		JuzID  string `gorm:"column:juz_id"`
		Index  int    `gorm:"column:index"`
	}
	var rows []row
	err := r.db.
		Table("juz_items").
		Select("juz_items.item_id, juz_items.juz_id, juzs.index").
		Joins("JOIN juzs ON juzs.id = juz_items.juz_id").
		Where("juz_items.item_id IN ?", itemIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]JuzInfo)
	for _, r := range rows {
		result[r.ItemID] = JuzInfo{JuzID: r.JuzID, JuzIndex: r.Index}
	}
	return result, nil
}

// JuzItemStatusCount holds per-status item counts for a juz
type JuzItemStatusCount struct {
	JuzID  string `gorm:"column:juz_id"`
	Status string `gorm:"column:status"`
	Count  int    `gorm:"column:count"`
}

// CountItemStatusByJuzIDs returns item counts grouped by juz_id and status
func (r *JuzItemRepository) CountItemStatusByJuzIDs(juzIDs []string) ([]JuzItemStatusCount, error) {
	var results []JuzItemStatusCount
	err := r.db.
		Table("juz_items").
		Select("juz_items.juz_id, items.status, COUNT(*) as count").
		Joins("JOIN items ON items.id = juz_items.item_id").
		Where("juz_items.juz_id IN ?", juzIDs).
		Group("juz_items.juz_id, items.status").
		Scan(&results).Error
	return results, err
}
