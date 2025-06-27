package server

import (
	"fmt"
	"net/http"
	"server/internal/domain/service/auth"
	"server/pkg/contextx"
	"server/pkg/failure"
	"server/pkg/logx"
	"strings"
)

type AuthServer struct {
	auth *auth.AuthService
}

func NewAuthServer(auth *auth.AuthService) *AuthServer {
	return &AuthServer{auth: auth}
}

func (s *AuthServer) ApiHandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid form"+": "+err.Error()))
		return
	}

	session, err := s.auth.Login(ctx, r.Form.Get("username"), r.Form.Get("password"))
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	w.Header().Add("Token", session)

	// if user logged, try logout
	typeAndToken := strings.Split(r.Header.Get("Authorization"), " ")
	if len(typeAndToken) != 2 {
		return
	}
	if err := s.auth.Logout(ctx, typeAndToken[1]); err != nil {
		contextx.GetLoggerOrDefault(ctx).Error("logout auth user fail: ", logx.Error(err))
	}
}

func (s *AuthServer) ApiHandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	typeAndToken := strings.Split(r.Header.Get("Authorization"), " ")
	if len(typeAndToken) != 2 {
		return
	}

	if err := s.auth.Logout(ctx, typeAndToken[1]); err != nil {
		writeAndLogErr(ctx, w, err)
	}
}

func (s *AuthServer) MwAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		str := r.Header.Get("Authorization")
		if len(str) == 0 {
			str = r.FormValue("authorization")
		}

		typeAndToken := strings.Split(str, " ")
		if len(typeAndToken) != 2 {
			writeAndLogErr(ctx, w, failure.NewUnauthorizedError("invalid authorization header"+fmt.Sprint(typeAndToken)))
			return
		}
		if typeAndToken[0] != "Bearer" {
			writeAndLogErr(ctx, w, failure.NewUnauthorizedError("invalid auth type"))
			return
		}
		token := typeAndToken[1]
		ok, err := s.auth.CheckSession(ctx, token)
		if err != nil {
			writeAndLogErr(ctx, w, failure.NewUnauthorizedError(err.Error()))
			return
		}
		if !ok {
			writeAndLogErr(ctx, w, failure.NewUnauthorizedError("invalid token"))
			return
		}
		next(w, r)
	}
}
