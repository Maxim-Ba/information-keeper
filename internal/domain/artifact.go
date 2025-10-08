package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	pb "github.com/Maxim-Ba/information-keeper/pkg/proto"
)

type Artifact struct {
	ID        string       `json:"id" db:"id"`
	OwnerID   string       `json:"owner_id" db:"owner_id"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	ExpiredAt time.Time    `json:"expired_at" db:"expired_at"`
	Type      ArtifactType `json:"type" db:"type_id"`
	MetaInfo  string       `json:"meta_info" db:"meta_info"`
	// Файл с полезной нагрузкой (используется для создания и редактирования), если при редактировании передается не пустое значение  значит происходит замена в s3 файла
	Payload []byte `json:"payload" db:"-"`
	// Ссылка на скачивание созданного артефакта
	Link string `json:"link" db:"link"`
}

type ArtifactType struct {
	Id   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

// ArtifactData - интерфейс для данных всех типов артефактов
type ArtifactData interface {
	Validate() error
	GetType() pb.ArtifactTypeEnum
	// Преобразование в бинарные данные для Payload
	ToBinary() ([]byte, error)
	// Парсинг из бинарных данных Payload
	FromBinary(data []byte) error
	// Получение мета-информации для отображения
	GetMetaInfo() string
}

// Конкретные типы данных
type LoginPasswordData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Site     string `json:"site,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

type TextData struct {
	Content string `json:"content"`
	Title   string `json:"title,omitempty"`
}

type BankCardData struct {
	Number     uint64 `json:"number"`
	Holder     string `json:"holder"`
	ExpiryDate string `json:"expiry_date"` // Формат "MM/YY"
	CVV        uint   `json:"cvv,omitempty"`
}

type BinaryData struct {
	Data        []byte `json:"-"`
	Description string `json:"description,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

// Реализация интерфейса ArtifactData для каждого типа
var _ ArtifactData = (*LoginPasswordData)(nil)

func (d *LoginPasswordData) Validate() error {
	if d.Login == "" || d.Password == "" {
		return fmt.Errorf("login and password are required")
	}
	return nil
}

func (d *LoginPasswordData) ToBinary() ([]byte, error) {
	return json.Marshal(d)
}

func (d *LoginPasswordData) FromBinary(data []byte) error {
	return json.Unmarshal(data, d)
}

func (d *LoginPasswordData) GetType() pb.ArtifactTypeEnum {
	return pb.ArtifactTypeEnum_LOGIN_PASSWORD
}

func (d *LoginPasswordData) GetMetaInfo() string {
	if d.Site != "" {
		return fmt.Sprintf("Логин для %s", d.Site)
	}
	return "Логин и пароль"
}

var _ ArtifactData = (*TextData)(nil)

func (d *TextData) Validate() error {
	if d.Content == "" {
		return fmt.Errorf("content is required")
	}
	return nil
}

func (d *TextData) ToBinary() ([]byte, error) {
	return json.Marshal(d)
}

func (d *TextData) FromBinary(data []byte) error {
	return json.Unmarshal(data, d)
}

func (d *TextData) GetType() pb.ArtifactTypeEnum {
	return pb.ArtifactTypeEnum_TEXT
}

func (d *TextData) GetMetaInfo() string {
	if d.Title != "" {
		return d.Title
	}
	if len(d.Content) > 50 {
		return d.Content[:50] + "..."
	}
	return d.Content
}

var _ ArtifactData = (*BankCardData)(nil)

func (d *BankCardData) Validate() error {
	if d.Number == 0 || d.Holder == "" || d.ExpiryDate == "" {
		return fmt.Errorf("number, holder and expiry date are required")
	}
	return nil
}

func (d *BankCardData) ToBinary() ([]byte, error) {
	return json.Marshal(d)
}

func (d *BankCardData) FromBinary(data []byte) error {
	return json.Unmarshal(data, d)
}

func (d *BankCardData) GetType() pb.ArtifactTypeEnum {
	return pb.ArtifactTypeEnum_BANK_CARD
}

func (d *BankCardData) GetMetaInfo() string {

	return "Банковская карта"
}

var _ ArtifactData = (*BinaryData)(nil)

func (d *BinaryData) Validate() error {
	if len(d.Data) == 0 {
		return fmt.Errorf("data is required")
	}

	return nil
}

func (d *BinaryData) ToBinary() ([]byte, error) {
	// Для BinaryData просто возвращаем данные как есть
	return d.Data, nil
}

func (d *BinaryData) FromBinary(data []byte) error {
	d.Data = data
	return nil
}

func (d *BinaryData) GetType() pb.ArtifactTypeEnum {
	return pb.ArtifactTypeEnum_BINARY
}

func (d *BinaryData) GetMetaInfo() string {
	if d.Description != "" {
		return d.Description
	}
	return "Файл"
}

// Вспомогательные функции
func CreateArtifactData(artifactType pb.ArtifactTypeEnum) ArtifactData {
	switch artifactType {
	case pb.ArtifactTypeEnum_LOGIN_PASSWORD:
		return &LoginPasswordData{}
	case pb.ArtifactTypeEnum_TEXT:
		return &TextData{}
	case pb.ArtifactTypeEnum_BANK_CARD:
		return &BankCardData{}
	case pb.ArtifactTypeEnum_BINARY:
		return &BinaryData{}
	default:
		return nil
	}
}
func ParseArtifactData(artifactType pb.ArtifactTypeEnum, payload []byte) (ArtifactData, error) {
	data := CreateArtifactData(artifactType)
	if data == nil {
		return nil, fmt.Errorf("unsupported artifact type: %v", artifactType)
	}

	if err := data.FromBinary(payload); err != nil {
		return nil, fmt.Errorf("failed to parse payload for type %v: %w", artifactType, err)
	}

	return data, nil
}

// Преобразования
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
		OwnerID:   dto.OwerID,
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
	}

	// Конвертируем тип
	var artifactType pb.ArtifactTypeEnum
	switch artifact.Type.Name {
	case "loginpassword":
		artifactType = pb.ArtifactTypeEnum_LOGIN_PASSWORD
	case "text":
		artifactType = pb.ArtifactTypeEnum_TEXT
	case "bankcard":
		artifactType = pb.ArtifactTypeEnum_BANK_CARD
	case "binary":
		artifactType = pb.ArtifactTypeEnum_BINARY
	default:
		artifactType = pb.ArtifactTypeEnum_TEXT
	}

	return &pb.Artifact{
		Id:        artifact.ID,
		OwnerId:   artifact.OwnerID,
		CreatedAt: artifact.CreatedAt.Unix(),
		UpdatedAt: artifact.UpdatedAt.Unix(),
		ExpiredAt: artifact.ExpiredAt.Unix(),
		Type:      artifactType,
		Link:      artifact.Link,
		MetaInfo:  artifact.MetaInfo,
	}
}
func PBToDomainArtifact(pbArtifact *pb.Artifact) *Artifact {
	if pbArtifact == nil {
		return nil
	}

	typeName := ""
	switch pbArtifact.Type {
	case pb.ArtifactTypeEnum_LOGIN_PASSWORD:
		typeName = "loginpassword"
	case pb.ArtifactTypeEnum_TEXT:
		typeName = "text"
	case pb.ArtifactTypeEnum_BANK_CARD:
		typeName = "bankcard"
	case pb.ArtifactTypeEnum_BINARY:
		typeName = "binary"
	default:
		typeName = "text"
	}

	return &Artifact{
		ID:        pbArtifact.Id,
		OwnerID:   pbArtifact.OwnerId,
		CreatedAt: time.Unix(pbArtifact.CreatedAt, 0),
		UpdatedAt: time.Unix(pbArtifact.UpdatedAt, 0),
		ExpiredAt: time.Unix(pbArtifact.ExpiredAt, 0),
		Type: ArtifactType{
			Id:   int(pbArtifact.Type),
			Name: typeName,
		},
		MetaInfo: pbArtifact.MetaInfo,
		Link:     pbArtifact.Link,
	}
}
