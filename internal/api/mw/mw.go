// Copyright © 2023 AkaIM open source community. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mw

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/1nterdigital/aka-im-tools/apiresp"
	"github.com/1nterdigital/aka-im-tools/errs"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/constant"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/db/database"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/tokenverify"
)

func New(token *tokenverify.Token, db database.WalletDatabaseInterface) *MW {
	return &MW{
		token:    token,
		Database: db,
	}
}

type MW struct {
	Database database.WalletDatabaseInterface
	token    *tokenverify.Token
}

func (o *MW) CheckToken(c *gin.Context) {
	userID, userType, _, err := o.parseToken(c)
	if err != nil {
		c.Abort()
		apiresp.GinError(c, err)
		return
	}
	setToken(c, userID, userType)
}

func (o *MW) CheckAdmin(c *gin.Context) {
	userID, _, err := o.parseTokenType(c, constant.AdminUser)
	if err != nil {
		c.Abort()
		apiresp.GinError(c, err)
		return
	}
	setToken(c, userID, constant.AdminUser)
}

func (o *MW) parseToken(c *gin.Context) (userID string, userType int32, userTypeValue string, err error) {
	token := c.GetHeader("token")
	if token == "" {
		return "", 0, "", errs.ErrArgs.WrapMsg("token is empty")
	}

	userID, userType, err = o.token.GetToken(token)
	if err != nil {
		return "", 0, "", err
	}

	m, err := o.Database.GetTokens(c, userID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", 0, "", err
	}

	if len(m) == 0 {
		return "", 0, "", errs.ErrTokenNotExist.Wrap()
	}

	return userID, userType, token, nil
}

func (o *MW) parseTokenType(c *gin.Context, userType int32) (userID, userToken string, err error) {
	userID, t, token, err := o.parseToken(c)
	if err != nil {
		return "", "", err
	}
	if t != userType {
		return "", "", errs.ErrArgs.WrapMsg("token type error")
	}
	return userID, token, nil
}

func setToken(c *gin.Context, userID string, userType int32) {
	SetToken(c, userID, userType)
}

func SetToken(c *gin.Context, userID string, userType int32) {
	c.Set(constant.RpcOpUserID, userID)
	c.Set(constant.RpcOpUserType, []string{strconv.Itoa(int(userType))})
	c.Set(constant.RpcCustomHeader, []string{constant.RpcOpUserType})
}

func (o *MW) GinParseToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, wApi := range whitelist {
			if strings.HasPrefix(c.Request.URL.Path, wApi) {
				c.Next()
				return
			}
		}

		userID, userType, _, err := o.parseToken(c)
		if err != nil {
			c.Abort()
			apiresp.GinError(c, err)
			return
		}
		setToken(c, userID, userType)
		c.Next()
	}
}

// Whitelist api not parse token
var whitelist = []string{
	"/health",
}
