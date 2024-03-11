package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func sendHTTPRequest(method string, url string, headers map[string]string, body io.Reader) *io.ReadCloser {
	req, err := http.NewRequest(method, url, body)
	ProcessError(err, fmt.Sprintf("Error in creating %s request (%s)", method, url))

	for header, headerValue := range headers {
		req.Header.Add(header, headerValue)
	}

	res, err := http.DefaultClient.Do(req)
	ProcessError(err, fmt.Sprintf("Error in http %s response (%s)", method, url))

	if res.StatusCode >= 400 {
		fmt.Printf("Response statusCode[%d] (%s:%s)", res.StatusCode, method, url)
	}

	return &res.Body
}

func MakeGETRequest(url string, headers map[string]string) *io.ReadCloser {
	return sendHTTPRequest(http.MethodGet, url, headers, nil)
}

func ProcessError(err error, errorMessage string) {
	if err != nil {
		fmt.Printf("%s: %v\n", errorMessage, err)
	}
}

func ExitOnError(err error, errorMessage string, errorCode int) {
	ProcessError(err, errorMessage)
	if err != nil {
		os.Exit(errorCode)
	}
}

func UntarFile(sourceFilePath string, destinationDirectoryPath string) {
	err := exec.Command("tar", "-xf", sourceFilePath, "-C", destinationDirectoryPath).Run()
	ProcessError(err, fmt.Sprintf("Error while file (%s) untar", sourceFilePath))
}

func WriteToFile(data []byte, filePath string) {
	file, err := os.OpenFile(filePath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0777)
	ProcessError(err, fmt.Sprintf("Error while creating file (%s)", filePath))

	_, err = file.Write(data)
	ProcessError(err, fmt.Sprintf("Error while writing data to file (%s)", filePath))

	defer file.Close()
}
