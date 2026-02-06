package entity

import "gorm.io/gorm"

func RegisterMigrator(db *gorm.DB) error {
	if err := MigrateTables(db); err != nil {
		return err
	}
	return MigrateIndexes(db)
}

func MigrateTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&Product{},
	)
}

func MigrateIndexes(db *gorm.DB) error {
	//m := db.Migrator()

	// Create unique index on slug where deleted_at IS NULL (for soft deletes)
	//if !m.HasIndex(&Category{}, "idx_categories_slug_unique") {
	//	if err := db.Exec("CREATE UNIQUE INDEX idx_categories_slug_unique ON categories(slug) WHERE deleted_at IS NULL").Error; err != nil {
	//		// Index might already exist, continue
	//	}
	//}
	return nil
}
