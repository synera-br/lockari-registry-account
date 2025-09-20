package entitytenant

import (
	"fmt"
	"registry-account/pkg/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TenantResponseTestSuite struct {
	suite.Suite
	tenant *Tenant
}

func (s *TenantResponseTestSuite) TestNewTenantResponse_FailedNil() {
	var tenant *Tenant

	tenantResponse, err := NewTenantResponse(tenant)
	assert.Nil(s.T(), tenantResponse)
	assert.EqualError(s.T(), err, "tenant is nil")
}

func (s *TenantResponseTestSuite) TestNewTenantResponse() {
	name := "Test Tenant"
	slug := "test-tenant"
	owner := "owner@example.com"
	displayName := "Test Tenant Display"
	description := "This is a test tenant"

	tenant := &Tenant{
		Name:        name,
		Slug:        slug,
		Owner:       owner,
		DisplayName: displayName,
		Description: description,
	}

	s.tenant = tenant

	tenantResponse, err := NewTenantResponse(tenant)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), tenantResponse)
	assert.Equal(s.T(), name, tenantResponse.Name)
	assert.Equal(s.T(), slug, tenantResponse.Slug)
	assert.Equal(s.T(), owner, tenantResponse.Owner)
	assert.Equal(s.T(), displayName, tenantResponse.DisplayName)
	assert.Equal(s.T(), description, tenantResponse.Description)
}

func (s *TenantResponseTestSuite) TestTenantResponseValidate_FailedNil() {

	var response *TenantResponse
	err := response.Validate()
	assert.EqualError(s.T(), err, "invalid tenant response")
}

func (s *TenantResponseTestSuite) TestTenantResponseValidate_FailedID() {

	// how debug (print) tenantResponse
	fmt.Printf("tenant: %+v\n", *s.tenant)
	tenantResponse, err := NewTenantResponse(s.tenant)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), tenantResponse)

	tenantResponse.ID = ""
	err = tenantResponse.Validate()
	assert.EqualError(s.T(), err, utils.ErrInvalidTenantID.Error())
}

func (s *TenantResponseTestSuite) TestTenantResponseValidate_FailedTenant() {

	tenantResponse, err := NewTenantResponse(s.tenant)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), tenantResponse)

	tenantResponse.Name = ""
	err = tenantResponse.Validate()
	assert.EqualError(s.T(), err, utils.ErrInvalidTenantName.Error())
}

func (s *TenantResponseTestSuite) TestTenantResponseIsID_FailedNil() {

	var tenantResponse *TenantResponse

	b := tenantResponse.IsID()
	assert.False(s.T(), b)
}

func (s *TenantResponseTestSuite) TestTenantResponseIsID_FailedEmpty() {

	tenantResponse := &TenantResponse{
		Tenant: s.tenant,
		ID:     "",
	}

	b := tenantResponse.IsID()
	assert.False(s.T(), b)
}

func (s *TenantResponseTestSuite) TestTenantResponseIsID_Success() {

	tenantResponse := &TenantResponse{
		Tenant: s.tenant,
		ID:     "123",
	}

	b := tenantResponse.IsID()
	assert.True(s.T(), b)
}

func (s *TenantResponseTestSuite) TestTenantResponseIsGetID_FailedNil() {

	var tenantResponse *TenantResponse

	id := tenantResponse.GetID()
	assert.Empty(s.T(), id)
}

func (s *TenantResponseTestSuite) TestTenantResponseGetID_FailedEmpty() {

	tenantResponse := &TenantResponse{
		Tenant: s.tenant,
		ID:     "",
	}

	id := tenantResponse.GetID()
	assert.Empty(s.T(), id)
}

func (s *TenantResponseTestSuite) TestTenantResponseGetID_Success() {

	tenantResponse := &TenantResponse{
		Tenant: s.tenant,
		ID:     "123",
	}

	id := tenantResponse.GetID()
	assert.Equal(s.T(), "123", id)
}

func (s *TenantResponseTestSuite) TestTenantResponseGetTenant_FailedNil() {

	var tenantResponse *TenantResponse

	tenant := tenantResponse.GetTenant()
	assert.Empty(s.T(), tenant)
}

func (s *TenantResponseTestSuite) TestTenantResponseGetTenant_Success() {

	tenantResponse := &TenantResponse{
		Tenant: s.tenant,
		ID:     "123",
	}

	tenant := tenantResponse.GetTenant()
	assert.Equal(s.T(), s.tenant, tenant)
}

func TestTenantResponseTestSuite(t *testing.T) {
	suite.Run(t, new(TenantResponseTestSuite))
}
