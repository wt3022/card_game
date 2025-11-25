package entity

import (
	"errors"
	"fmt"
)

// ========================================
// エラー階層設計
// ========================================

// DomainError ドメインエラーの基底インターフェース
// すべてのドメインエラーはこのインターフェースを実装する
type DomainError interface {
	error
	Code() ErrorCode         // エラーコード
	Category() ErrorCategory // エラーカテゴリ
	IsRetryable() bool       // リトライ可能か
}

// ErrorCode エラーコード（API/ログで使用）
type ErrorCode string

const (
	// リソース関連
	ErrorCodeNotFound        ErrorCode = "NOT_FOUND"        // リソースが見つからない
	ErrorCodeAlreadyExists   ErrorCode = "ALREADY_EXISTS"   // リソースが既に存在する
	ErrorCodeInvalidResource ErrorCode = "INVALID_RESOURCE" // リソースが無効な場合

	// アクション関連
	ErrorCodeInvalidAction ErrorCode = "INVALID_ACTION" // 無効なアクション
	ErrorCodeInvalidState  ErrorCode = "INVALID_STATE"  // 無効な状態
	ErrorCodeNotYourTurn   ErrorCode = "NOT_YOUR_TURN"  // あなたのターンではない

	// リソース不足
	ErrorCodeInsufficientMana ErrorCode = "INSUFFICIENT_MANA" // マナ不足
	ErrorCodeInsufficientHP   ErrorCode = "INSUFFICIENT_HP"   // HP不足

	// バリデーション
	ErrorCodeInvalidCardType ErrorCode = "INVALID_CARD_TYPE" // 無効なカード種別
	ErrorCodeInvalidTarget   ErrorCode = "INVALID_TARGET"    // 無効な対象
	ErrorCodeMissingTarget   ErrorCode = "MISSING_TARGET"    // 対象が不足している

	// 効果関連
	ErrorCodeEffectFailed         ErrorCode = "EFFECT_FAILED"          // 効果処理失敗
	ErrorCodeEffectNotImplemented ErrorCode = "EFFECT_NOT_IMPLEMENTED" // 効果未実装

	// 内部エラー
	ErrorCodeInternal ErrorCode = "INTERNAL_ERROR" // 内部エラー
)

// ErrorCategory エラーカテゴリ（エラーハンドリングで使用）
type ErrorCategory string

const (
	ErrorCategoryNotFound     ErrorCategory = "NOT_FOUND"     // 404相当
	ErrorCategoryInvalidInput ErrorCategory = "INVALID_INPUT" // 400相当
	ErrorCategoryValidation   ErrorCategory = "VALIDATION"    // 400相当（バリデーションエラー）
	ErrorCategoryPrecondition ErrorCategory = "PRECONDITION"  // 412相当（ターン外、マナ不足、ゲーム終了後など）
	ErrorCategoryConflict     ErrorCategory = "CONFLICT"      // 409相当（既に存在など）
	ErrorCategoryInternal     ErrorCategory = "INTERNAL"      // 500相当（内部エラー）
)

// ========================================
// 具体的なエラー型
// ========================================

// ErrNotFound リソースが見つからないエラー
type ErrNotFound struct {
	ResourceType string `json:"resource_type"` // リソースタイプ
	ResourceID   string `json:"resource_id"`   // リソースID
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found: %s", e.ResourceType, e.ResourceID)
}

func (e *ErrNotFound) Code() ErrorCode {
	return ErrorCodeNotFound
}

func (e *ErrNotFound) Category() ErrorCategory {
	return ErrorCategoryNotFound
}

func (e *ErrNotFound) IsRetryable() bool {
	return false
}

