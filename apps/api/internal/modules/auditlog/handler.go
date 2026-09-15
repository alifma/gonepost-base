package auditlog

import (
	"net/http"
	"strconv"

	httpx "basecode/api/internal/platform/http"
	"basecode/api/internal/platform/openapigen"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func toEventResponse(e Event) openapigen.AuditLogEvent {
	resp := openapigen.AuditLogEvent{
		Id:          e.ID,
		ActorUserId: e.ActorUserID,
		Action:      e.Action,
		Resource:    e.Resource,
		ResourceId:  e.ResourceID,
		Result:      openapigen.AuditLogEventResult(e.Result),
		CreatedAt:   e.CreatedAt,
	}
	if e.RequestID != "" {
		resp.RequestId = &e.RequestID
	}
	if len(e.Metadata) > 0 {
		resp.Metadata = &e.Metadata
	}
	return resp
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := ListFilter{
		Action:      q.Get("action"),
		ActorUserID: q.Get("actor_user_id"),
		Limit:       atoiDefault(q.Get("limit"), 20),
		Offset:      atoiDefault(q.Get("offset"), 0),
	}

	list, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	resp := openapigen.AuditLogListResponse{Data: make([]openapigen.AuditLogEvent, 0, len(list)), Total: total, Limit: filter.Limit}
	for _, e := range list {
		resp.Data = append(resp.Data, toEventResponse(e))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
