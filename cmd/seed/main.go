package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"card_game/internal/fixture/deck"
	infraLogger "card_game/internal/infrastructure/logger"
	"card_game/internal/infrastructure/persistence"
	"card_game/internal/infrastructure/repository"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 環境変数の読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// ロガーの初期化
	logger := infraLogger.NewConsoleLogger()

	// データベース接続
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	logger.Info("Connected to database")

	// マイグレーションの実行
	if err := persistence.RunGormMigrations(db, logger); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	logger.Info("Migrations completed")

	// リポジトリの初期化
	cardRepo := repository.NewCardRepository(db)
	deckRepo := repository.NewDeckRepository(db)

	// カードマスターデータの投入
	if err := seedCards(cardRepo, logger); err != nil {
		log.Fatalf("Failed to seed cards: %v", err)
	}

	// デッキマスターデータの投入
	if err := seedDecks(deckRepo, logger); err != nil {
		log.Fatalf("Failed to seed decks: %v", err)
	}

	logger.Info("Seed data successfully inserted")
}

func seedCards(cardRepo port.CardRepository, logger port.Logger) error {
	logger.Info("Seeding card master data...")

	// fixtureからカードデータを取得
	cards := deck.GenerateSampleDeck()

	// 既存のカードをチェックして、存在しない場合のみ投入
	for i, card := range cards {
		existingCard, err := cardRepo.FindByID(card.ID)
		if err == nil && existingCard != nil {
			logger.Info(fmt.Sprintf("Card already exists, skipping: %s", card.ID))
			continue
		}

		// カード基本情報を作成
		if err := cardRepo.Create(&card); err != nil {
			return fmt.Errorf("failed to create card %s: %w", card.ID, err)
		}

		// CardEffectがある場合は保存
		if card.CardEffect != nil {
			if err := cardRepo.SaveCardEffect(card.ID, card.CardEffect); err != nil {
				return fmt.Errorf("failed to save card effect for %s: %w", card.ID, err)
			}
			logger.Info(fmt.Sprintf("Saved card effect for: %s", card.ID))
		}

		logger.Info(fmt.Sprintf("Created card [%d/%d]: %s (%s)", i+1, len(cards), card.Name, card.ID))
	}

	logger.Info(fmt.Sprintf("Successfully seeded %d cards", len(cards)))
	return nil
}

func seedDecks(deckRepo port.DeckRepository, logger port.Logger) error {
	logger.Info("Seeding deck master data...")

	// マスターデッキの作成（システム用）
	masterDeck, err := createMasterDeck()
	if err != nil {
		return fmt.Errorf("failed to create master deck: %w", err)
	}

	// 既存のデッキをチェック
	ctx := context.Background()

	existingDeck, err := deckRepo.FindByID(ctx, masterDeck.ID)
	if err == nil && existingDeck != nil {
		// 既存のデッキがある場合、カードIDを更新
		logger.Info("Master deck already exists, updating card IDs...")
		existingDeck.CardIDs = masterDeck.CardIDs
		if err := deckRepo.Update(ctx, existingDeck); err != nil {
			return fmt.Errorf("failed to update master deck: %w", err)
		}
		logger.Info("Master deck updated successfully")
		return nil
	}

	// 新規作成
	if err := deckRepo.Create(ctx, masterDeck); err != nil {
		return fmt.Errorf("failed to create master deck: %w", err)
	}

	logger.Info(fmt.Sprintf("Created master deck: %s", masterDeck.Name))
	return nil
}

func createMasterDeck() (*entity.Deck, error) {
	// fixtureからカードデータを取得してカードIDリストを作成
	cards := deck.GenerateSampleDeck()

	// デッキは40枚まで。41枚ある場合は最後の1枚を除外
	maxCards := 40
	if len(cards) > maxCards {
		cards = cards[:maxCards]
	}

	cardIDs := make([]string, len(cards))
	for i, card := range cards {
		cardIDs[i] = card.ID
	}

	// マスターデッキを作成
	return entity.NewDeck(
		"master-deck-001",
		"Master Deck - Balanced",
		"A well-balanced master deck with various card types and strategies",
		"system",
		cardIDs,
	)
}
