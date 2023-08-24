package webhook_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/surik/k8s-image-warden/pkg/engine"
	"github.com/surik/k8s-image-warden/pkg/webhook"
	admissionv1 "k8s.io/api/admission/v1"
)

func TestHandlers_Validate(t *testing.T) {
	r := gin.Default()

	rules := []engine.Rule{
		{
			Name: "docker.io is default",
			MutationRule: engine.MutationRule{
				Type:     engine.MutationTypeDefaultRegistry,
				Registry: "docker.io",
			},
		},
		{
			Name: "No Latest",
			ValidationRule: engine.ValidationRule{
				Type:  engine.ValidateTypeLatest,
				Allow: false,
			},
		},
	}

	engine, err := engine.NewEngine(nil, nil, rules)
	require.NoError(t, err)

	r.POST("/mutate", func(c *gin.Context) {
		webhook.MutateHandler(engine, c)
	})
	r.POST("/validate", func(c *gin.Context) {
		webhook.ValidateHandler(engine, c)
	})

	t.Run("nginx:latest mutated to docker.io/nginx:latest", func(t *testing.T) {
		resp := makeRequst(t, r, "mutate", "../../testdata/admission_review.json")

		var patches []webhook.Patch
		err = json.Unmarshal(resp.Response.Patch, &patches)
		require.NoError(t, err)

		require.Len(t, patches, 1)
		require.Equal(t, "docker.io/nginx:latest", patches[0].Value)
	})

	t.Run("nginx:latest not valid because of latest tag", func(t *testing.T) {
		resp := makeRequst(t, r, "validate", "../../testdata/admission_review.json")
		require.Equal(t, false, resp.Response.Allowed)
		require.Contains(t, resp.Response.Result.Message, "No Latest")
	})
}

func makeRequst(t *testing.T, r *gin.Engine, action, filename string) *admissionv1.AdmissionReview {
	t.Helper()

	var review admissionv1.AdmissionReview

	file, err := os.ReadFile(filename)
	require.NoError(t, err)

	err = json.Unmarshal(file, &review)
	require.NoError(t, err)

	var b bytes.Buffer
	err = json.NewEncoder(&b).Encode(review)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/"+action, &b)
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)

	var resp admissionv1.AdmissionReview
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	return &resp
}
