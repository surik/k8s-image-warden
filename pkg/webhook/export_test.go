package webhook

import (
	"github.com/gin-gonic/gin"
	"github.com/surik/k8s-image-warden/pkg/engine"
)

func MutateHandler(engine *engine.Engine, c *gin.Context) {
	mutateHandler(engine, c)
}

func ValidateHandler(engine *engine.Engine, c *gin.Context) {
	validateHandler(engine, c)
}
