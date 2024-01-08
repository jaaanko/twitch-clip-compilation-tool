package compiler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type compiler struct {
	outputDir      string
	outputFileName string
	ffmpegPath     string
	cleanup        bool
}

const fileListName = "list.txt"

func New(options ...func(*compiler)) compiler {
	c := compiler{
		outputDir:      "out",
		outputFileName: "compilation.mp4",
		ffmpegPath:     "ffmpeg",
		cleanup:        true,
	}

	for _, opt := range options {
		opt(&c)
	}

	return c
}

func WithOutputDir(outputDir string) func(*compiler) {
	return func(c *compiler) {
		c.outputDir = outputDir
	}
}

func WithOutputFileName(outputFileName string) func(*compiler) {
	return func(c *compiler) {
		c.outputFileName = outputFileName
	}
}

func WithFFmpegPath(ffmpegPath string) func(*compiler) {
	return func(c *compiler) {
		c.ffmpegPath = ffmpegPath
	}
}

func WithCleanup(cleanup bool) func(*compiler) {
	return func(c *compiler) {
		c.cleanup = cleanup
	}
}

func (c compiler) Run(filePaths []string) error {
	modifiedFileNames, err := c.equalizeTimebase(filePaths)
	if err != nil {
		return err
	}

	fileListPath, err := c.prepareFileList(modifiedFileNames)
	if err != nil {
		return fmt.Errorf("unable to prepare file list: %v", err)
	}

	cmd := exec.Command(
		c.ffmpegPath, "-y", "-f", "concat", "-safe", "0", "-i",
		fileListPath, "-c", "copy", filepath.Join(c.outputDir, c.outputFileName),
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	filesToRemove := []string{fileListPath}
	for _, name := range modifiedFileNames {
		filesToRemove = append(filesToRemove, filepath.Join(c.outputDir, name))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to compile clips: %v: %v", err, stderr.String())
	}

	if err := removeAll(filesToRemove); err != nil {
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
		newPath := filepath.Join(c.outputDir, newFileName)
		cmd := exec.Command(
			c.ffmpegPath, "-i", path, "-c", "copy",
			"-video_track_timescale", "15360", newPath,
		)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("skipped %v: %v: %v", fileName, err, stderr.String()))
		}

		if c.cleanup {
			err := os.Remove(path)
			if err != nil {
				errs = errors.Join(errs, err)
			}
		}

		modifiedFileNames = append(modifiedFileNames, newFileName)
	}

	return modifiedFileNames, errs
}

func (c compiler) prepareFileList(fileNames []string) (string, error) {
	path := filepath.Join(c.outputDir, fileListName)
	fileList, err := os.Create(path)
	if err != nil {
		return "", err
	}

	defer fileList.Close()
	errAppendName := appendFileNames(fileNames, fileList)
	if errAppendName != nil {
		errRemoveFile := os.Remove(path)
		return "", errors.Join(errAppendName, errRemoveFile)
	}

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

func removeAll(filePaths []string) error {
	var errs error
	for _, path := range filePaths {
		err := os.Remove(path)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}
