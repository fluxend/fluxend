package webhook

import (
	"fluxend/internal/domain/auth"
	"fluxend/internal/domain/organization"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type Policy struct {
	organizationRepo organization.Repository
}

func NewWebhookPolicy(injector *do.Injector) (*Policy, error) {
	repo := do.MustInvoke[organization.Repository](injector)

	return &Policy{organizationRepo: repo}, nil
}

func (p *Policy) CanAccess(organizationUUID uuid.UUID, authUser auth.User) bool {
	isMember, err := p.organizationRepo.IsOrganizationMember(organizationUUID, authUser.Uuid)
	if err != nil {
		return false
	}

	return isMember
}

func (p *Policy) CanCreate(organizationUUID uuid.UUID, authUser auth.User) bool {
	if !authUser.IsDeveloperOrMore() {
		return false
	}

	isMember, err := p.organizationRepo.IsOrganizationMember(organizationUUID, authUser.Uuid)
	if err != nil {
		return false
	}

	return isMember
}

func (p *Policy) CanDelete(organizationUUID uuid.UUID, authUser auth.User) bool {
	if !authUser.IsDeveloperOrMore() {
		return false
	}

	isMember, err := p.organizationRepo.IsOrganizationMember(organizationUUID, authUser.Uuid)
	if err != nil {
		return false
	}

	return isMember
}
