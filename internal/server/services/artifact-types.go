package services

import "github.com/Maxim-Ba/information-keeper/internal/domain"

type ArtifactCreated struct {
	Artifact *domain.Artifact `json:"artifact"`
}

type ArtifactUpdated struct {
	Artifact *domain.Artifact `json:"artifact"`
}

type ArtifactDeleted struct {
}
