package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/{instance}", DeployInstance)
	r.Get("/{instance}/{image}/{service}", DeployContainer)

	http.ListenAndServe(":80", r)
}

// DeployInstance initializes a new set of containers for
// a flow instance.
func DeployInstance(w http.ResponseWriter, r *http.Request) {
	instance := chi.URLParam(r, "instance")

	execCmd("helm chart pull us-central1-docker.pkg.dev/${PROJECT_ID}/repo/deploy:0.1.0")

	backendFlowImage := "backend-flow"
	deployBackendFlow := fmt.Sprintf(
		"helm upgrade --install -f ./helm/deploy/values.yaml %s ./helm/deploy --set name=%s-service-%s-%s",
		backendFlowImage, instance, "backend", backendFlowImage,
	)

	execCmd(deployBackendFlow)

	frontendDashboardImage := "frontend-dashboard"
	deployFrontendDashboard := fmt.Sprintf(
		"helm upgrade --install -f ./helm/deploy/values.yaml %s ./helm/deploy --set name=%s-service-%s-%s",
		frontendDashboardImage, instance, "frontend", frontendDashboardImage,
	)

	execCmd(deployFrontendDashboard)
}

// DeployContainer deploys a container inside of an
// instance. The image is the name of the image and
// the service is either frontend or backend.
func DeployContainer(w http.ResponseWriter, r *http.Request) {
	instance := chi.URLParam(r, "instance")
	image := chi.URLParam(r, "image")
	service := chi.URLParam(r, "type")

	execCmd("helm chart pull us-central1-docker.pkg.dev/${PROJECT_ID}/repo/deploy:0.1.0")

	deploy := fmt.Sprintf(
		"helm upgrade --install -f ./helm/deploy/values.yaml %s ./helm/deploy --set name=%s-service-%s-%s",
		image, instance, service, image,
	)

	execCmd(deploy)
}

func execCmd(command string) {
	log.Printf("executing command %s", command)

	fields := strings.Fields(command)
	cmd := exec.Command(fields[0], fields[1:]...)
	stdoutStderr, err := cmd.CombinedOutput()

	fmt.Printf("%s\n", stdoutStderr)
	handleError(err)
}

func handleError(err error) {
	if err != nil {
		log.Println(err)
	}
}
