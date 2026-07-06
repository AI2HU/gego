package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/models"
)

func toBrandResponse(brand *models.Brand) models.BrandResponse {
	aliases := make([]models.BrandAliasResponse, len(brand.Aliases))
	for i, alias := range brand.Aliases {
		aliases[i] = models.BrandAliasResponse{
			ID:            alias.ID,
			BrandID:       alias.BrandID,
			Alias:         alias.Alias,
			CaseSensitive: alias.CaseSensitive,
			CreatedAt:     alias.CreatedAt,
			UpdatedAt:     alias.UpdatedAt,
		}
	}
	return models.BrandResponse{
		ID:        brand.ID,
		Name:      brand.Name,
		Aliases:   aliases,
		CreatedAt: brand.CreatedAt,
		UpdatedAt: brand.UpdatedAt,
	}
}

func (s *Server) listBrands(c *gin.Context) {
	brands, err := s.brandsService.ListBrands(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list brands: "+err.Error())
		return
	}

	responses := make([]models.BrandResponse, len(brands))
	for i, brand := range brands {
		responses[i] = toBrandResponse(brand)
	}
	s.successResponse(c, responses)
}

func (s *Server) createBrand(c *gin.Context) {
	var req models.CreateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	brand, err := s.brandsService.CreateBrand(c.Request.Context(), req.Name, req.Aliases)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to create brand: "+err.Error())
		return
	}
	s.successResponse(c, toBrandResponse(brand))
}

func (s *Server) updateBrand(c *gin.Context) {
	var req models.UpdateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	brand, err := s.brandsService.UpdateBrand(c.Request.Context(), c.Param("id"), req.Name)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to update brand: "+err.Error())
		return
	}
	s.successResponse(c, toBrandResponse(brand))
}

func (s *Server) deleteBrand(c *gin.Context) {
	if err := s.brandsService.DeleteBrand(c.Request.Context(), c.Param("id")); err != nil {
		s.errorResponse(c, http.StatusNotFound, "Failed to delete brand: "+err.Error())
		return
	}
	s.successResponse(c, gin.H{"id": c.Param("id")})
}

func (s *Server) createBrandAlias(c *gin.Context) {
	var req models.CreateBrandAliasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	alias, err := s.brandsService.AddAlias(c.Request.Context(), c.Param("id"), req.Alias, req.CaseSensitive)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to create brand alias: "+err.Error())
		return
	}

	s.successResponse(c, models.BrandAliasResponse{
		ID:            alias.ID,
		BrandID:       alias.BrandID,
		Alias:         alias.Alias,
		CaseSensitive: alias.CaseSensitive,
		CreatedAt:     alias.CreatedAt,
		UpdatedAt:     alias.UpdatedAt,
	})
}

func (s *Server) updateBrandAlias(c *gin.Context) {
	var req models.UpdateBrandAliasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	alias, err := s.brandsService.UpdateAlias(
		c.Request.Context(),
		c.Param("id"),
		c.Param("aliasId"),
		req.Alias,
		req.CaseSensitive,
	)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to update brand alias: "+err.Error())
		return
	}

	s.successResponse(c, models.BrandAliasResponse{
		ID:            alias.ID,
		BrandID:       alias.BrandID,
		Alias:         alias.Alias,
		CaseSensitive: alias.CaseSensitive,
		CreatedAt:     alias.CreatedAt,
		UpdatedAt:     alias.UpdatedAt,
	})
}

func (s *Server) deleteBrandAlias(c *gin.Context) {
	if err := s.brandsService.DeleteAlias(c.Request.Context(), c.Param("id"), c.Param("aliasId")); err != nil {
		s.errorResponse(c, http.StatusNotFound, "Failed to delete brand alias: "+err.Error())
		return
	}
	s.successResponse(c, gin.H{"id": c.Param("aliasId")})
}

func (s *Server) mapBrandFromDetection(c *gin.Context) {
	var req models.MapBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	brand, err := s.brandsService.MapFromDetection(
		c.Request.Context(),
		req.Alias,
		req.Name,
		req.CaseSensitive,
	)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to map brand: "+err.Error())
		return
	}
	s.successResponse(c, toBrandResponse(brand))
}
