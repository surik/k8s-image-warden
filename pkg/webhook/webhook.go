package webhook

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/surik/k8s-image-warden/pkg/engine"

	"github.com/gin-gonic/gin"
)

type WebhookServer struct {
	endpoint string
	certFile string
	keyFile  string
	r        *gin.Engine
	srv      *http.Server
}

func NewWebhookServer(endpoint, certFile, keyFile string) (*WebhookServer, error) {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	r := gin.New()
	r.Use(gin.Recovery())

	srv := &http.Server{
		Addr:    endpoint,
		Handler: r,
	}

	return &WebhookServer{
		endpoint: endpoint,
		certFile: certFile,
		keyFile:  keyFile,
		r:        r,
		srv:      srv,
	}, nil
}

func (wh *WebhookServer) Run(ctx context.Context, engine *engine.Engine) {
	wh.r.POST("/mutate", func(c *gin.Context) {
		mutateHandler(engine, c)
	})

	wh.r.POST("/validate", func(c *gin.Context) {
		validateHandler(engine, c)
	})

	log.Printf("Listening webhook on %s", wh.endpoint)
	if wh.certFile == "" || wh.keyFile == "" {
		if err := wh.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	} else {
		if err := wh.srv.ListenAndServeTLS(wh.certFile, wh.keyFile); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}
}

func (wh *WebhookServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := wh.srv.Shutdown(ctx); err != nil {
		log.Println(err)
	}
	log.Println("Webhook Server was shutdown")
}
