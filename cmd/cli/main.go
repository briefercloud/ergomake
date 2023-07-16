package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/briandowns/spinner"
)

func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Archiving..."
	s.Start()
	defer s.Stop()

	archivePath, err := archiveDirectory(currentDir)
	if err != nil {
		panic(err)
	}

	s.Suffix = " Uploading..."
	uploadURL := "http://localhost:8080/v2/deploy"
	res, err := uploadArchive(archivePath, uploadURL)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	sse(res.Body, func(e message, err error, done bool) {
		if err != nil {
			panic(err)
		}

		if done {
			return
		}

		var data map[string]string
		switch e.event {
		case "progress":
			err := json.Unmarshal([]byte(e.data), &data)
			if err != nil {
				s.FinalMSG = "Oops! Something went wrong. Try again later."
				return
			}
			switch data["status"] {
			case "pending":
				s.Suffix = " Preparing project"
			case "building":
				s.Suffix = " Building"
			case "deploying":
				s.Suffix = " Deploying"
			}
		case "finish":
			err := json.Unmarshal([]byte(e.data), &data)
			if err != nil {
				s.FinalMSG = "Oops! Something went wrong. Try again later."
				return
			}
			switch data["status"] {
			case "error":
				s.FinalMSG = "Oops! Something went wrong. Try again later."
			case "validation":
				s.FinalMSG = data["reason"]
				if s.FinalMSG == "" {
					s.FinalMSG = "Valdation error"
				}
			case "success":
				s.FinalMSG = data["url"]
			}
		}
	})
}
