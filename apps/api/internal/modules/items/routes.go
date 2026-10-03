package items

import "net/http"

// Register mounts the item endpoints. Each route needs a session plus
// items:read (GET) or items:write (everything else). Data is always scoped
// to the caller.
//
// A new module copies this shape: one Register per module keeps main.go
// down to a single line per feature.
func (h *Handler) Register(mux *http.ServeMux, requireAuth, requireRead, requireWrite func(http.Handler) http.Handler) {
	read := func(fn http.HandlerFunc) http.Handler { return requireAuth(requireRead(fn)) }
	write := func(fn http.HandlerFunc) http.Handler { return requireAuth(requireWrite(fn)) }
	withID := func(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) { fn(w, r, r.PathValue("id")) }
	}

	mux.Handle("GET /api/v1/items", read(h.List))
	mux.Handle("POST /api/v1/items", write(h.Create))
	mux.Handle("GET /api/v1/items/{id}", read(withID(h.Get)))
	mux.Handle("PATCH /api/v1/items/{id}", write(withID(h.Update)))
	mux.Handle("DELETE /api/v1/items/{id}", write(withID(h.Delete)))
}
