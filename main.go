package main

import (
	"log"
	"net/http"

	"github.com/adamavix/thinksystem/package/kubefactory"
	"github.com/adamavix/thinksystem/package/response"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	client = kubefactory.NewKubeClient("gcr.io", "thinksystem")
)

func main() {
	log.Println("starting")
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/*", NotFound)
	r.Get("/{clusterName}", DeployCluster)
	r.Get("/{clusterName}/{appType}/{appName}/{appPort}", DeployLocalApp)
	r.Get("/{clusterName}/{appType}/{appName}/{appPort}/{image}/{imageName}", DeployRemoteApp)

	http.ListenAndServe(":81", r)
}

// NotFound redirects to the not found page
func NotFound(w http.ResponseWriter, r *http.Request) {
	response := response.CreateResponse()
	response.SendDataWithStatusCode(w, "not found", http.StatusNotFound)
}

func DeployCluster(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "clusterName")
	response := response.CreateResponse()
	log.Println(name)

	configs := createDefaultConfigs(response, name)

	for _, config := range configs {
		err := client.Deploy(config)
		if err != nil {
			log.Println(err)
			response.SendErrorWithStatusCode(w, err, http.StatusInternalServerError)
			return
		}
	}

	response.SendWithStatusCode(w, http.StatusAccepted)
}

func createDefaultConfigs(response *response.Response, clusterName string) []*kubefactory.KubeConfig {
	configs := []*kubefactory.KubeConfig{}

	mongo, err := client.NewKubeConfig(clusterName, "database", "mongo", "27017", "mongo", "mongo")
	if err != nil {
		log.Println(err)
		response.AppendError(err)
		return configs
	}
	configs = append(configs, mongo)

	flow, err := client.NewKubeConfig(clusterName, "backend", "flow", "81", "", "")
	if err != nil {
		log.Println(err)
		response.AppendError(err)
		return configs
	}
	configs = append(configs, flow)

	dashboard, err := client.NewKubeConfig(clusterName, "frontend", "dashboard", "80", "", "")
	if err != nil {
		log.Println(err)
		response.AppendError(err)
		return configs
	}
	configs = append(configs, dashboard)

	return configs
}

func DeployLocalApp(w http.ResponseWriter, r *http.Request) {
	response := response.CreateResponse()
	clusterName := chi.URLParam(r, "clusterName")
	appType := chi.URLParam(r, "appType")
	appName := chi.URLParam(r, "appName")
	appPort := chi.URLParam(r, "appPort")

	config, err := client.NewKubeConfig(clusterName, appType, appName, appPort, "", "")
	if err != nil {
		log.Println(err)
		response.SendErrorWithStatusCode(w, err, http.StatusInternalServerError)
		return
	}

	if err := client.Deploy(config); err != nil {
		log.Println(err)
		response.SendErrorWithStatusCode(w, err, http.StatusInternalServerError)
		return
	}

	response.SendWithStatusCode(w, http.StatusAccepted)
}

func DeployRemoteApp(w http.ResponseWriter, r *http.Request) {
	response := response.CreateResponse()
	clusterName := chi.URLParam(r, "clusterName")
	appType := chi.URLParam(r, "appType")
	appName := chi.URLParam(r, "appName")
	appPort := chi.URLParam(r, "appPort")
	image := chi.URLParam(r, "image")
	imageName := chi.URLParam(r, "imageName")

	config, err := client.NewKubeConfig(clusterName, appType, appName, appPort, image, imageName)
	if err != nil {
		log.Println(err)
		response.SendErrorWithStatusCode(w, err, http.StatusInternalServerError)
		return
	}

	if err := client.Deploy(config); err != nil {
		log.Println(err)
		response.SendErrorWithStatusCode(w, err, http.StatusInternalServerError)
		return
	}

	response.SendWithStatusCode(w, http.StatusAccepted)
}
