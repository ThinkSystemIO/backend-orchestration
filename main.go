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

	r.Get("/", deploy)

	http.ListenAndServe(":80", r)
}

func deploy(w http.ResponseWriter, r *http.Request) {
	execCmd("echo {$_KEY} | helm registry login -u _json_key_base64 --password-stdin https://us-central1-docker.pkg.dev")
	execCmd("helm chart pull us-central1-docker.pkg.dev/${PROJECT_ID}/repo/deploy:0.1.0")
	execCmd("helm chart export us-central1-docker.pkg.dev/${PROJECT_ID}/repo/deploy:0.1.0")
	execCmd("helm upgrade --install thinksystemio-${_ENV} thinksystemio -n ${_ENV} -f thinksystemio/values-${_ENV}.yaml")
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
		log.Fatalln(err)
	}
}
