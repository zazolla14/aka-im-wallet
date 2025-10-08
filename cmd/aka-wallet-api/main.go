package main

import (
	"github.com/1nterdigital/aka-im-tools/system/program"
	_ "github.com/1nterdigital/aka-im-wallet/docs/swag/wallet"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/cmd"
)

// @title						Wallet API Documentation
// @version						1.0
// @description					This is the Wallet API server
// @termsOfService				http://swagger.io/terms/
// @contact.name				API Support
// @contact.url					http://www.swagger.io/support
// @contact.email				support@swagger.io
// @license.name				Apache 2.0
// @license.url					http://www.apache.org/licenses/LICENSE-2.0.html
// @host						ap11.0dev.cc
// @BasePath					/
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						token
func main() {
	if err := cmd.NewWalletApiCmd().Exec(); err != nil {
		program.ExitWithError(err)
	}
}
