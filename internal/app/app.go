package app

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/leo-jcn/du-split/internal/report"
)

const chunkSize = 5 * 1024 * 1024 // 5MB

func Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: du-split <split|merge> [options]")
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "split":
		return splitCmd(rest)
	case "merge":
		return mergeCmd(rest)
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func splitCmd(args []string) error {
	var outDir string
	var filePath string

	for i := 0; i < len(args); i++ {
		if args[i] == "-o" {
			if i+1 < len(args) {
				outDir = args[i+1]
				i++
			} else {
				return fmt.Errorf("-o requires an argument")
			}
		} else if strings.HasPrefix(args[i], "-o=") {
			outDir = args[i][3:]
		} else if !strings.HasPrefix(args[i], "-") {
			filePath = args[i]
		}
	}

	if filePath == "" {
		return fmt.Errorf("usage: du-split split <file> [-o <dir>]")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("cannot stat file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	origName := info.Name()
	dirName := strings.TrimSuffix(origName, filepath.Ext(origName))

	parentDir := outDir
	if parentDir == "" {
		parentDir, _ = os.Getwd()
	}

	outputDir := filepath.Join(parentDir, dirName)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	origMD5, chunks, err := splitFile(filePath, outputDir, origName)
	if err != nil {
		return err
	}

	readmePath := filepath.Join(outputDir, "read.md")
	if err := report.WriteReadme(readmePath, origName, info.Size(), origMD5, chunks); err != nil {
		return fmt.Errorf("cannot write read.md: %w", err)
	}

	fmt.Printf("split complete: %d chunks in %s\n", len(chunks), outputDir)
	return nil
}

func splitFile(filePath, outputDir, origName string) (string, []report.ChunkInfo, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	baseName := strings.TrimSuffix(origName, filepath.Ext(origName))

	fileMD5 := md5.New()
	var chunks []report.ChunkInfo
	buf := make([]byte, chunkSize)
	index := 1

	for {
		n, err := f.Read(buf)
		if n > 0 {
			fileMD5.Write(buf[:n])

			chunkName := fmt.Sprintf("%s_part_%03d", baseName, index)
			chunkPath := filepath.Join(outputDir, chunkName)

			if writeErr := os.WriteFile(chunkPath, buf[:n], 0644); writeErr != nil {
				return "", nil, fmt.Errorf("cannot write chunk: %w", writeErr)
			}

			chunkMD5 := md5.Sum(buf[:n])
			chunks = append(chunks, report.ChunkInfo{
				Name: chunkName,
				Size: int64(n),
				MD5:  hex.EncodeToString(chunkMD5[:]),
			})
			index++
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", nil, fmt.Errorf("read error: %w", err)
		}
	}

	return hex.EncodeToString(fileMD5.Sum(nil)), chunks, nil
}

func mergeCmd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: du-split merge <chunk-dir>")
	}
	chunkDir := args[0]

	info, err := os.Stat(chunkDir)
	if err != nil {
		return fmt.Errorf("cannot stat chunk directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", chunkDir)
	}

	origName, origSize, origMD5, chunks, err := report.ParseReadme(filepath.Join(chunkDir, "read.md"))
	if err != nil {
		return fmt.Errorf("cannot parse read.md: %w", err)
	}

	parentDir := filepath.Dir(chunkDir)
	outputPath := filepath.Join(parentDir, origName)

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer outFile.Close()

	for _, c := range chunks {
		chunkPath := filepath.Join(chunkDir, c.Name)
		data, err := os.ReadFile(chunkPath)
		if err != nil {
			return fmt.Errorf("cannot read chunk %s: %w", c.Name, err)
		}
		if _, err := outFile.Write(data); err != nil {
			return fmt.Errorf("cannot write to output: %w", err)
		}
	}

	outInfo, _ := os.Stat(outputPath)
	if outInfo.Size() != origSize {
		return fmt.Errorf("size mismatch: expected %d, got %d", origSize, outInfo.Size())
	}

	mergeMD5, err := fileMD5(outputPath)
	if err != nil {
		return err
	}
	if mergeMD5 != origMD5 {
		return fmt.Errorf("MD5 mismatch: expected %s, got %s", origMD5, mergeMD5)
	}

	fmt.Printf("merge complete: %s (%d bytes, md5=%s)\n", outputPath, outInfo.Size(), mergeMD5)
	return nil
}

func fileMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
