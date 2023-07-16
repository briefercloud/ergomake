package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	defaultBufferSize = 32 * 1024 * 1024  // 32MB
	maxBufferSize     = 512 * 1024 * 1024 // 512MB
)

func getBufferSize(fileSize int64) int {
	bufferSize := defaultBufferSize
	if fileSize > defaultBufferSize {
		bufferSize = int(fileSize)
		if bufferSize > maxBufferSize {
			bufferSize = maxBufferSize
		}
	}

	return bufferSize
}

func archiveDirectory(directory string) (string, error) {
	tempDir, err := ioutil.TempDir("", "archive")
	if err != nil {
		return "", err
	}

	archivePath := filepath.Join(tempDir, "archive.tar.gz")

	// Create the output tarball file
	outputFile, err := os.Create(archivePath)
	if err != nil {
		return "", err
	}
	defer outputFile.Close()

	// Create a gzip writer
	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	// Create a tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Walk through the current directory and add files to the tarball
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "fail while walking, path %s", path)
		}

		// Get the relative path of the file
		relPath, err := filepath.Rel(directory, path)
		if err != nil {
			return errors.Wrapf(err, "fail to get relative path for %s", path)
		}

		// Create a new tar header with the relative path
		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return errors.Wrapf(err, "fail to extract tar header from %s", path)
		}

		// Update the name field in the header with the relative path
		header.Name = filepath.ToSlash(relPath)

		// Write the tar header
		err = tarWriter.WriteHeader(header)
		if err != nil {
			return errors.Wrapf(err, "fail to write tar header for %s", path)
		}

		// If it's a regular file, write the file content to the tarball
		if !info.IsDir() && info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return errors.Wrapf(err, "fail to open %s", path)
			}
			defer file.Close()

			fileSize := info.Size()
			bufferSize := getBufferSize(fileSize)
			buffer := make([]byte, bufferSize)

			_, err = io.CopyBuffer(tarWriter, file, buffer)
			if err != nil {
				return errors.Wrapf(err, "fail to copy %s into tar", path)
			}
		}

		return nil
	})

	return archivePath, err
}
