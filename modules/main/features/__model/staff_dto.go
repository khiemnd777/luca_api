package model

import "time"

type StaffDTO struct {
	ID         int       `json:"id,omitempty"`
	Email      string    `json:"email,omitempty"`
	Password   *string   `json:"password,omitempty"`
	Name       string    `json:"name,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	Active     bool      `json:"active,omitempty"`
	Avatar     string    `json:"avatar,omitempty"`
	QrCode     *string   `json:"qr_code,omitempty"`
	SectionIDs []int     `json:"section_ids,omitempty"`
	RoleIDs    []int     `json:"role_ids,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}
