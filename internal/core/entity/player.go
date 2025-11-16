package entity

import (
	"time"
)

// ========================================
// プレイヤー
// プレイヤーの状態とリソース管理を定義
// ========================================

// Player プレイヤーの状態
type Player struct {
	ID               string        `json:"id"`                  // プレイヤーID
	Name             string        `json:"name"`                // プレイヤー名
	HP               int           `json:"hp"`                  // 現在のHP
	MaxHP            int           `json:"max_hp"`              // 最大HP
	CurrentTurnMana  int           `json:"current_turn_mana"`   // 現在のターンのマナ
	MaxRecoveryMana  int           `json:"max_recovery_mana"`   // 最大回復マナ
	Hand             []Card        `json:"hand"`                // 手札
	Deck             []Card        `json:"deck"`                // デッキ
	Graveyard        []Card        `json:"graveyard"`           // 墓地
	Field            []Unit        `json:"field"`               // フィールド
	Leader           *Card         `json:"Leader"`              // リーダー
	TimeRemaining    time.Duration `json:"time_remaining"`      // ターン残り時間
	HasDrawnThisTurn bool          `json:"has_drawn_this_turn"` // このターンにドローしたか
	IsFirstTurn      bool          `json:"is_first_turn"`       // 初ターンかどうか
}

// ========================================
// ファクトリー関数
// ========================================

// NewPlayer 新しいプレイヤーを作成
func NewPlayer(id, name string, deck []Card) *Player {
	return &Player{
		ID:               id,
		Name:             name,
		HP:               InitialHP,
		MaxHP:            InitialHP,
		CurrentTurnMana:  InitialMana,
		MaxRecoveryMana:  InitialMana,
		Hand:             []Card{},
		Deck:             deck,
		Graveyard:        []Card{},
		Field:            []Unit{},
		Leader:           nil,
		TimeRemaining:    DefaultTurnTime,
		HasDrawnThisTurn: false,
		IsFirstTurn:      true,
	}
}

// DrawInitialHand 初期手札をドローする
func (p *Player) DrawInitialHand() error {
	for i := 0; i < InitialHandSize; i++ {
		if _, err := p.DrawCard(); err != nil {
			return err
		}
	}
	return nil
}

// ========================================
// カードドロー
// ========================================

// DrawCard デッキからカードを1枚引く
func (p *Player) DrawCard() (*Card, error) {
	if len(p.Deck) == 0 {
		return nil, NewErrInvalidState("deck", "deck exhausted")
	}

	card := p.Deck[0]
	p.Deck = p.Deck[1:]
	p.Hand = append(p.Hand, card)
	p.HasDrawnThisTurn = true

	return &card, nil
}

// DrawCards 複数枚のカードをドロー
func (p *Player) DrawCards(count int) ([]Card, error) {
	drawn := []Card{}
	for i := 0; i < count; i++ {
		card, err := p.DrawCard()
		if err != nil {
			return drawn, err
		}
		drawn = append(drawn, *card)
	}
	return drawn, nil
}

// ========================================
// HP管理
// ========================================

// TakeDamage ダメージを受ける
func (p *Player) TakeDamage(amount int) {
	p.HP -= amount
	if p.HP < 0 {
		p.HP = 0
	}
}

// HealHP HPを回復
func (p *Player) HealHP(amount int) {
	p.HP += amount
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
}

// IsDefeated 敗北したか確認
func (p *Player) IsDefeated() bool {
	return p.HP <= 0
}

// ========================================
// マナ管理
// ========================================

// SpendMana マナを消費
func (p *Player) SpendMana(amount int) error {
	if p.CurrentTurnMana < amount {
		return NewErrInsufficientMana(amount, p.CurrentTurnMana)
	}
	p.CurrentTurnMana -= amount
	return nil
}

// RecoverMana マナを回復
// ルール: CurrentTurnMana = BeforeTurnMana + MaxRecoveryMana
func (p *Player) RecoverMana() {
	p.CurrentTurnMana += p.MaxRecoveryMana
	if p.CurrentTurnMana > MaxMana {
		p.CurrentTurnMana = MaxMana
	}
}

