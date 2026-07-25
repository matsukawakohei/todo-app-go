package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/matsukawakohei/todo-app-go/domain"
	"github.com/matsukawakohei/todo-app-go/internal/api"
	"github.com/matsukawakohei/todo-app-go/usecase"
)

type LabelHandler struct {
	lu usecase.LabelUsecase
}

func NewLabelHandler(lu usecase.LabelUsecase) *LabelHandler {
	return &LabelHandler{lu}
}

func toLabelResponse(label *usecase.Label) api.Label {
	return api.Label{
		Id:   int(label.ID),
		Name: label.Name,
	}
}

func (lh *LabelHandler) GetLabels(ctx context.Context, _ api.GetLabelsRequestObject) (api.GetLabelsResponseObject, error) {
	userID, err := getLoginUserFromContext(ctx)
	if err != nil {
		return api.GetLabels500JSONResponse{Error: err.Error()}, nil
	}

	outputs, err := lh.lu.GetAllByUserID(ctx, domain.UserID(userID))
	if err != nil {
		return api.GetLabels500JSONResponse{Error: err.Error()}, nil
	}

	labels := make([]api.Label, 0, len(outputs.Labels))
	for _, l := range outputs.Labels {
		labels = append(labels, toLabelResponse(l))
	}

	return api.GetLabels200JSONResponse{Data: labels}, nil
}

func (lh *LabelHandler) CreateLabel(ctx context.Context, request api.CreateLabelRequestObject) (api.CreateLabelResponseObject, error) {
	if request.Body == nil {
		slog.ErrorContext(ctx, "request body is nil")
		return api.CreateLabel400JSONResponse{Error: domain.ErrBadRequest.Error()}, nil
	}
	userID, err := getLoginUserFromContext(ctx)
	if err != nil {
		return api.CreateLabel500JSONResponse{Error: err.Error()}, nil
	}

	input := usecase.LabelCreateInput{Name: request.Body.Name}
	output, err := lh.lu.Create(ctx, domain.UserID(userID), input)
	if err != nil {
		message := err.Error()
		if errorStatus(err) == http.StatusInternalServerError {
			return api.CreateLabel500JSONResponse{Error: message}, nil
		}
		return api.CreateLabel400JSONResponse{Error: message}, nil
	}

	return api.CreateLabel201JSONResponse{Data: toLabelResponse(output)}, nil
}

func (lh *LabelHandler) DeleteLabel(ctx context.Context, request api.DeleteLabelRequestObject) (api.DeleteLabelResponseObject, error) {
	userID, err := getLoginUserFromContext(ctx)
	if err != nil {
		return api.DeleteLabel500JSONResponse{Error: err.Error()}, nil
	}

	if request.Id <= 0 {
		slog.ErrorContext(ctx, "id less than equals 0")
		return api.DeleteLabel400JSONResponse{Error: domain.ErrBadRequest.Error()}, nil
	}

	if err := lh.lu.Delete(ctx, domain.UserID(userID), domain.LabelID(request.Id)); err != nil {
		message := err.Error()
		switch errorStatus(err) {
		case http.StatusNotFound:
			return api.DeleteLabel404JSONResponse{Error: message}, nil
		case http.StatusInternalServerError:
			return api.DeleteLabel500JSONResponse{Error: message}, nil
		default:
			return api.DeleteLabel400JSONResponse{Error: message}, nil
		}
	}

	return api.DeleteLabel204Response{}, nil
}
