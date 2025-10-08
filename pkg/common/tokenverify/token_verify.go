package tokenverify

import (
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/1nterdigital/aka-im-tools/errs"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/constant"
)

const (
	TokenUser  = constant.NormalUser
	TokenAdmin = constant.AdminUser
)

type Token struct {
	Expires time.Duration
	Secret  string
}

func (t *Token) secret() jwt.Keyfunc {
	return func(_ *jwt.Token) (any, error) {
		return []byte(t.Secret), nil
	}
}

type claims struct {
	UserID     string
	UserType   int32
	PlatformID int32
	jwt.RegisteredClaims
}

func (t *Token) getToken(str string) (userID string, userType int32, err error) {
	token, err := jwt.ParseWithClaims(str, &claims{}, t.secret())
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return "", 0, errs.ErrTokenMalformed.Wrap()
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return "", 0, errs.ErrTokenExpired.Wrap()
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return "", 0, errs.ErrTokenNotValidYet.Wrap()
			}

			return "", 0, errs.ErrTokenUnknown.Wrap()
		}

		return "", 0, errs.ErrTokenNotValidYet.Wrap()
	}

	claims, ok := token.Claims.(*claims)
	if claims.PlatformID != 0 {
		return "", 0, errs.ErrTokenExpired.Wrap()
	}
	if ok && token.Valid {
		return claims.UserID, claims.UserType, nil
	}

	return "", 0, errs.ErrTokenNotValidYet.Wrap()
}

func (t *Token) GetToken(token string) (userID string, userType int32, err error) {
	userID, userType, err = t.getToken(token)
	if err != nil {
		return "", 0, err
	}

	if userType != TokenUser && userType != TokenAdmin {
		return "", 0, errs.ErrTokenUnknown.WrapMsg("token type unknown")
	}

	return userID, userType, nil
}
