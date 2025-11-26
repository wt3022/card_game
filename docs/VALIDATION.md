# バリデーション強化ドキュメント

## 実施日: 2025年11月26日

## 概要

プロジェクト全体のバリデーションを大幅に強化しました。データの整合性とセキュリティを向上させるため、エンティティ層とアプリケーション層の両方でバリデーションを実装しました。

## 主な改善点

### 1. デッキサイズの柔軟化

#### 変更前
```go
// デッキは厳密に40枚
if len(d.CardIDs) != DeckSize {
    return NewErrInvalidDeck("cards", "デッキは正確に40枚である必要があります")
}
```

#### 変更後
```go
const (
    DeckSize = 40        // 推奨枚数
    MinDeckSize = 30     // 最小枚数
    MaxDeckSize = 60     // 最大枚数
)

// デッキの枚数は30〜60枚（推奨は40枚）
if len(d.CardIDs) < MinDeckSize {
    return NewErrInvalidDeck("cards", fmt.Sprintf("デッキは最低%d枚必要です", MinDeckSize))
}
if len(d.CardIDs) > MaxDeckSize {
    return NewErrInvalidDeck("cards", fmt.Sprintf("デッキは最大%d枚までです", MaxDeckSize))
}
```

**メリット:**
- ユーザーの柔軟性が向上
- 異なるゲームモードに対応可能
- 推奨枚数は40枚のまま維持

### 2. カードバリデーションの強化

#### 新規追加された定数
```go
const (
    MaxCardNameLength   = 50
    MaxCardEffectLength = 500
    MaxCardCost         = 10
    MaxCardAttack       = 20
    MaxCardDefense      = 20
    MaxCardTraits       = 5
)
```

#### 強化されたバリデーション項目

**カード名:**
- 空文字チェック
- 最大長チェック（50文字）

**カードタイプ:**
- 空文字チェック
- 有効な値の検証（Unit/Spell/Leader）

**コスト:**
- 下限チェック（0以上）
- 上限チェック（10以下）

**攻撃力・防御力:**
- ユニットカードの必須チェック
- 下限チェック（0以上）
- 上限チェック（20以下）
- スペル・リーダーカードへの設定禁止

**特性:**
- 数の上限チェック（5個まで）
- スペル・リーダーカードへの設定禁止

**効果テキスト:**
- 最大長チェック（500文字）

### 3. デッキバリデーションの強化

#### 新規追加された定数
```go
const (
    MaxDeckNameLength        = 100
    MaxDeckDescriptionLength = 500
)
```

#### 強化されたバリデーション項目

**デッキ名:**
- 空文字チェック
- 最大長チェック（100文字）

**デッキ説明:**
- 最大長チェック（500文字）

**ユーザーID:**
- 空文字チェック

**カード:**
- 空のデッキチェック
- 枚数の範囲チェック（30〜60枚）
- 空のカードIDチェック
- 同じカードの枚数制限（3枚まで）

### 4. ユーザーバリデーションの新規実装

#### 新規追加された定数
```go
const (
    MinUsernameLength = 3
    MaxUsernameLength = 50
    MinPasswordLength = 8
    MaxPasswordLength = 128
)
```

#### バリデーション項目

**ユーザーID:**
- 空文字チェック

**ユーザー名:**
- 空文字チェック
- 最小長チェック（3文字）
- 最大長チェック（50文字）
- 使用可能文字チェック（英数字、アンダースコア、ハイフンのみ）

**パスワードハッシュ:**
- 最小長チェック（20文字、ハッシュ化済みの場合）

**ロール:**
- 有効な値の検証（admin/editor/viewer）

### 5. エラー処理の統一

#### 新規追加されたエラー型

```go
// ErrInvalidInput 無効な入力エラー（汎用バリデーションエラー）
type ErrInvalidInput struct {
    Field  string
    Reason string
}

func NewErrInvalidInput(field, reason string) DomainError
```

#### 使用例

```go
// カードバリデーション
if c.Name == "" {
    return NewErrInvalidInput("card.name", "カード名は必須です")
}

// ユーザーバリデーション
if err := ValidateUsername(username); err != nil {
    return err
}

// デッキバリデーション
if len(d.CardIDs) == 0 {
    return NewErrInvalidDeck("cards", "デッキにカードが含まれていません")
}
```

## バリデーション階層

### レイヤー1: エンティティ層（ドメインロジック）

**責務:**
- ドメインルールの検証
- データの整合性チェック
- ビジネスロジックの制約

**実装箇所:**
- `internal/core/entity/card.go` - Card.Validate()
- `internal/core/entity/deck.go` - Deck.Validate()
- `internal/core/entity/user.go` - User.Validate()

### レイヤー2: アプリケーション層（ビジネスロジック）

**責務:**
- 入力値の事前チェック
- 依存関係の検証
- リソースの存在確認

**実装箇所:**
- `internal/application/service/card_service.go`
- `internal/application/service/deck_service.go`
- `internal/application/service/auth_service.go`

