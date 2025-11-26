package entity

import (
	"fmt"
	"strings"
	"time"
)

// ========================================
// ユーザー
// ========================================

const (
	// ユーザーバリデーション定数
	MinUsernameLength = 3
	MaxUsernameLength = 50
	MinPasswordLength = 8
	MaxPasswordLength = 128
)

// UserRole ユーザーのロール
type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"  // 管理者
	UserRoleEditor UserRole = "editor" // 編集者
	UserRoleViewer UserRole = "viewer" // 閲覧者
)

// User ユーザーエンティティ
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // JSONには含めない
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// IsAdmin 管理者かどうか
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// IsEditor 編集者かどうか
func (u *User) IsEditor() bool {
	return u.Role == UserRoleEditor
}

// CanEdit 編集権限があるか（adminまたはeditor）
func (u *User) CanEdit() bool {
	return u.Role == UserRoleAdmin || u.Role == UserRoleEditor
}

// CanView 閲覧権限があるか（すべてのロール）
func (u *User) CanView() bool {
	return true
}

// Validate ユーザーのバリデーション
func (u *User) Validate() error {
	// ID検証
	if u.ID == "" {
		return NewErrInvalidInput("user.id", "ユーザーIDは必須です")
	}

	// ユーザー名検証
	if err := ValidateUsername(u.Username); err != nil {
		return err
	}

	// パスワードハッシュ検証（新規作成時は空でも可）
	if u.PasswordHash != "" && len(u.PasswordHash) < 20 {
		return NewErrInvalidInput("user.password_hash", "パスワードハッシュが無効です")
	}

	// ロール検証
	if err := ValidateUserRole(u.Role); err != nil {
		return err
	}

	return nil
}

// ValidateUsername ユーザー名のバリデーション
func ValidateUsername(username string) error {
	if username == "" {
		return NewErrInvalidInput("username", "ユーザー名は必須です")
	}

	trimmed := strings.TrimSpace(username)
	if len(trimmed) < MinUsernameLength {
		return NewErrInvalidInput("username", fmt.Sprintf("ユーザー名は%d文字以上である必要があります", MinUsernameLength))
	}
	if len(trimmed) > MaxUsernameLength {
		return NewErrInvalidInput("username", fmt.Sprintf("ユーザー名は%d文字以内である必要があります", MaxUsernameLength))
	}

	// 使用可能な文字のチェック（英数字、アンダースコア、ハイフン）
	for _, r := range trimmed {
		if !isValidUsernameChar(r) {
			return NewErrInvalidInput("username", "ユーザー名には英数字、アンダースコア、ハイフンのみ使用できます")
		}
	}

	return nil
}

// ValidateUserRole ロールのバリデーション
func ValidateUserRole(role UserRole) error {
	switch role {
	case UserRoleAdmin, UserRoleEditor, UserRoleViewer:
		return nil
	default:
		return NewErrInvalidInput("user.role", "無効なユーザーロールです")
	}
}

// isValidUsernameChar ユーザー名に使用可能な文字かチェック
func isValidUsernameChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_' || r == '-'
}
