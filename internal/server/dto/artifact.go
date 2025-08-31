package dto

import "time"

type ArtifactDTO struct {
	ID        string
	OwerID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiredAt time.Time
	Type      ArtifactTypeDTO
	MetaInfo  string
	Link      string
}

type ArtifactTypeDTO struct {
	Id   int
	Name string
}
