package domain

import (
	"time"

	pb "github.com/Maxim-Ba/information-keeper/pkg/proto"

	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
)

type Artifact struct {
	ID        string       `json:"id" db:"id"`
	OwnerID    string       `json:"owner_id" db:"owner_id"`
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
func DomainToArtifactDTO(artifact *Artifact) dto.ArtifactDTO {
    return dto.ArtifactDTO{
        ID:        artifact.ID,
        OwerID:    artifact.OwnerID,
        CreatedAt: artifact.CreatedAt,
        UpdatedAt: artifact.UpdatedAt,
        ExpiredAt: artifact.ExpiredAt,
        Type: dto.ArtifactTypeDTO{
            Id:   artifact.Type.Id,
            Name: artifact.Type.Name,
        },
        MetaInfo: artifact.MetaInfo,
        Link:     artifact.Link,
    }
}

func ArtifactDTOToDomain(dto dto.ArtifactDTO) Artifact {
    return Artifact{
        ID:        dto.ID,
        OwnerID:    dto.OwerID,
        CreatedAt: dto.CreatedAt,
        UpdatedAt: dto.UpdatedAt,
        ExpiredAt: dto.ExpiredAt,
        Type: ArtifactType{
            Id:   dto.Type.Id,
            Name: dto.Type.Name,
        },
        MetaInfo: dto.MetaInfo,
        Link:     dto.Link,
    }
}
func DomainToPBArtifact(artifact *Artifact) *pb.Artifact {
	if artifact == nil {
		return nil
	} else {
		return &pb.Artifact{
			Id:        artifact.ID,
		OwnerId: artifact.OwnerID,
		CreatedAt: artifact.CreatedAt.Unix(),
		UpdatedAt: artifact.UpdatedAt.Unix(),
		ExpiredAt: artifact.ExpiredAt.Unix(),
			
		Link: artifact.Link,
		MetaInfo: artifact.MetaInfo,
			// Type: ,
		
		}
	}}
