package form

import (
	"fluxend/internal/domain/shared"
	"github.com/google/uuid"
	"time"
)

type FieldResponse struct {
	shared.BaseEntity
	Uuid             uuid.UUID `db:"uuid" json:"uuid"`
	FormResponseUuid uuid.UUID `db:"form_response_uuid" json:"formResponseUuid"`
	FormFieldUuid    uuid.UUID `db:"form_field_uuid" json:"formFieldUuid"`
	Value            string    `db:"value" json:"value"`
	CreatedAt        time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time `db:"updated_at" json:"updatedAt"`
}
