package helper

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	stdDraw "image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	maxIncomingFileBytes = 25 << 20
	maxImageDimension    = 1600
	jpegQuality          = 75
)

type UploadFileResult struct {
	SecureURL string
	PublicID  string
}

func UploadDir() string {
	root := strings.TrimSpace(os.Getenv("UPLOAD_DIR"))
	if root == "" {
		root = filepath.Join(".", "uploads")
	}

	abs, err := filepath.Abs(root)
	if err == nil {
		return abs
	}
	return filepath.Clean(root)
}

func UploadImage(file multipart.File, folder string) (string, error) {
	relativeURL, err := saveLocalFile(file, folder, "", true)
	if err != nil {
		return "", err
	}
	return makePublicURL(relativeURL), nil
}

func UploadFile(file multipart.File, folder string, fileName string) (*UploadFileResult, error) {
	relativeURL, err := saveLocalFile(file, folder, fileName, false)
	if err != nil {
		return nil, err
	}

	return &UploadFileResult{
		SecureURL: makePublicURL(relativeURL),
		PublicID:  strings.TrimPrefix(relativeURL, "/uploads/"),
	}, nil
}

func saveLocalFile(file multipart.File, folder, originalName string, requireImage bool) (string, error) {
	if file == nil {
		return "", errors.New("file upload kosong")
	}

	data, err := io.ReadAll(io.LimitReader(file, maxIncomingFileBytes+1))
	if err != nil {
		return "", fmt.Errorf("gagal membaca file upload: %w", err)
	}
	if len(data) == 0 {
		return "", errors.New("file upload kosong")
	}
	if len(data) > maxIncomingFileBytes {
		return "", fmt.Errorf("ukuran file melebihi batas internal %d MB", maxIncomingFileBytes>>20)
	}

	storedData := data
	ext := safeFileExtension(originalName)

	decoded, format, decodeErr := image.Decode(bytes.NewReader(data))
	if decodeErr == nil {
		ext = extensionForImageFormat(format)

		if format != "gif" {
			optimized, optimizeErr := compressToJPEG(decoded)
			if optimizeErr != nil {
				return "", fmt.Errorf("gagal mengompres gambar: %w", optimizeErr)
			}

			if len(optimized) < len(data) {
				storedData = optimized
				ext = ".jpg"
			}
		}
	} else if requireImage {
		return "", errors.New("format gambar tidak didukung; gunakan JPG, PNG, atau WebP")
	}

	if ext == "" {
		contentType := http.DetectContentType(data)
		if exts, _ := mime.ExtensionsByType(contentType); len(exts) > 0 {
			ext = safeFileExtension("file" + exts[0])
		}
		if ext == "" {
			ext = ".bin"
		}
	}

	cleanFolder, err := sanitizeFolder(folder)
	if err != nil {
		return "", err
	}

	id, err := randomID()
	if err != nil {
		return "", fmt.Errorf("gagal membuat nama file: %w", err)
	}
	storedName := id + ext

	root := UploadDir()
	directory := filepath.Join(root, cleanFolder)
	if err := os.MkdirAll(directory, 0750); err != nil {
		return "", fmt.Errorf("gagal membuat direktori upload: %w", err)
	}

	destination := filepath.Join(directory, storedName)
	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan file secara lokal: %w", err)
	}

	if _, err = out.Write(storedData); err != nil {
		_ = out.Close()
		_ = os.Remove(destination)
		return "", fmt.Errorf("gagal menulis file upload: %w", err)
	}
	if err = out.Close(); err != nil {
		_ = os.Remove(destination)
		return "", fmt.Errorf("gagal menutup file upload: %w", err)
	}

	urlPath := path.Join("/uploads", filepath.ToSlash(cleanFolder), storedName)
	if !strings.HasPrefix(urlPath, "/uploads/") {
		return "", errors.New("path upload tidak valid")
	}
	return urlPath, nil
}

func compressToJPEG(src image.Image) ([]byte, error) {
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, errors.New("dimensi gambar tidak valid")
	}

	longest := width
	if height > longest {
		longest = height
	}
	if longest > maxImageDimension {
		ratio := float64(maxImageDimension) / float64(longest)
		width = max(1, int(float64(width)*ratio))
		height = max(1, int(float64(height)*ratio))
	}

	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	xdraw.CatmullRom.Scale(resized, resized.Bounds(), src, bounds, xdraw.Src, nil)

	flattened := image.NewRGBA(resized.Bounds())
	stdDraw.Draw(flattened, flattened.Bounds(), image.NewUniform(color.White), image.Point{}, stdDraw.Src)
	stdDraw.Draw(flattened, flattened.Bounds(), resized, image.Point{}, stdDraw.Over)

	var out bytes.Buffer
	if err := jpeg.Encode(&out, flattened, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func extensionForImageFormat(format string) string {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return ".jpg"
	case "png":
		return ".png"
	case "gif":
		return ".gif"
	case "webp":
		return ".webp"
	default:
		return ".img"
	}
}

func safeFileExtension(fileName string) string {
	if strings.TrimSpace(fileName) == "" {
		return ""
	}
	ext := strings.ToLower(filepath.Ext(filepath.Base(fileName)))
	if len(ext) < 2 || len(ext) > 10 {
		return ""
	}
	for _, r := range ext[1:] {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') {
			return ""
		}
	}
	return ext
}

func sanitizeFolder(folder string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(strings.TrimSpace(folder)))
	if clean == "." {
		return "", nil
	}
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", errors.New("folder upload tidak valid")
	}
	return clean, nil
}

func randomID() (string, error) {
	var bytes16 [16]byte
	if _, err := rand.Read(bytes16[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes16[:]), nil
}

func makePublicURL(relativeURL string) string {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")), "/")
	if base == "" {
		return relativeURL
	}
	return base + relativeURL
}
