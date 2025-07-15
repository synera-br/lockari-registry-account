package corev1

import "errors"

const (
	ErrInvalidPlan           = "invalid plan: plan cannot be nil"
	ErrInvalidPlanID         = "invalid plan: plan must be a valid UUID"
	ErrPlanAlreadySet        = "invalid plan: plan is already set"
	ErrPlanRequired          = "invalid plan: plan is required"
	ErrInvalidPlanInfo       = "invalid plan info: plan info cannot be nil"
	ErrInvalidPlanRequiresIs = "invalid plan info: requires is free, pro, enterprise or on_premise"
)

type Plan string

const (
	FREE_PLAN       Plan = "free"
	PRO_PLAN        Plan = "pro"
	ENTERPRISE_PLAN Plan = "enterprise"
	ON_PREMISE_PLAN Plan = "on_premise"
)

func (p *Plan) IsValid() error {
	if p == nil {
		return errors.New(ErrInvalidPlan)
	}

	switch *p {
	case FREE_PLAN, PRO_PLAN, ENTERPRISE_PLAN, ON_PREMISE_PLAN:
		return nil
	default:
		return errors.New(ErrInvalidPlanInfo)
	}
}
func (p *Plan) GetPlan() Plan {
	if p == nil {
		return ""
	}
	return *p
}

func (p *Plan) String() string {
	if p == nil {
		return ""
	}
	return string(*p)
}

func (p *Plan) GetPlanFeatures(plan string) error {
	if p == nil {
		return errors.New(ErrInvalidPlan)
	}

	switch plan {
	case string(FREE_PLAN):
		*p = FREE_PLAN
	case string(PRO_PLAN):
		*p = PRO_PLAN
	case string(ENTERPRISE_PLAN):
		*p = ENTERPRISE_PLAN
	case string(ON_PREMISE_PLAN):
		*p = ON_PREMISE_PLAN
	default:
		return errors.New(ErrInvalidPlanInfo)
	}
	return nil
}

func (p *Plan) GetFeaturesFromPlan(plan string) PlanFeatures {
	if p == nil {
		return nil
	}

	switch plan {
	case string(FREE_PLAN):
		return PlanFree
	case string(PRO_PLAN):
		return PlanPro
	case string(ENTERPRISE_PLAN):
		return PlanEnterprise
	case string(ON_PREMISE_PLAN):
		return PlanOnPremise
	default:
		return nil
	}
}

func (p *Plan) GetFeatures() PlanFeatures {
	if p == nil {
		return nil
	}

	switch p.String() {
	case string(FREE_PLAN):
		return PlanFree
	case string(PRO_PLAN):
		return PlanPro
	case string(ENTERPRISE_PLAN):
		return PlanEnterprise
	case string(ON_PREMISE_PLAN):
		return PlanOnPremise
	default:
		return nil
	}
}

// ===== PLAN FEATURES =====

// PlanFeature representa recursos do plano
type PlanFeature string

const (
	PlanFeatureBasic               PlanFeature = "basic"
	PlanFeatureAdvancedPermissions PlanFeature = "advanced_permissions"
	PlanFeatureCrossTenantSharing  PlanFeature = "cross_tenant_sharing"
	PlanFeatureAuditLogs           PlanFeature = "audit_logs"
	PlanFeatureBackup              PlanFeature = "backup"
	PlanFeatureExternalSharing     PlanFeature = "external_sharing"
	PlanFeatureVaultLimit          PlanFeature = "vault_limit"
	PlanFeatureUserLimit           PlanFeature = "user_limit"
	PlanFeatureUnlimitedVaults     PlanFeature = "unlimited_vaults"
	PlanFeatureUnlimitedUsers      PlanFeature = "unlimited_users"
	PlanFeatureAPIAccess           PlanFeature = "api_access"
	PlanFeatureGroupManagement     PlanFeature = "group_management"
	PlanFeatureSSO                 PlanFeature = "sso"
	PlanFeatureAdvancedSecurity    PlanFeature = "advanced_security"
	PlanFeatureBasicSharing        PlanFeature = "basic_sharing"
	PlanFeatureAdvancedSharing     PlanFeature = "advanced_sharing"
)

// AllPlanFeatures retorna todos os recursos válidos
func AllPlanFeatures() []PlanFeature {
	return []PlanFeature{
		PlanFeatureBasic, PlanFeatureAdvancedPermissions, PlanFeatureCrossTenantSharing,
		PlanFeatureAuditLogs, PlanFeatureBackup, PlanFeatureExternalSharing,
		PlanFeatureVaultLimit, PlanFeatureUserLimit, PlanFeatureUnlimitedVaults,
		PlanFeatureUnlimitedUsers, PlanFeatureAPIAccess, PlanFeatureGroupManagement,
		PlanFeatureSSO, PlanFeatureAdvancedSecurity, PlanFeatureBasicSharing,
		PlanFeatureAdvancedSharing,
	}
}

// IsValid verifica se o recurso é válido
func (pf PlanFeature) IsValid() bool {
	for _, valid := range AllPlanFeatures() {
		if pf == valid {
			return true
		}
	}
	return false
}

// String implementa fmt.Stringer
func (pf PlanFeature) String() string {
	return string(pf)
}

type PlanFeatures []PlanFeature

var (
	PlanFree PlanFeatures = []PlanFeature{
		PlanFeatureBasic,
		PlanFeatureVaultLimit,
		PlanFeatureUserLimit,
		PlanFeatureBasicSharing,
	}
	PlanPro PlanFeatures = []PlanFeature{
		PlanFeatureBasic,
		PlanFeatureAdvancedPermissions, PlanFeatureCrossTenantSharing,
		PlanFeatureVaultLimit,
		PlanFeatureUserLimit,
		PlanFeatureUnlimitedVaults,
		PlanFeatureUnlimitedUsers,
	}

	PlanEnterprise PlanFeatures = []PlanFeature{
		PlanFeatureBasic,
		PlanFeatureAdvancedPermissions,
		PlanFeatureCrossTenantSharing,
		PlanFeatureAuditLogs,
		PlanFeatureBackup,
		PlanFeatureExternalSharing,
		PlanFeatureVaultLimit,
		PlanFeatureUserLimit,
		PlanFeatureUnlimitedVaults,
		PlanFeatureUnlimitedUsers,
		PlanFeatureAPIAccess,
		PlanFeatureGroupManagement,
		PlanFeatureSSO,
		PlanFeatureAdvancedSecurity,
		PlanFeatureBasicSharing,
		PlanFeatureAdvancedSharing,
	}

	PlanOnPremise PlanFeatures = PlanEnterprise
)

func (pf *Plan) GetAuditLogLimit() int {
	if pf == nil {
		return 0
	}

	switch *pf {
	case FREE_PLAN:
		return 10 // Free plan has a limit of 10 audit logs
	case PRO_PLAN:
		return 500 // Pro plan has a limit of 500 audit logs
	case ENTERPRISE_PLAN, ON_PREMISE_PLAN:
		return 100000 // Enterprise and On-Premise plans have a limit of 100,000 audit logs
	default:
		return 0 // Invalid plan or no limit
	}
}
