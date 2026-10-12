package report

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type ChunkInfo struct {
	Name string
	Size int64
	MD5  string
}

type FileInfo struct {
	Name string
	Size int64
	MD5  string
}

func WriteReadme(path, origName string, origSize int64, origMD5 string, chunks []ChunkInfo) error {
	var sb strings.Builder
	sb.WriteString("# du-split 切片信息\n\n")
	sb.WriteString("## 原始文件信息\n\n")
	sb.WriteString(fmt.Sprintf("- 文件名: %s\n", origName))
	sb.WriteString(fmt.Sprintf("- 大小: %d\n", origSize))
	sb.WriteString(fmt.Sprintf("- MD5: %s\n\n", origMD5))
	sb.WriteString("## 切片文件信息\n\n")
	sb.WriteString("| 文件名 | 大小 | MD5 |\n")
	sb.WriteString("|--------|------|-----|\n")
	for _, c := range chunks {
		sb.WriteString(fmt.Sprintf("| %s | %d | %s |\n", c.Name, c.Size, c.MD5))
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func ParseReadme(path string) (origName string, origSize int64, origMD5 string, chunks []ChunkInfo, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := string(data)

	nameRe := regexp.MustCompile(`(?m)^- 文件名: (.+)$`)
	sizeRe := regexp.MustCompile(`(?m)^- 大小: (\d+)$`)
	md5Re := regexp.MustCompile(`(?m)^- MD5: ([0-9a-f]{32})$`)

	if m := nameRe.FindStringSubmatch(content); m != nil {
		origName = strings.TrimSpace(m[1])
	}
	if m := sizeRe.FindStringSubmatch(content); m != nil {
		origSize, _ = strconv.ParseInt(m[1], 10, 64)
	}
	if m := md5Re.FindStringSubmatch(content); m != nil {
		origMD5 = strings.TrimSpace(m[1])
	}

	rowRe := regexp.MustCompile(`^\| (\S+) \| (\d+) \| ([0-9a-f]{32}) \|$`)
	for _, line := range strings.Split(content, "\n") {
		if m := rowRe.FindStringSubmatch(line); m != nil {
			size, _ := strconv.ParseInt(m[2], 10, 64)
			chunks = append(chunks, ChunkInfo{
				Name: m[1],
				Size: size,
				MD5:  m[3],
			})
		}
	}

	return
}

func ComputeMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	buf := make([]byte, 8192)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			h.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
