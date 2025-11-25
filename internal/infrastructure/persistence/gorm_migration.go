package persistence

import (
	"card_game/internal/core/port"
	"card_game/internal/infrastructure/persistence/model"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// GormMigration GORMマイグレーション情報
type GormMigration struct {
	Version     string    `gorm:"primaryKey;type:varchar(255)"`
	Description string    `gorm:"type:varchar(500)"`
	AppliedAt   time.Time `gorm:"autoCreateTime"`
}

// TableName テーブル名を指定
func (GormMigration) TableName() string {
	return "gorm_migrations"
}

// RunGormMigrations GORMのAutoMigrateを使用してマイグレーションを実行
func RunGormMigrations(db *gorm.DB, logger port.Logger) error {
	// マイグレーション管理テーブルを作成
	if err := db.AutoMigrate(&GormMigration{}); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// すべてのモデルを定義（時系列順）
	models := []struct {
		version     string
		description string
		migrateFunc func(*gorm.DB) error
	}{
		{
			version:     "001",
			description: "Create users table",
			migrateFunc: func(db *gorm.DB) error {
				return db.AutoMigrate(&model.UserModel{})
			},
		},
		{
			version:     "002",
			description: "Create cards and card_traits tables",
			migrateFunc: func(db *gorm.DB) error {
				return db.AutoMigrate(
					&model.CardModel{},
					&model.CardTraitModel{},
				)
			},
		},
		{
			version:     "003",
			description: "Create card_effects and effect_definitions tables",
			migrateFunc: func(db *gorm.DB) error {
				return db.AutoMigrate(
					&model.CardEffectModel{},
					&model.EffectDefinitionModel{},
				)
			},
		},
		{
			version:     "004",
			description: "Create effect_chain_nodes and concrete node tables",
			migrateFunc: func(db *gorm.DB) error {
				return db.AutoMigrate(
					&model.EffectChainNodeModel{},
					&model.SequentialNodeModel{},
					&model.ParallelNodeModel{},
					&model.ParallelNodeChildModel{},
					&model.IfElseNodeModel{},
					&model.RepeatNodeModel{},
					&model.ForEachNodeModel{},
				)
			},
		},
		{
			version:     "005",
			description: "Create atomic_effects table",
			migrateFunc: func(db *gorm.DB) error {
				return db.AutoMigrate(&model.AtomicEffectModel{})
			},
		},
		{
			version:     "006",
			description: "Create target_selectors and target_filters tables",
			migrateFunc: func(db *gorm.DB) error {
				return db.AutoMigrate(
					&model.TargetSelectorModel{},
					&model.TargetFilterModel{},
					&model.TargetFilterTraitModel{},
				)
			},
		},
		{
			version:     "007",
			description: "Create conditions table",
			migrateFunc: func(db *gorm.DB) error {
				return db.AutoMigrate(&model.ConditionModel{})
			},
		},
		{
			version:     "008",
			description: "Remove name column from effect_definitions table",
			migrateFunc: func(db *gorm.DB) error {
				// nameカラムが存在する場合のみ削除
				if db.Migrator().HasColumn(&model.EffectDefinitionModel{}, "name") {
					if err := db.Exec("ALTER TABLE effect_definitions DROP COLUMN name").Error; err != nil {
						return fmt.Errorf("failed to drop name column: %w", err)
					}
				}
				return nil
			},
		},
		{
			version:     "009",
			description: "Create decks and deck_cards tables",
			migrateFunc: func(db *gorm.DB) error {
				return db.AutoMigrate(
					&model.DeckModel{},
					&model.DeckCardModel{},
				)
			},
		},
	}

	// 実行済みマイグレーションを取得
	var executedMigrations []GormMigration
	if err := db.Find(&executedMigrations).Error; err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	executedMap := make(map[string]bool)
	for _, m := range executedMigrations {
		executedMap[m.Version] = true
	}

	// 未実行のマイグレーションを実行
	for _, migration := range models {
		if executedMap[migration.version] {
			logger.Info("Migration %s (%s) already executed, skipping", migration.version, migration.description)
			continue
		}

		logger.Info("Running migration %s: %s...", migration.version, migration.description)

		// トランザクション内でマイグレーションを実行
		err := db.Transaction(func(tx *gorm.DB) error {
			// マイグレーション実行
			if err := migration.migrateFunc(tx); err != nil {
				return fmt.Errorf("failed to execute migration: %w", err)
			}

			// 実行記録を保存
			record := GormMigration{
				Version:     migration.version,
				Description: migration.description,
			}
			if err := tx.Create(&record).Error; err != nil {
				return fmt.Errorf("failed to record migration: %w", err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.version, err)
		}

		logger.Info("✅ Migration %s completed", migration.version)
	}

	return nil
}
