package compiler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type compiler struct {
	sourcePath string
	outputPath string
	outputName string
}

const fileListName = "list.txt"

func New(sourcePath, outputPath, outputName string) compiler {
	return compiler{
		sourcePath: sourcePath,
		outputPath: outputPath,
		outputName: outputName,
	}
}

func (c compiler) Run() error {
	fileNames, err := find(c.sourcePath, ".mp4")
	if err != nil {
		return err
	}

	filesModified, err := c.equalizeTimebase(fileNames)
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

	return nil
}

func (c compiler) equalizeTimebase(fileNames []string) ([]string, error) {
	var filesModified []string
	var errs error

	for _, fileName := range fileNames {
		newFileName := fmt.Sprintf("%v_modified.mp4", strings.TrimSuffix(fileName, filepath.Ext(fileName)))
		newFilePath := filepath.Join(c.outputPath, newFileName)
		cmd := exec.Command(
			"ffmpeg", "-i",
			filepath.Join(c.sourcePath, fileName), "-c", "copy",
			"-video_track_timescale", "15360", newFilePath,
		)
		if err := cmd.Run(); err != nil {
			errs = errors.Join(errs, err)
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

func find(root, ext string) ([]string, error) {
	var fileNames []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(d.Name()) == ext {
			fileNames = append(fileNames, d.Name())
		}
		return nil
	})
	return fileNames, err
}
