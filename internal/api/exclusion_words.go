package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/models"
)

func (s *Server) listExclusionWords(c *gin.Context) {
	words, err := s.exclusionWordsService.ListExclusionWords(c.Request.Context())
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list exclusion words: "+err.Error())
		return
	}

	responses := make([]models.ExclusionWordResponse, len(words))
	for i, word := range words {
		responses[i] = models.ExclusionWordResponse{
			ID:        word.ID,
			Word:      word.Word,
			CreatedAt: word.CreatedAt,
			UpdatedAt: word.UpdatedAt,
		}
	}
	s.successResponse(c, responses)
}

func (s *Server) createExclusionWord(c *gin.Context) {
	var req models.CreateExclusionWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	word, err := s.exclusionWordsService.CreateExclusionWord(c.Request.Context(), req.Word)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to create exclusion word: "+err.Error())
		return
	}

	s.successResponse(c, models.ExclusionWordResponse{
		ID:        word.ID,
		Word:      word.Word,
		CreatedAt: word.CreatedAt,
		UpdatedAt: word.UpdatedAt,
	})
}

func (s *Server) deleteExclusionWord(c *gin.Context) {
	id := c.Param("id")
	if err := s.exclusionWordsService.DeleteExclusionWord(c.Request.Context(), id); err != nil {
		s.errorResponse(c, http.StatusNotFound, "Failed to delete exclusion word: "+err.Error())
		return
	}
	s.successResponse(c, gin.H{"id": id})
}

func (s *Server) listSuggestedBrandWords(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 {
		limit = 50
	}

	suggestions, err := s.exclusionWordsService.GetSuggestedBrandWords(c.Request.Context(), limit)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to get brand word suggestions: "+err.Error())
		return
	}
	s.successResponse(c, suggestions)
}
