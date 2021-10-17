package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	response "github.com/thinksystemio/package-response"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", Echo)
	r.Get("/*", NotFound)

	r.Get("/api/{instance}", DeployInstance)
	r.Get("/api/{instance}/{image}", DeployContainer)

	http.ListenAndServe(":80", r)
}

// Echo allows pinging of this service
func Echo(w http.ResponseWriter, r *http.Request) {
	res := response.CreateResponse()
	res.SendDataWithStatusCode(w, "echo", http.StatusOK)
}

// NotFound redirects to the not found page
func NotFound(w http.ResponseWriter, r *http.Request) {
	res := response.CreateResponse()
	res.SendDataWithStatusCode(w, "not found", http.StatusNotFound)
}

// DeployInstance initializes a new set of containers for
// a flow instance.
func DeployInstance(w http.ResponseWriter, r *http.Request) {
	res := response.CreateResponse()

	instance := chi.URLParam(r, "instance")

	err := execCmd("helm chart pull us-central1-docker.pkg.dev/${PROJECT_ID}/repo/deploy:0.1.0")
	if err != nil {
		res.SendErrorWithStatusCode(w, err, http.StatusInternalServerError)
	}

	backendFlowImage := "backend-flow"
	deployBackendFlow := fmt.Sprintf(
		"helm upgrade --install -f /dist/helm/deploy/values.yaml %s /dist/helm/deploy --set name=%s-service-%s",
		backendFlowImage, instance, backendFlowImage,
	)

	err = execCmd(deployBackendFlow)
	if err != nil {
		res.SendErrorWithStatusCode(w, err, http.StatusInternalServerError)
	}

	frontendDashboardImage := "frontend-dashboard"
	deployFrontendDashboard := fmt.Sprintf(
		"helm upgrade --install -f /dist/helm/deploy/values.yaml %s /dist/helm/deploy --set name=%s-service-%s",
		frontendDashboardImage, instance, frontendDashboardImage,
	)

	err = execCmd(deployFrontendDashboard)
	if err != nil {
		res.SendErrorWithStatusCode(w, err, http.StatusInternalServerError)
	}

	res.SendDataWithStatusCode(w, "successfully deployed instance", http.StatusOK)
}

// DeployContainer deploys a container inside of an
// instance. The image is the name of the image and
// the service is either frontend or backend.
func DeployContainer(w http.ResponseWriter, r *http.Request) {
	res := response.CreateResponse()

	instance := chi.URLParam(r, "instance")
	image := chi.URLParam(r, "image")

	err := execCmd("helm charr pull us-central1-docker.pkg.dev/${PROJECT_ID}/repo/deploy:0.1.0")
	if err != nil {
		res.SendErrorWithStatusCode(w, err, http.StatusInternalServerError)
	}

	deploy := fmt.Sprintf(
		"helm upgrade --install -f /dist/helm/deploy/values.yaml %s /dist/helm/deploy --set name=%s-service-%s",
		image, instance, image,
	)

	err = execCmd(deploy)
	if err != nil {
		res.SendErrorWithStatusCode(w, err, http.StatusInternalServerError)
	}

	res.SendDataWithStatusCode(w, "successfully deployed instance", http.StatusOK)
}

func execCmd(command string) error {
	log.Printf("executing command %s", command)

	fields := strings.Fields(command)
	cmd := exec.Command(fields[0], fields[1:]...)
	stdoutStderr, err := cmd.CombinedOutput()

	fmt.Printf("%s\n", stdoutStderr)
	return err
}
