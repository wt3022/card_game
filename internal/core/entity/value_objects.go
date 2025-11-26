package entity

import (
	"fmt"
	"strings"
)

// GameID はゲームIDを表すValue Object
type GameID string

// NewGameID は新しいGameIDを生成する
func NewGameID(id string) (GameID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("ゲームIDは空にできません")
	}
	return GameID(id), nil
}

// String はGameIDの文字列表現を返す
func (g GameID) String() string {
	return string(g)
}

// IsValid はGameIDが有効かどうかを判定する
func (g GameID) IsValid() bool {
	return g != ""
}

// PlayerID はプレイヤーIDを表すValue Object
type PlayerID string

// NewPlayerID は新しいPlayerIDを生成する
func NewPlayerID(id string) (PlayerID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("プレイヤーIDは空にできません")
	}
	return PlayerID(id), nil
}

// String はPlayerIDの文字列表現を返す
func (p PlayerID) String() string {
	return string(p)
}

// IsValid はPlayerIDが有効かどうかを判定する
func (p PlayerID) IsValid() bool {
	return p != ""
}

// CardID はカードIDを表すValue Object
type CardID string

// NewCardID は新しいCardIDを生成する
func NewCardID(id string) (CardID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("カードIDは空にできません")
	}
	return CardID(id), nil
}

// String はCardIDの文字列表現を返す
func (c CardID) String() string {
	return string(c)
}

// IsValid はCardIDが有効かどうかを判定する
func (c CardID) IsValid() bool {
	return c != ""
}

// UnitInstanceID はユニットのインスタンスIDを表すValue Object
type UnitInstanceID string

// NewUnitInstanceID は新しいUnitInstanceIDを生成する
func NewUnitInstanceID(id string) (UnitInstanceID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("ユニットインスタンスIDは空にできません")
	}
	return UnitInstanceID(id), nil
}

// String はUnitInstanceIDの文字列表現を返す
func (u UnitInstanceID) String() string {
	return string(u)
}

// IsValid はUnitInstanceIDが有効かどうかを判定する
func (u UnitInstanceID) IsValid() bool {
	return u != ""
}
