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
	filesModified, err := c.equalizeTimebase(filePaths)
	if err != nil {
		log.Println(err)
	}

	fileListPath, err := c.prepareFileList(filesModified)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i",
		fileListPath, "-c", "copy", filepath.Join(c.outputPath, c.outputName),
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v: %v", err, stderr.String())
	}

	c.cleanup(filesModified)
	return nil
}

func (c compiler) equalizeTimebase(filePaths []string) ([]string, error) {
	var filesModified []string
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
		if err := cmd.Run(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("skipped %v: %v", fileName, err))
		}
		filesModified = append(filesModified, newFileName)
	}
	return filesModified, errs
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

func (c compiler) cleanup(fileNames []string) {
	for _, fileName := range fileNames {
		os.Remove(filepath.Join(c.outputPath, fileName))
	}
}
