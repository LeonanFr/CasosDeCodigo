package auth

import (
	"context"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type contextKey string

const (
	userContextKey    = contextKey("user")
	sessionContextKey = contextKey("session")
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionHeader := r.Header.Get("X-Session-ID")
		var sessionID primitive.ObjectID
		var err error
		if sessionHeader != "" && sessionHeader != "null" && sessionHeader != "undefined" {
			sessionID, err = primitive.ObjectIDFromHex(sessionHeader)
		}
		if err != nil || sessionHeader == "" {
			sessionID = primitive.NewObjectID()
		}
		w.Header().Set("X-Session-ID", sessionID.Hex())
		ctx := context.WithValue(r.Context(), sessionContextKey, sessionID)

		authHeader := r.Header.Get("Authorization")
		guestHeader := r.Header.Get("X-Guest-ID")

		if authHeader != "" {
			claims, err := ValidateToken(authHeader)
			if err == nil {
				userID, err := primitive.ObjectIDFromHex(claims.UserID)
				if err == nil {
					ctx = context.WithValue(ctx, userContextKey, userID)
				}
			}
		} else if guestHeader != "" && guestHeader != "null" && guestHeader != "undefined" {
			guestID, err := primitive.ObjectIDFromHex(guestHeader)
			if err == nil {
				ctx = context.WithValue(ctx, userContextKey, guestID)
			} else {
				guestID = primitive.NewObjectID()
				w.Header().Set("X-Guest-ID", guestID.Hex())
				ctx = context.WithValue(ctx, userContextKey, guestID)
			}
		} else {
			guestID := primitive.NewObjectID()
			w.Header().Set("X-Guest-ID", guestID.Hex())
			ctx = context.WithValue(ctx, userContextKey, guestID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (primitive.ObjectID, bool) {
	userID, ok := ctx.Value(userContextKey).(primitive.ObjectID)
	return userID, ok
}

func GetSessionIDFromContext(ctx context.Context) (primitive.ObjectID, bool) {
	sessionID, ok := ctx.Value(sessionContextKey).(primitive.ObjectID)
	return sessionID, ok
}
