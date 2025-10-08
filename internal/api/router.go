package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	middleware "github.com/1nterdigital/aka-im-tools/mw"
	http_api "github.com/1nterdigital/aka-im-wallet/internal/api/http"
	walletmw "github.com/1nterdigital/aka-im-wallet/internal/api/mw"
	walletapi "github.com/1nterdigital/aka-im-wallet/internal/service"
)

func SetRouter(svcName string, api *walletapi.Api, mw *walletmw.MW) *gin.Engine {
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Use(gin.Recovery(), middleware.CorsHandler(), middleware.GinParseOperationID())
	r.Use(mw.GinParseToken())
	r.Use(otelgin.Middleware(svcName))

	handler := http_api.NewWalletHandler(api)

	wallet := r.Group("wallet")
	wallet.GET("/detail", handler.GetWalletDetail)
	wallet.POST("/create", handler.CreateWallet)

	transaction := r.Group("/wallet_transactions")
	transaction.GET("/cash_flow", handler.GetListTransaction)
	transaction.POST("/create", handler.CreateTransaction)

	transfer := r.Group("/transfer")
	transfer.POST("/create", handler.CreateTransfer)
	transfer.POST("/claim", handler.ClaimTransfer)
	transfer.POST("/refund", handler.RefundTransfer)
	transfer.GET("/:transfer_id/detail", handler.GetDetailTransfer)

	envelope := r.Group("/envelope")
	envelope.POST("/", handler.CreateEnvelopeHandler)
	envelope.POST("/claim", handler.ClaimEnvelopeHandler)
	envelope.GET("/:envelope_id/details", handler.GetEnvelopeDetail)
	envelope.GET("/autoRefund/", handler.AutoRefundEnvelopeHandler)
	envelope.GET("/refund/:envelope_id", handler.RefundEnvelopeByID)

	// admin
	boRouter := r.Group("/bo", mw.CheckAdmin)

	boWalletRecharge := boRouter.Group("/deposit")
	boWalletRecharge.POST("/process", handler.ProcessDepositByAdmin)
	boWalletRecharge.GET("/list", handler.GetListDeposit)

	walletMonitoring := boRouter.Group("/wallet-monitoring")
	walletMonitoring.GET("/transaction-volume", handler.GetDashboardTransactionVolume)
	walletMonitoring.GET("/transactions", handler.GetListTransactionMonitoring)
	walletMonitoring.GET("/envelopes", handler.GetListEnvelope)
	walletMonitoring.GET("/envelopes/:envelope_id/details", handler.GetEnvelopeDetail)
	walletMonitoring.GET("/transfers", handler.GetTransferHistory)
	walletMonitoring.GET("/top-users", handler.GetTop10Users)

	boRouter.POST("/refund/manual_process", handler.ProcessManualRefund)

	boBalanceAdjustment := boRouter.Group("/balance_adjustment")
	boBalanceAdjustment.POST("/process", handler.BalanceAdjustmentByAdmin)
	boBalanceAdjustment.GET("/list", handler.GetListBalanceAdjustment)

	return r
}
