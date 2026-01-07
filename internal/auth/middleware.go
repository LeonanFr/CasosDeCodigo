package auth

import (
	"context"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type contextKey string

const userContextKey = contextKey("user")
const isGuestKey = contextKey("isGuest")

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		guestHeader := r.Header.Get("X-Guest-ID")

		if authHeader != "" {
			claims, err := ValidateToken(authHeader)
			if err == nil {
				userID, err := primitive.ObjectIDFromHex(claims.UserID)
				if err == nil {
					ctx := context.WithValue(r.Context(), userContextKey, userID)
					ctx = context.WithValue(ctx, isGuestKey, false)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		var guestID primitive.ObjectID
		if guestHeader != "" {
			id, err := primitive.ObjectIDFromHex(guestHeader)
			if err == nil {
				guestID = id
			} else {
				guestID = primitive.NewObjectID()
			}
		} else {
			guestID = primitive.NewObjectID()
		}

		w.Header().Set("X-Guest-ID", guestID.Hex())
		ctx := context.WithValue(r.Context(), userContextKey, guestID)
		ctx = context.WithValue(ctx, isGuestKey, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (primitive.ObjectID, bool) {
	userID, ok := ctx.Value(userContextKey).(primitive.ObjectID)
	return userID, ok
}

func IsGuest(ctx context.Context) bool {
	guest, ok := ctx.Value(isGuestKey).(bool)
	return ok && guest
}
