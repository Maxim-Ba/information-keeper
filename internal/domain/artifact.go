package domain

import "time"

type Artifact struct {
	OwerId    string
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiredAt time.Time
	Type ArtifactType    
	MetaInfo string  
}

type ArtifactType struct {
	Id int
	Name string
}