// AddMana マナを加算（上限あり）
func (p *Player) AddMana(amount int) {
	p.CurrentTurnMana += amount
	if p.CurrentTurnMana > MaxMana {
		p.CurrentTurnMana = MaxMana
	}
}

// IncrementMaxRecoveryMana 最大回復マナを増加
func (p *Player) IncrementMaxRecoveryMana() {
	if p.MaxRecoveryMana < MaxRecoveryMana {
		p.MaxRecoveryMana++
	}
}

// GetCurrentTurnMana 現在ターンのマナを取得
func (p *Player) GetCurrentTurnMana() int {
	return p.CurrentTurnMana
}

// ========================================
// カードプレイ
// ========================================

// PlayCardFromHand 手札からカードをプレイ
func (p *Player) PlayCardFromHand(cardID string) (*Card, error) {
	for i, card := range p.Hand {
		if card.ID == cardID {
			// マナチェック
			if err := p.SpendMana(card.Cost); err != nil {
				return nil, err
			}

			// 手札から削除
			p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
			return &card, nil
		}
	}
	return nil, NewErrNotFound("card", cardID)
}

// SummonUnit ユニットを盤面に配置
func (p *Player) SummonUnit(card Card, instanceID string) error {
	if card.Attack == nil || card.Defense == nil {
		return NewErrInvalidCardType("unit", "spell")
	}

	unit := Unit{
		CardID:         card.ID,
		InstanceID:     instanceID,
		Name:           card.Name,
		Cost:           card.Cost,
		Attack:         *card.Attack,
		Defense:        *card.Defense,
		CurrentDefense: *card.Defense,
		Traits:         card.Traits,
		Effect:         card.Effect,
		OwnerID:        p.ID,
	}

	// 召喚時の初期化
	unit.InitializeOnSummon()

	p.Field = append(p.Field, unit)
	return nil
}

// UseSpell スペルカードを使用して墓地に送る
func (p *Player) UseSpell(card Card) {
	p.Graveyard = append(p.Graveyard, card)
}

// ========================================
// フィールド管理
// ========================================

// RemoveUnitFromField フィールドからユニットを削除
func (p *Player) RemoveUnitFromField(instanceID string) *Unit {
	for i, unit := range p.Field {
		if unit.InstanceID == instanceID {
			// スライスから削除
			p.Field = append(p.Field[:i], p.Field[i+1:]...)
			return &unit
		}
	}
	return nil
}

// GetUnitByInstanceID フィールドに存在するユニットをInstanceIDで取得
func (p *Player) GetUnitByInstanceID(instanceID string) *Unit {
	for i := range p.Field {
		if p.Field[i].InstanceID == instanceID {
			return &p.Field[i]
		}
	}
	return nil
}

// ResetUnitsForNewTurn 新しいターンのためにユニットをリセット
func (p *Player) ResetUnitsForNewTurn() {
	for i := range p.Field {
		p.Field[i].ResetForNewTurn()
	}
}

// GetFieldSize フィールドのユニット数を取得
func (p *Player) GetFieldSize() int {
	return len(p.Field)
}

// ========================================
// ターン管理
// ========================================

// AddTimeToTurn ターン開始時に時間を追加
func (p *Player) AddTimeToTurn(duration time.Duration) {
	p.TimeRemaining += duration
}

// ResetTurnFlags ターンフラグをリセット
func (p *Player) ResetTurnFlags() {
	p.HasDrawnThisTurn = false
	p.ResetUnitsForNewTurn()
	// p.IsFirstTurnはここでリセットしない（1度ターンを開始したらfalseにするのはドローフェイズ側で行う）
}

// ========================================
// 状態取得
// ========================================

// GetID プレイヤーIDを取得
func (p *Player) GetID() string {
	return p.ID
}

// HasCardsInDeck デッキにカードがあるか確認
func (p *Player) HasCardsInDeck() bool {
	return len(p.Deck) > 0
}

// IsDeckOut デッキ切れか確認
func (p *Player) IsDeckOut() bool {
	return !p.HasCardsInDeck()
}

// GetHandSize 手札の枚数を取得
func (p *Player) GetHandSize() int {
	return len(p.Hand)
}
