package handler

import (
	"encoding/json"
	"net/http"

	"github.com/user/deer-flow-go/internal/skills"
)

type SkillsListResponse struct {
	Skills []skills.Skill `json:"skills"`
}

type SkillUpdateRequest struct {
	Enabled bool `json:"enabled"`
}

type SkillsHandler struct {
	loader *skills.Loader
}

func NewSkillsHandler(loader *skills.Loader) *SkillsHandler {
	return &SkillsHandler{loader: loader}
}

func (h *SkillsHandler) List(w http.ResponseWriter, r *http.Request) {
	skillList := h.loader.List()

	response := SkillsListResponse{Skills: skillList}
	writeJSON(w, http.StatusOK, response)
}

func (h *SkillsHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "skill name is required", "")
		return
	}

	skill := h.loader.Get(name)
	if skill == nil {
		writeError(w, http.StatusNotFound, "skill not found", "")
		return
	}

	writeJSON(w, http.StatusOK, skill)
}

func (h *SkillsHandler) Update(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "skill name is required", "")
		return
	}

	var req SkillUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if err := h.loader.SetEnabled(name, req.Enabled); err != nil {
		writeError(w, http.StatusNotFound, "skill not found", "")
		return
	}

	skill := h.loader.Get(name)
	writeJSON(w, http.StatusOK, skill)
}
