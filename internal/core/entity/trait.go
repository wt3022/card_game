package entity

// ========================================
// 特性（Trait）
// ユニットが持つ特殊能力を定義
// 設計方針:
// - 各特性は明確な効果を持つ
// - 戦闘、防御の2カテゴリに分類
// - メタデータで効果説明を提供
// ========================================

// Trait カードの特性（特殊能力）
type Trait string

// TraitCategory 特性のカテゴリ
type TraitCategory string

const (
	TraitCategoryCombat  TraitCategory = "COMBAT"  // 戦闘系
	TraitCategoryDefense TraitCategory = "DEFENSE" // 防御系
)

// ========================================
// 戦闘系特性
// ========================================

const (
	TraitRush     Trait = "RUSH"     // 疾走: 召喚酔いなし
	TraitDirect   Trait = "DIRECT"   // 直接攻撃: 相手プレイヤーに直接攻撃できる
	TraitWindfury Trait = "WINDFURY" // 疾風: 1ターンに2回攻撃可能
	TraitPierce   Trait = "PIERCE"   // 貫通: 余剰ダメージがプレイヤーに貫通
)

// ========================================
// 防御系特性
// ========================================

const (
	TraitGuardian     Trait = "GUARDIAN"      // 守護: 相手はこのユニットを攻撃しなければならない
	TraitEffectShield Trait = "EFFECT_SHIELD" // 効果盾: 効果によるダメージを受けない
	TraitUntargetable Trait = "UNTARGETABLE"  // 対象不可: 効果の対象にならない
)

// ========================================
// 特性メタデータ
// ========================================

// TraitInfo 特性の詳細情報
type TraitInfo struct {
	Trait       Trait         `json:"trait"`
	Name        string        `json:"name"`         // 表示名
	Description string        `json:"description"`  // 効果説明
	Category    TraitCategory `json:"category"`     // カテゴリ
	IsStackable bool          `json:"is_stackable"` // 重複可能か
}

// 特性情報マップ
var traitInfoMap = map[Trait]TraitInfo{
	// 戦闘系
	TraitRush: {
		Trait:       TraitRush,
		Name:        "疾走",
		Description: "召喚したターンに攻撃できる",
		Category:    TraitCategoryCombat,
		IsStackable: false,
	},
	TraitDirect: {
		Trait:       TraitDirect,
		Name:        "直接攻撃",
		Description: "相手プレイヤーに直接攻撃できる",
		Category:    TraitCategoryCombat,
		IsStackable: false,
	},
	TraitWindfury: {
		Trait:       TraitWindfury,
		Name:        "疾風",
		Description: "1ターンに2回攻撃できる",
		Category:    TraitCategoryCombat,
		IsStackable: false,
	},
	TraitPierce: {
		Trait:       TraitPierce,
		Name:        "貫通",
		Description: "余剰ダメージがプレイヤーに貫通",
		Category:    TraitCategoryCombat,
		IsStackable: false,
	},
	// 防御系
	TraitGuardian: {
		Trait:       TraitGuardian,
		Name:        "守護",
		Description: "相手はこのユニットを優先的に攻撃しなければならない",
		Category:    TraitCategoryDefense,
		IsStackable: false,
	},
	TraitEffectShield: {
		Trait:       TraitEffectShield,
		Name:        "効果盾",
		Description: "スペル・ユニット効果によるダメージを受けない（戦闘ダメージは受ける）",
		Category:    TraitCategoryDefense,
		IsStackable: false,
	},
	TraitUntargetable: {
		Trait:       TraitUntargetable,
		Name:        "対象不可",
		Description: "スペル・ユニット効果の対象にならない",
		Category:    TraitCategoryDefense,
		IsStackable: false,
	},
}

// ========================================
// ヘルパー関数
// ========================================

// GetTraitInfo 特性情報を取得
func GetTraitInfo(trait Trait) (TraitInfo, bool) {
	info, ok := traitInfoMap[trait]
	return info, ok
}

