// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gcp

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func setCookie(c *gin.Context, token *oauth2.Token) {
	c.SetCookie("access_token", token.AccessToken, 0, "", "", false, false)
	c.SetCookie("refresh_token", token.RefreshToken, 0, "", "", false, false)
	c.SetCookie("expiry", token.Expiry.Format(time.RFC3339), 0, "", "", false, false)
}
