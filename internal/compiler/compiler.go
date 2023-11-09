package compiler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type compiler struct {
	outputPath string
	outputName string
}

const fileListName = "list.txt"

func New(outputPath, outputName string) compiler {
	return compiler{
		outputPath: outputPath,
		outputName: outputName,
	}
}

func (c compiler) Run(filePaths []string) error {
	modifiedFileNames, err := c.equalizeTimebase(filePaths)
	if err != nil {
		log.Println(err)
	}

	fileListPath, err := c.prepareFileList(modifiedFileNames)
	if err != nil {
		return fmt.Errorf("unable to prepare file list: %v", err)
	}

	cmd := exec.Command(
		"ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i",
		fileListPath, "-c", "copy", filepath.Join(c.outputPath, c.outputName),
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to compile clips: %v: %v", err, stderr.String())
	}

	toRemove := []string{fileListPath}
	toRemove = append(toRemove, filePaths...)
	for _, name := range modifiedFileNames {
		toRemove = append(toRemove, filepath.Join(c.outputPath, name))
	}

	err = remove(toRemove)
	if err != nil {
		return err
	}

	return nil
}

func (c compiler) equalizeTimebase(filePaths []string) ([]string, error) {
	var modifiedFileNames []string
	var errs error

	for _, path := range filePaths {
		fileName := filepath.Base(path)
		newFileName := fmt.Sprintf("%v_modified.mp4", strings.TrimSuffix(fileName, filepath.Ext(fileName)))
		newPath := filepath.Join(c.outputPath, newFileName)
		cmd := exec.Command(
			"ffmpeg", "-i",
			path, "-c", "copy",
			"-video_track_timescale", "15360", newPath,
		)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("skipped %v: %v: %v", fileName, err, stderr.String()))
		}
		modifiedFileNames = append(modifiedFileNames, newFileName)
	}
	return modifiedFileNames, errs
}

func (c compiler) prepareFileList(fileNames []string) (string, error) {
	path := filepath.Join(c.outputPath, fileListName)
	fileList, err := os.Create(path)
	if err != nil {
		return "", err
	}

	errAppendName := appendFileNames(fileNames, fileList)
	if errAppendName != nil {
		fileList.Close()
		errRemoveFile := os.Remove(path)
		return "", errors.Join(errAppendName, errRemoveFile)
	}
	defer fileList.Close()
	return path, nil
}

func appendFileNames(fileNames []string, dest io.Writer) error {
	for _, fileName := range fileNames {
		_, err := dest.Write([]byte(fmt.Sprintf("file '%v'\n", fileName)))
		if err != nil {
			return err
		}
	}
	return nil
}

func remove(filePaths []string) error {
	var errs error
	for _, path := range filePaths {
		err := os.Remove(path)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}
