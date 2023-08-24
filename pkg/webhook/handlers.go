package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/surik/k8s-image-warden/pkg/engine"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Patch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func mutateHandler(engine *engine.Engine, c *gin.Context) {
	review, err := getAdmissionReview(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	log.Printf("mutate for %s\n", review.Request.Name)

	pod, err := getPodFromAdmissionReview(review)
	if err != nil {
		reject(c, review, http.StatusForbidden, err.Error())
		return
	}

	patches := mutate(c, engine, pod)
	if len(patches) > 0 {
		allowWithPatches(c, review, patches)
	} else {
		allow(c, review)
	}
}

func validateHandler(engine *engine.Engine, c *gin.Context) {
	review, err := getAdmissionReview(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	log.Printf("validate for %s\n", review.Request.Name)

	pod, err := getPodFromAdmissionReview(review)
	if err != nil {
		reject(c, review, http.StatusForbidden, err.Error())
		return
	}

	valid, message := validate(c, engine, pod)
	if valid {
		allow(c, review)
	} else {
		reject(c, review, http.StatusForbidden, message)
	}
}

func reject(c *gin.Context, review *admissionv1.AdmissionReview, code int, message string) {
	data := admissionv1.AdmissionReview{
		TypeMeta: review.TypeMeta,
		Response: &admissionv1.AdmissionResponse{
			UID:     review.Request.UID,
			Allowed: false,
			Result:  &metav1.Status{Code: int32(code), Message: message},
		},
	}

	c.JSON(http.StatusOK, data)
}

func allow(c *gin.Context, review *admissionv1.AdmissionReview) {
	data := admissionv1.AdmissionReview{
		TypeMeta: review.TypeMeta,
		Response: &admissionv1.AdmissionResponse{
			UID:     review.Request.UID,
			Allowed: true,
		},
	}

	c.JSON(http.StatusOK, data)
}

func allowWithPatches(c *gin.Context, review *admissionv1.AdmissionReview, patches []Patch) {
	bytes, err := json.Marshal(patches)
	if err != nil {
		reject(c, review, http.StatusInternalServerError, err.Error())
		return
	}

	jsonPatch := admissionv1.PatchTypeJSONPatch

	data := admissionv1.AdmissionReview{
		TypeMeta: review.TypeMeta,
		Response: &admissionv1.AdmissionResponse{
			UID:       review.Request.UID,
			Allowed:   true,
			PatchType: &jsonPatch,
			Patch:     bytes,
		},
	}

	c.JSON(http.StatusOK, data)
}

func getAdmissionReview(c *gin.Context) (*admissionv1.AdmissionReview, error) {
	var review admissionv1.AdmissionReview
	if err := c.Bind(&review); err != nil {
		return nil, err
	}

	return &review, nil
}

func getPodFromAdmissionReview(review *admissionv1.AdmissionReview) (*corev1.Pod, error) {
	raw := review.Request.Object.Raw
	pod := corev1.Pod{}
	if err := json.Unmarshal(raw, &pod); err != nil {
		return nil, err
	}
	return &pod, nil
}

func validate(ctx context.Context, ruleEngine *engine.Engine, pod *corev1.Pod) (bool, string) {
	containers := make([]corev1.Container, 0, len(pod.Spec.Containers)+len(pod.Spec.InitContainers))

	containers = append(containers, pod.Spec.Containers...)
	containers = append(containers, pod.Spec.InitContainers...)

	log.Printf("validate containers: %d\n", len(containers))

	for _, container := range containers {
		result, rule := ruleEngine.Validate(ctx, container.Image)
		if !result {
			return false, fmt.Sprintf("'%s' is not allowed by rule '%s'", container.Image, rule)
		}
	}

	return true, ""
}

func mutate(ctx context.Context, ruleEngine *engine.Engine, pod *corev1.Pod) []Patch {
	var patches []Patch

	initContainers := pod.Spec.InitContainers

	log.Printf("mutating init containers: %d\n", len(initContainers))

	for i, container := range initContainers {
		image, rules := ruleEngine.Mutate(ctx, container.Image)
		if len(rules) > 0 {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/spec/initContainers/" + strconv.Itoa(i) + "/image",
				Value: image,
			})
		}
	}

	containers := pod.Spec.Containers

	log.Printf("mutating containers: %d\n", len(containers))

	for i, container := range containers {
		image, rules := ruleEngine.Mutate(ctx, container.Image)
		if len(rules) > 0 {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/spec/containers/" + strconv.Itoa(i) + "/image",
				Value: image,
			})
		}
	}

	return patches
}