### レイヤー3: アダプター層（外部インターフェース）

**責務:**
- プロトコル固有のバリデーション
- DTOからエンティティへの変換前チェック

## バリデーション実装パターン

### パターン1: エンティティでの自己検証

```go
// エンティティ作成時に自動的にバリデーション
func NewDeck(id, name, description, userID string, cardIDs []string) (*Deck, error) {
    deck := &Deck{...}
    if err := deck.Validate(); err != nil {
        return nil, err
    }
    return deck, nil
}
```

### パターン2: サービス層での事前検証

```go
func (s *CardService) CreateCard(card *entity.Card) error {
    // バリデーション
    if err := card.Validate(); err != nil {
        s.logger.Error("Card validation failed: %v", err)
        return err
    }
    
    // 既存のカードをチェック
    _, err := s.cardRepo.FindByID(card.ID)
    if err == nil {
        return entity.NewErrAlreadyExists("card", card.ID)
    }
    
    // 作成処理
    ...
}
```

### パターン3: ヘルパー関数での独立検証

```go
// ValidateUsername ユーザー名のバリデーション（他のレイヤーから呼び出し可能）
func ValidateUsername(username string) error {
    if username == "" {
        return NewErrInvalidInput("username", "ユーザー名は必須です")
    }
    // 追加の検証...
    return nil
}

// 使用例
func (s *AuthService) Login(username, password string) (*entity.LoginResponse, error) {
    if err := entity.ValidateUsername(username); err != nil {
        return nil, fmt.Errorf("認証情報が無効です")
    }
    // ログイン処理...
}
```

## セキュリティ上の改善

### 1. 入力サイズ制限

すべての文字列フィールドに最大長を設定し、DoS攻撃やメモリ枯渇を防止。

### 2. 使用可能文字の制限

ユーザー名に英数字とアンダースコア、ハイフンのみを許可し、SQLインジェクションやXSSのリスクを軽減。

### 3. 範囲チェック

数値フィールドに上限・下限を設定し、オーバーフローやゲームバランスの崩壊を防止。

### 4. 空文字チェック

重要なフィールドで空文字を許可しないことで、データの整合性を保証。

## テスト推奨事項

### ユニットテスト例

```go
func TestCard_Validate(t *testing.T) {
    tests := []struct {
        name    string
        card    *entity.Card
        wantErr bool
    }{
        {
            name: "正常なユニットカード",
            card: &entity.Card{
                ID:      "card001",
                Name:    "テストユニット",
                Type:    entity.CardTypeUnit,
                Cost:    3,
                Attack:  intPtr(5),
                Defense: intPtr(4),
            },
            wantErr: false,
        },
        {
            name: "名前が長すぎる",
            card: &entity.Card{
                ID:   "card002",
                Name: strings.Repeat("a", 51),
                Type: entity.CardTypeUnit,
            },
            wantErr: true,
        },
        {
            name: "攻撃力が上限を超える",
            card: &entity.Card{
                ID:      "card003",
                Name:    "強すぎるユニット",
                Type:    entity.CardTypeUnit,
                Cost:    10,
                Attack:  intPtr(21),
                Defense: intPtr(10),
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.card.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Card.Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## マイグレーション時の注意事項

### 既存データの対応

1. **デッキ枚数の移行**
   - 既存の40枚デッキはそのまま有効
   - 30〜60枚の範囲外のデッキは修正が必要

2. **カード名の長さ**
   - 50文字を超える既存カード名の確認と修正

3. **ユーザー名の文字**
   - 英数字以外を含む既存ユーザー名の確認と移行

### データベース制約の追加推奨

```sql
-- カード名の長さ制限
ALTER TABLE cards MODIFY COLUMN name VARCHAR(50) NOT NULL;

-- デッキ名の長さ制限
ALTER TABLE decks MODIFY COLUMN name VARCHAR(100) NOT NULL;
ALTER TABLE decks MODIFY COLUMN description VARCHAR(500);

-- ユーザー名の長さ制限
ALTER TABLE users MODIFY COLUMN username VARCHAR(50) NOT NULL;
```

## まとめ

このバリデーション強化により、以下を実現しました：

1. **柔軟性の向上**
   - デッキサイズの範囲を30〜60枚に拡大
   - ゲームの多様性をサポート

2. **セキュリティの強化**
   - すべての入力に対する厳密な検証
   - DoS攻撃やインジェクション攻撃への耐性向上

3. **データ品質の向上**
   - 一貫したエラーメッセージ
   - 明確なバリデーションルール

4. **保守性の向上**
   - 定数による設定の一元管理
   - 再利用可能なヘルパー関数

5. **ユーザー体験の改善**
   - より明確なエラーメッセージ
   - 適切な制約による直感的な使用感

## 今後の課題

- [ ] バリデーションのユニットテスト追加
- [ ] 既存データのマイグレーションスクリプト作成
- [ ] APIレスポンスでのバリデーションエラーの詳細化
- [ ] カスタムバリデーションルールの追加機能
- [ ] 多言語対応のエラーメッセージ
