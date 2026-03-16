package main

import (
	"log/slog"
	"net/http"

	"microservices-demo/common/messaging/userqueue"
	"microservices-demo/common/pb/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (*cntrlHandlers) index(w http.ResponseWriter, r *http.Request) {
	// span := trace.SpanFromContext(r.Context())
	// span.SetAttributes(attribute.String("request-id", string(r.Context().Value(reqIDKey{}).(string))))
	w.Write([]byte("OK"))
}

func (ctl *cntrlHandlers) setUser(w http.ResponseWriter, r *http.Request) {
	if err := ctl.messagePublisher.publishUser(r.Context(), userqueue.User{
		ID:   r.PathValue("id"),
		Name: "user name" + r.PathValue("id"),
		Data: "data string" + r.PathValue("id"),
	}); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Error("failed to publish user", "error", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("OK"))
}

func (ctl *cntrlHandlers) getUser(w http.ResponseWriter, r *http.Request) {
	userService := pb.NewUserServiceClient(ctl.userStoreClient)
	if rsp, err := userService.GetUser(r.Context(), &pb.GetUserRequest{Id: r.PathValue("id")}); err != nil {
		slog.Error("failed to get user grpc", "error", err)
		errStatus := status.Convert(err)
		if errStatus.Code() == codes.NotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else {
		w.Write([]byte("User: ID=" + rsp.GetUser().GetId() + ", Name=" + rsp.GetUser().GetName() + ", Data=" + rsp.GetUser().GetData()))
		return
	}
}
