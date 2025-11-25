package entity

import "time"

// ========================================
// ユーザー
// ========================================

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
