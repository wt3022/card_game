package entity

// ========================================
// 戦闘結果
// 戦闘結果を表す構造体を定義する
// ========================================

// 戦闘結果
type CombatResult struct {
	AttackerDestroyed bool // 攻撃側のユニットが破壊されたか
	DefenderDestroyed bool // 防御側のユニットが破壊されたか
	Damage            int  // 戦闘で発生したユニットへのダメージ
	DirectDamage      int  // プレイヤーへの直接ダメージ
}

func (c *CombatResult) Error() string {
	panic("unimplemented")
}
