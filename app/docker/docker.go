package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/codecrafters-io/docker-starter-go/app/util"
)

const registry = "registry.docker.io"
const repository = "library"
const mediaTypeHeader = "application/vnd.docker.distribution.manifest.v2+json"

type authTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Token       string `json:"token"`
}

type layer struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

type manifestResponse struct {
	Config struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"config"`
	Layers        []layer `json:"layers"`
	MediaType     string  `json:"mediaType"`
	SchemaVersion int     `json:"schemaVersion"`
}

type Image struct {
	name    string
	version string
}

func (image *Image) getAuthToken() authTokenResponse {
	url := fmt.Sprintf("https://auth.docker.io/token?service=%s&scope=repository:%s/%s:pull", registry, repository, image.name)

	responseBody := *util.MakeGETRequest(url, make(map[string]string))
	defer responseBody.Close()

	var response authTokenResponse
	err := json.NewDecoder(responseBody).Decode(&response)
	util.ProcessError(err, fmt.Sprintf("Error in decoding response json (%s)", url))
	return response
}

func (image *Image) getManifests(token string) manifestResponse {
	url := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/%s/manifests/%s", repository, image.name, image.version)

	headers := make(map[string]string)
	headers["Accept"] = mediaTypeHeader
	headers["Authorization"] = fmt.Sprintf("Bearer %s", token)

	responseBody := *util.MakeGETRequest(url, headers)
	defer responseBody.Close()

	var response manifestResponse
	err := json.NewDecoder(responseBody).Decode(&response)
	util.ProcessError(err, fmt.Sprintf("Error in decoding response json (%s)", url))
	return response
}

func (image *Image) pullLayers(token string, layers []layer, destinationDirectory string) {
	for _, layer := range layers {
		url := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/%s/blobs/%s", repository, image.name, layer.Digest)

		headers := make(map[string]string)
		headers["Accept"] = mediaTypeHeader
		headers["Authorization"] = fmt.Sprintf("Bearer %s", token)

		responseBody := *util.MakeGETRequest(url, headers)
		defer responseBody.Close()

		data, err := io.ReadAll(responseBody)
		util.ProcessError(err, fmt.Sprintf("Error in reading blob data of layer (%s)", layer.Digest))

		layerFile := fmt.Sprintf("%s:%s-%s.tar", image.name, image.version, layer.Digest)
		util.WriteToFile(data, layerFile)
		util.UntarFile(layerFile, destinationDirectory)
	}
}

func (image *Image) GetImageString() string {
	return image.name + ":" + image.version
}

func (image *Image) PullImage(destinationDirectory string) {
	tokenResponse := image.getAuthToken()
	manifests := image.getManifests(tokenResponse.Token)
	image.pullLayers(tokenResponse.Token, manifests.Layers, destinationDirectory)
}

func NewImage(imageString string) Image {
	imageName, version, notLatest := strings.Cut(imageString, ":")
	if !notLatest {
		version = "latest"
	}
	return Image{name: imageName, version: version}
}
