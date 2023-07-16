package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

func uploadArchive(archivePath string, uploadURL string) (*http.Response, error) {
	// Open the archive file
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to open archive %s", archivePath)
	}
	defer file.Close()

	// Create a new buffer to store the file content
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		return nil, errors.Wrapf(err, "fail to copy archive %s into buffer", archivePath)
	}

	// Create a new HTTP request with a multipart form
	requestBody := &bytes.Buffer{}
	multipartWriter := multipart.NewWriter(requestBody)
	fileWriter, err := multipartWriter.CreateFormFile("file", archivePath)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create form file")
	}

	// Copy the file content to the multipart writer
	if _, err := io.Copy(fileWriter, buffer); err != nil {
		return nil, errors.Wrap(err, "fail to copy buffer into form")
	}

	// Close the multipart writer
	if err := multipartWriter.Close(); err != nil {
		return nil, errors.Wrap(err, "fail to close muiltipart writer")
	}

	// Create a new HTTP POST request with the multipart form data
	request, err := http.NewRequest("POST", uploadURL, requestBody)
	if err != nil {
		return nil, errors.Wrap(err, "failt to create upload http request")
	}
	request.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// Send the HTTP request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "fail to make upload http request")
	}

	// Check the response status
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, errors.Wrapf(err, "upload failed with status %s", response.Status)
	}

	return response, nil
}