// GetTraitName 特性の表示名を取得
func GetTraitName(trait Trait) string {
	if info, ok := traitInfoMap[trait]; ok {
		return info.Name
	}
	return string(trait)
}

// GetTraitDescription 特性の説明を取得
func GetTraitDescription(trait Trait) string {
	if info, ok := traitInfoMap[trait]; ok {
		return info.Description
	}
	return ""
}

// GetTraitCategory 特性のカテゴリを取得
func GetTraitCategory(trait Trait) TraitCategory {
	if info, ok := traitInfoMap[trait]; ok {
		return info.Category
	}
	return TraitCategoryDefense
}

// IsTraitStackable 特性が重複可能か判定
func IsTraitStackable(trait Trait) bool {
	if info, ok := traitInfoMap[trait]; ok {
		return info.IsStackable
	}
	return false
}

// IsValidTrait 有効な特性か検証
func IsValidTrait(trait Trait) bool {
	_, ok := traitInfoMap[trait]
	return ok
}

// GetAllTraits すべての特性を取得
func GetAllTraits() []Trait {
	traits := make([]Trait, 0, len(traitInfoMap))
	for t := range traitInfoMap {
		traits = append(traits, t)
	}
	return traits
}

// GetTraitsByCategory カテゴリ別に特性を取得
func GetTraitsByCategory(category TraitCategory) []Trait {
	traits := make([]Trait, 0)
	for t, info := range traitInfoMap {
		if info.Category == category {
			traits = append(traits, t)
		}
	}
	return traits
}

// GetCombatTraits 戦闘系特性をすべて取得
func GetCombatTraits() []Trait {
	return GetTraitsByCategory(TraitCategoryCombat)
}

// GetDefenseTraits 防御系特性をすべて取得
func GetDefenseTraits() []Trait {
	return GetTraitsByCategory(TraitCategoryDefense)
}

// IsCombatTrait 戦闘系特性か判定
func IsCombatTrait(trait Trait) bool {
	return GetTraitCategory(trait) == TraitCategoryCombat
}

// IsDefenseTrait 防御系特性か判定
func IsDefenseTrait(trait Trait) bool {
	return GetTraitCategory(trait) == TraitCategoryDefense
}

// ========================================
// 特性判定ヘルパー
// ========================================

// HasTrait 特性リストに指定の特性が含まれているか判定
func HasTrait(traits []Trait, target Trait) bool {
	for _, t := range traits {
		if t == target {
			return true
		}
	}
	return false
}

// HasAnyTrait 特性リストに指定の特性のいずれかが含まれているか判定
func HasAnyTrait(traits []Trait, targets []Trait) bool {
	for _, target := range targets {
		if HasTrait(traits, target) {
			return true
		}
	}
	return false
}

// HasAllTraits 特性リストに指定の特性がすべて含まれているか判定
func HasAllTraits(traits []Trait, targets []Trait) bool {
	for _, target := range targets {
		if !HasTrait(traits, target) {
			return false
		}
	}
	return true
}

// AddTrait 特性を追加（重複チェック付き）
func AddTrait(traits []Trait, trait Trait) []Trait {
	// 既に持っている場合は追加しない
	if HasTrait(traits, trait) {
		return traits
	}
	return append(traits, trait)
}

// RemoveTrait 特性を削除
func RemoveTrait(traits []Trait, trait Trait) []Trait {
	result := make([]Trait, 0, len(traits))
	for _, t := range traits {
		if t != trait {
			result = append(result, t)
		}
	}
	return result
}

// FilterValidTraits 有効な特性のみをフィルタリング
func FilterValidTraits(traits []Trait) []Trait {
	result := make([]Trait, 0, len(traits))
	for _, t := range traits {
		if IsValidTrait(t) {
			result = append(result, t)
		}
	}
	return result
}
