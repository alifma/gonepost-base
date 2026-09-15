package main

import (
	httpx "basecode/api/internal/platform/http"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"basecode/api/internal/config"
	"basecode/api/internal/middleware"
	"basecode/api/internal/modules/auditlog"
	"basecode/api/internal/modules/auth"
	"basecode/api/internal/modules/permissions"
	"basecode/api/internal/modules/roles"
	"basecode/api/internal/modules/users"
	"basecode/api/internal/platform/authctx"
	"basecode/api/internal/platform/database"
	"basecode/api/internal/platform/httpserver"
	"basecode/api/internal/platform/logger"
	"basecode/api/internal/platform/openapigen"
)

func main() {

	// bikin logger pakai package logger
	appLogger := logger.New()

	// bikin konfigurasi package config
	cfg, err := config.Load()
	if err != nil {
		appLogger.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	// buka koneksi database dan mastiin beneran nyambung
	pool, err := database.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		appLogger.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// modul auditlog: repository + service (dipake handler lain buat nyatet aksi sensitif)
	auditRepo := auditlog.NewRepository(pool)
	auditService := auditlog.NewService(auditRepo, appLogger)
	auditHandler := auditlog.NewHandler(auditRepo)

	// modul users: repository + handler
	usersRepo := users.NewRepository(pool)
	usersHandler := users.NewHandler(usersRepo, auditService)

	// modul auth: session repo + service (rate limit 5 percobaan / 15 menit) + handler
	sessionsRepo := auth.NewSessionRepository(pool)
	loginLimiter := auth.NewRateLimiter(5, 15*time.Minute)
	authService := auth.NewService(usersRepo, sessionsRepo, loginLimiter)
	authHandler := auth.NewHandler(authService, auditService, cfg.CookieSecure)
	requireAuth := auth.RequireAuth(authService)

	// modul roles: repository + authz service + handler
	rolesRepo := roles.NewRepository(pool)
	authz := roles.NewAuthzService(rolesRepo)
	rolesHandler := roles.NewHandler(rolesRepo, auditService)
	requireUsersRead := roles.RequirePermission(authz, auditService, permissions.UsersRead)
	requireUsersWrite := roles.RequirePermission(authz, auditService, permissions.UsersWrite)
	requireRolesRead := roles.RequirePermission(authz, auditService, permissions.RolesRead)
	requireRolesWrite := roles.RequirePermission(authz, auditService, permissions.RolesWrite)
	requireAuditRead := roles.RequirePermission(authz, auditService, permissions.AuditRead)

	// bikin router HTTP untuk daftarin endpoint HTTP
	mux := http.NewServeMux()

	// endpoint buat health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		appLogger.Info("health check", "request_id", middleware.FromContext(r.Context()))
		httpx.WriteJSON(w, http.StatusOK, openapigen.StatusResponse{Status: "ok"})
	})

	// endpoint buat check readyness
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, openapigen.StatusResponse{Status: "ok"})
		appLogger.Info("readiness check", "request_id", middleware.FromContext(r.Context()))
	})

	// auth endpoints
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("GET /api/v1/auth/me", authHandler.Me)
	mux.Handle("POST /api/v1/auth/password/change", requireAuth(http.HandlerFunc(authHandler.ChangePassword)))
	// bootstrap: cuma jalan sekali (selama belum ada SUPER_ADMIN), butuh login
	mux.Handle("POST /api/v1/auth/bootstrap-admin", requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := authctx.UserIDFromContext(r.Context())
		rolesHandler.BootstrapSuperAdmin(w, r, userID)
	})))

	// user endpoints — butuh login + permission users:read / users:write
	mux.Handle("POST /api/v1/users", requireAuth(requireUsersWrite(http.HandlerFunc(usersHandler.Create))))
	mux.Handle("GET /api/v1/users", requireAuth(requireUsersRead(http.HandlerFunc(usersHandler.List))))
	mux.Handle("GET /api/v1/users/{id}", requireAuth(requireUsersRead(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usersHandler.Get(w, r, r.PathValue("id"))
	}))))
	mux.Handle("PATCH /api/v1/users/{id}", requireAuth(requireUsersWrite(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usersHandler.Update(w, r, r.PathValue("id"))
	}))))
	mux.Handle("GET /api/v1/users/{id}/roles", requireAuth(requireUsersRead(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rolesHandler.ListUserRoles(w, r, r.PathValue("id"))
	}))))
	mux.Handle("POST /api/v1/users/{id}/roles", requireAuth(requireRolesWrite(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rolesHandler.AssignToUser(w, r, r.PathValue("id"))
	}))))
	mux.Handle("DELETE /api/v1/users/{id}/roles/{roleId}", requireAuth(requireRolesWrite(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rolesHandler.RemoveFromUser(w, r, r.PathValue("id"), r.PathValue("roleId"))
	}))))

	// role endpoints — butuh login + permission roles:read / roles:write
	mux.Handle("POST /api/v1/roles", requireAuth(requireRolesWrite(http.HandlerFunc(rolesHandler.Create))))
	mux.Handle("GET /api/v1/roles", requireAuth(requireRolesRead(http.HandlerFunc(rolesHandler.List))))
	mux.Handle("POST /api/v1/roles/{id}/permissions", requireAuth(requireRolesWrite(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rolesHandler.GrantPermission(w, r, r.PathValue("id"))
	}))))
	mux.Handle("GET /api/v1/roles/{id}/permissions", requireAuth(requireRolesRead(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rolesHandler.ListPermissions(w, r, r.PathValue("id"))
	}))))

	// audit log endpoint — read-only, permission-protected
	mux.Handle("GET /api/v1/audit-logs", requireAuth(requireAuditRead(http.HandlerFunc(auditHandler.List))))

	handler := http.Handler(mux)
	handler = middleware.SecurityHeaders(handler)
	handler = middleware.CORS(cfg.AllowedOrigins)(handler)
	handler = middleware.Recover(appLogger)(handler)
	handler = middleware.RequestID(handler)

	// nunggu sinyal shutdown biar server bisa berhenti dengan graceful
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// jalanin server dan catat error kalo server berhenti dadakan.
	appLogger.Info("api listening", "port", cfg.Port)
	if err := httpserver.Run(ctx, cfg.Port, handler); err != nil {
		appLogger.Error("server error", "err", err)
		os.Exit(1)
	}
}
