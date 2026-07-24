package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/matsukawakohei/todo-app-go/internal/api"
	"github.com/matsukawakohei/todo-app-go/usecase"
	"github.com/oapi-codegen/runtime/types"
)

type AuthHandler struct {
	uu usecase.UserUsecase
}

func NewAuthHandler(uu usecase.UserUsecase) *AuthHandler {
	return &AuthHandler{uu}
}

func newTokenCookie(value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     "token",
		Value:    value,
		Expires:  expires,
		Path:     "/",
		HttpOnly: true,
		Secure:   CookieSecure(),
		SameSite: http.SameSiteLaxMode,
	}
}

// loginJSONResponse はLogin成功時のレスポンスです。
// 生成された api.Login200JSONResponse はCookieの付与に対応していないため、
// 独自にVisitLoginResponseを実装しています。
type loginJSONResponse struct {
	body   api.LoginOutput
	cookie *http.Cookie
}

func (r loginJSONResponse) VisitLoginResponse(w http.ResponseWriter) error {
	http.SetCookie(w, r.cookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(r.body)
}

// logoutResponse はLogout時にCookieを失効させるための独自レスポンスです。
type logoutResponse struct {
	cookie *http.Cookie
}

func (r logoutResponse) VisitLogoutResponse(w http.ResponseWriter) error {
	http.SetCookie(w, r.cookie)
	w.WriteHeader(http.StatusOK)
	return nil
}

func (ah *AuthHandler) Signup(ctx context.Context, request api.SignupRequestObject) (api.SignupResponseObject, error) {
	input := usecase.SignUpInput{
		Name:     request.Body.Name,
		Email:    string(request.Body.Email),
		Password: request.Body.Password,
	}
	output, err := ah.uu.SignUp(ctx, input)
	if err != nil {
		message := err.Error()
		if errorStatus(err) == http.StatusInternalServerError {
			return api.Signup500JSONResponse{Error: message}, nil
		}
		return api.Signup400JSONResponse{Error: message}, nil
	}

	user := api.User{
		Id:    int(output.ID),
		Name:  output.Name,
		Email: types.Email(output.Email),
	}

	return api.Signup201JSONResponse{Data: user}, nil
}

func (ah *AuthHandler) Login(ctx context.Context, request api.LoginRequestObject) (api.LoginResponseObject, error) {
	input := usecase.LoginInput{
		Email:    string(request.Body.Email),
		Password: request.Body.Password,
	}
	output, err := ah.uu.Login(ctx, input)
	if err != nil {
		message := err.Error()
		switch errorStatus(err) {
		case http.StatusUnauthorized:
			return api.Login401JSONResponse{Error: message}, nil
		case http.StatusInternalServerError:
			return api.Login500JSONResponse{Error: message}, nil
		default:
			return api.Login400JSONResponse{Error: message}, nil
		}
	}

	cookie := newTokenCookie(output.Token, time.Now().Add(time.Hour*24))

	user := api.User{
		Id:    int(output.ID),
		Name:  output.Name,
		Email: types.Email(output.Email),
	}

	return loginJSONResponse{
		body:   api.LoginOutput{Data: user},
		cookie: cookie,
	}, nil
}

func (ah *AuthHandler) Logout(_ context.Context, _ api.LogoutRequestObject) (api.LogoutResponseObject, error) {
	cookie := newTokenCookie("", time.Now())
	cookie.MaxAge = -1

	return logoutResponse{cookie: cookie}, nil
}
