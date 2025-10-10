package dto

import "time"

// ArtifactDTO - структура для передачи данных артефакта
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

// ArtifactTypeDTO - структура для передачи данных типа артефакта
type ArtifactTypeDTO struct {
	Id   int
	Name string
}
