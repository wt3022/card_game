package game

import (
	"fmt"
)

// ========================================
// 盤面表示
// ゲームの盤面を表示
// 設計方針:
// - プレイヤー情報の表示
// - フィールド情報の表示
// - ユニットの状態表示
// ========================================

// 盤面を表示
func (s *State) PrintBattlefield() {
	fmt.Println("\n" + "================================================================================")
	fmt.Println("【盤面の状態}")
	fmt.Println("================================================================================")

	// プレイヤー1の情報
	p1 := s.Player1
	fmt.Printf("\n■ %s (HP: %d/%d, マナ: %d/%d)\n",
		p1.Name, p1.HP, p1.MaxHP, p1.CurrentTurnMana, p1.MaxRecoveryMana)
	fmt.Printf("  手札: %d枚, デッキ: %d枚, 墓地: %d枚\n",
		len(p1.Hand), len(p1.Deck), len(p1.Graveyard))

	if len(p1.Field) == 0 {
		fmt.Println("  フィールド: (なし)")
	} else {
		fmt.Println("  フィールド:")
		for _, unit := range p1.Field {
			attackStatus := "○"
			if unit.CanAttack() {
				attackStatus = "●"
			}
			traits := ""
			if len(unit.Traits) > 0 {
				for _, kw := range unit.Traits {
					traits += fmt.Sprintf("[%s]", kw)
				}
			}
			fmt.Printf("    %s 「%s」 (攻:%d/守:%d/%d) %s\n",
				attackStatus, unit.Name, unit.Attack, unit.CurrentDefense, unit.Defense, traits)
		}
	}

	fmt.Println()
	fmt.Println("  ～ vs ～")
	fmt.Println()

	// プレイヤー2の情報
	p2 := s.Player2
	fmt.Printf("■ %s (HP: %d/%d, マナ: %d/%d)\n",
		p2.Name, p2.HP, p2.MaxHP, p2.CurrentTurnMana, p2.MaxRecoveryMana)
	fmt.Printf("  手札: %d枚, デッキ: %d枚, 墓地: %d枚\n",
		len(p2.Hand), len(p2.Deck), len(p2.Graveyard))

	if len(p2.Field) == 0 {
		fmt.Println("  フィールド: (なし)")
	} else {
		fmt.Println("  フィールド:")
		for _, unit := range p2.Field {
			attackStatus := "○"
			if unit.CanAttack() {
				attackStatus = "●"
			}
			traits := ""
			if len(unit.Traits) > 0 {
				for _, kw := range unit.Traits {
					traits += fmt.Sprintf("[%s]", kw)
				}
			}
			fmt.Printf("    %s 「%s」 (攻:%d/守:%d/%d) %s\n",
				attackStatus, unit.Name, unit.Attack, unit.CurrentDefense, unit.Defense, traits)
		}
	}

	fmt.Println("================================================================================")
}
