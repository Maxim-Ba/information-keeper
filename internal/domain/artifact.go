package domain

import "time"

type Artifact struct {
	ID        string       `json:"id" db:"id"`
	OwerID    string       `json:"owner_id" db:"owner_id"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	ExpiredAt time.Time    `json:"expired_at" db:"expired_at"`
	Type      ArtifactType `json:"type" db:"type_id"`
	MetaInfo  string       `json:"meta_info" db:"meta_info"`
	Link      string       `json:"link" db:"link"` // URL to the ws3

}

type ArtifactType struct {
	Id   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}
