package handler

import (
	"net/http"
)

func (h *Handler) AdminMembers(w http.ResponseWriter, r *http.Request) {
	ms, err := h.Store.ListMembers(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": ms})
}

func (h *Handler) AdminStats(w http.ResponseWriter, r *http.Request) {
	var members, movies, reviews, votes int64

	err := h.Store.Pool.QueryRow(r.Context(), `
		SELECT
			(SELECT COUNT(*) FROM members),
			(SELECT COUNT(*) FROM movies),
			(SELECT COUNT(*) FROM reviews),
			(SELECT COUNT(*) FROM votes)
	`).Scan(&members, &movies, &reviews, &votes)

	if err != nil {
		internal(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"members": members,
		"movies":  movies,
		"reviews": reviews,
		"votes":   votes,
	})
}