func NewErrNotFound(resourceType, resourceID string) DomainError {
	return &ErrNotFound{
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
}

// ErrInsufficientResource リソース不足エラー
type ErrInsufficientResource struct {
	ResourceType string
	Required     int
	Current      int
}

func (e *ErrInsufficientResource) Error() string {
	return fmt.Sprintf("insufficient %s: required %d, current %d", e.ResourceType, e.Required, e.Current)
}

func (e *ErrInsufficientResource) Code() ErrorCode {
	return ErrorCodeInsufficientMana // デフォルトはマナ不足
}

func (e *ErrInsufficientResource) Category() ErrorCategory {
	return ErrorCategoryPrecondition
}

func (e *ErrInsufficientResource) IsRetryable() bool {
	return false // リソース不足はリトライ不可
}

func NewErrInsufficientMana(required, current int) DomainError {
	return &ErrInsufficientResource{
		ResourceType: "mana",
		Required:     required,
		Current:      current,
	}
}

func NewErrInsufficientHP(required, current int) DomainError {
	err := &ErrInsufficientResource{
		ResourceType: "hp",
		Required:     required,
		Current:      current,
	}
	// カスタムエラーコードを設定するためのヘルパー
	return &errWithCode{err, ErrorCodeInsufficientHP}
}

// ErrInvalidAction 無効なアクションエラー
type ErrInvalidAction struct {
	Action string
	Reason string
}

func (e *ErrInvalidAction) Error() string {
	return fmt.Sprintf("invalid action %s: %s", e.Action, e.Reason)
}

func (e *ErrInvalidAction) Code() ErrorCode {
	return ErrorCodeInvalidAction
}

func (e *ErrInvalidAction) Category() ErrorCategory {
	return ErrorCategoryPrecondition
}

func (e *ErrInvalidAction) IsRetryable() bool {
	return false
}

func NewErrInvalidAction(action, reason string) DomainError {
	return &ErrInvalidAction{
		Action: action,
		Reason: reason,
	}
}

// ErrAlreadyExists リソースが既に存在するエラー
type ErrAlreadyExists struct {
	ResourceType string
	ResourceID   string
}

func (e *ErrAlreadyExists) Error() string {
	return fmt.Sprintf("%s already exists: %s", e.ResourceType, e.ResourceID)
}

func (e *ErrAlreadyExists) Code() ErrorCode {
	return ErrorCodeAlreadyExists
}

func (e *ErrAlreadyExists) Category() ErrorCategory {
	return ErrorCategoryConflict
}

func (e *ErrAlreadyExists) IsRetryable() bool {
	return false
}

func NewErrAlreadyExists(resourceType, resourceID string) DomainError {
	return &ErrAlreadyExists{
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
}

// ErrInvalidCardType 無効なカード種別エラー
type ErrInvalidCardType struct {
	Expected string
	Actual   string
}

func (e *ErrInvalidCardType) Error() string {
	return fmt.Sprintf("invalid card type: expected %s, got %s", e.Expected, e.Actual)
}

func (e *ErrInvalidCardType) Code() ErrorCode {
	return ErrorCodeInvalidCardType
}

func (e *ErrInvalidCardType) Category() ErrorCategory {
	return ErrorCategoryInvalidInput
}

func (e *ErrInvalidCardType) IsRetryable() bool {
	return false
}

func NewErrInvalidCardType(expected, actual string) DomainError {
	return &ErrInvalidCardType{
		Expected: expected,
		Actual:   actual,
	}
}

// ErrInvalidState 無効な状態エラー（ターン外など）
type ErrInvalidState struct {
	State  string
	Reason string
}

func (e *ErrInvalidState) Error() string {
	return fmt.Sprintf("invalid state %s: %s", e.State, e.Reason)
}

func (e *ErrInvalidState) Code() ErrorCode {
	return ErrorCodeInvalidState
}

func (e *ErrInvalidState) Category() ErrorCategory {
	return ErrorCategoryPrecondition
}

func (e *ErrInvalidState) IsRetryable() bool {
	return false
}

func NewErrInvalidState(state, reason string) DomainError {
	return &ErrInvalidState{
		State:  state,
		Reason: reason,
	}
}

func NewErrNotYourTurn(playerID string) DomainError {
	return &ErrInvalidState{
		State:  "turn",
		Reason: fmt.Sprintf("not your turn: %s", playerID),
	}
}

// ErrEffectFailed 効果処理失敗エラー
type ErrEffectFailed struct {
	EffectType string
	Reason     string
}

func (e *ErrEffectFailed) Error() string {
	return fmt.Sprintf("effect failed: %s - %s", e.EffectType, e.Reason)
}

func (e *ErrEffectFailed) Code() ErrorCode {
	return ErrorCodeEffectFailed
}

func (e *ErrEffectFailed) Category() ErrorCategory {
	return ErrorCategoryInternal
}

func (e *ErrEffectFailed) IsRetryable() bool {
	return true // 効果失敗はリトライ可能な場合がある
}

func NewErrEffectFailed(effectType, reason string) DomainError {
	return &ErrEffectFailed{
		EffectType: effectType,
		Reason:     reason,
	}
}

// ErrEffectNotImplemented 効果未実装エラー
type ErrEffectNotImplemented struct {
	EffectType string
}

func (e *ErrEffectNotImplemented) Error() string {
	return fmt.Sprintf("effect not implemented: %s", e.EffectType)
}

func (e *ErrEffectNotImplemented) Code() ErrorCode {
	return ErrorCodeEffectNotImplemented
}

func (e *ErrEffectNotImplemented) Category() ErrorCategory {
	return ErrorCategoryInternal
}

func (e *ErrEffectNotImplemented) IsRetryable() bool {
	return false
}

func NewErrEffectNotImplemented(effectType string) DomainError {
	return &ErrEffectNotImplemented{
		EffectType: effectType,
	}
}

// ========================================
// ヘルパー型（カスタムエラーコード用）
// ========================================

// errWithCode エラーコードを上書きするラッパー
type errWithCode struct {
	DomainError
	code ErrorCode
}

func (e *errWithCode) Code() ErrorCode {
	return e.code
}

// ========================================
// エラーユーティリティ
// ========================================

// IsDomainError ドメインエラーかどうかを判定
func IsDomainError(err error) bool {
	var domainErr DomainError
	return errors.As(err, &domainErr)
}

// AsDomainError ドメインエラーに変換
func AsDomainError(err error) (DomainError, bool) {
	var domainErr DomainError
	ok := errors.As(err, &domainErr)
	return domainErr, ok
}

// GetErrorCode エラーコードを取得（ドメインエラーでない場合は空文字列）
func GetErrorCode(err error) ErrorCode {
	if domainErr, ok := AsDomainError(err); ok {
		return domainErr.Code()
	}
	return ""
}

// GetErrorCategory エラーカテゴリを取得
func GetErrorCategory(err error) ErrorCategory {
	if domainErr, ok := AsDomainError(err); ok {
		return domainErr.Category()
	}
	return ErrorCategoryInternal
}

// ErrInvalidDeck 無効なデッキエラー
type ErrInvalidDeck struct {
	Field  string
	Reason string
}

func (e *ErrInvalidDeck) Error() string {
	return fmt.Sprintf("invalid deck %s: %s", e.Field, e.Reason)
}

func (e *ErrInvalidDeck) Code() ErrorCode {
	return ErrorCodeInvalidResource
}

func (e *ErrInvalidDeck) Category() ErrorCategory {
	return ErrorCategoryValidation
}

func (e *ErrInvalidDeck) IsRetryable() bool {
	return false
}

func NewErrInvalidDeck(field, reason string) DomainError {
	return &ErrInvalidDeck{
		Field:  field,
		Reason: reason,
	}
}
