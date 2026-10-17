package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type LocalStorage struct {
	BasePath string
	BaseURL  string
}

func NewLocalStorage(
	basePath string,
	baseURL string,
) *LocalStorage {
	return &LocalStorage{
		BasePath: basePath,
		BaseURL:  strings.TrimRight(baseURL, "/"),
	}
}

// Save menyimpan file ke local storage.
func (s *LocalStorage) Save(
	file *multipart.FileHeader,
	folder string,
) (string, error) {
	if file == nil {
		return "", fmt.Errorf("file is required")
	}

	// Ambil extension dari nama file asli.
	ext := strings.ToLower(filepath.Ext(file.Filename))

	if ext == "" {
		return "", fmt.Errorf("file extension is required")
	}

	// Validasi extension yang diperbolehkan.
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		// allowed
	default:
		return "",
			fmt.Errorf(
				"unsupported file type: %s",
				ext,
			)
	}

	// Buat folder tujuan.
	targetDir := filepath.Join(
		s.BasePath,
		folder,
	)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "",
			fmt.Errorf(
				"failed to create storage directory: %w",
				err,
			)
	}

	// Generate nama file unik.
	filename := uuid.New().String() + ext

	targetPath := filepath.Join(
		targetDir,
		filename,
	)

	// Buka file upload.
	src, err := file.Open()
	if err != nil {
		return "",
			fmt.Errorf(
				"failed to open uploaded file: %w",
				err,
			)
	}
	defer src.Close()

	// Buat file tujuan.
	dst, err := os.Create(targetPath)
	if err != nil {
		return "",
			fmt.Errorf(
				"failed to create destination file: %w",
				err,
			)
	}
	defer dst.Close()

	// Copy file upload ke storage.
	if _, err := io.Copy(dst, src); err != nil {
		// Hapus file jika proses copy gagal.
		_ = os.Remove(targetPath)

		return "",
			fmt.Errorf(
				"failed to save file: %w",
				err,
			)
	}

	// Return path yang akan digunakan frontend.
	return "/" + strings.Trim(
		filepath.ToSlash(
			filepath.Join(
				s.BaseURL,
				folder,
				filename,
			),
		),
		"/",
	), nil
}

// SaveBytes menyimpan data file berupa byte ke local storage.
func (s *LocalStorage) SaveBytes(
	data []byte,
	filename string,
	folder string,
) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("file data is empty")
	}

	// Ambil extension dari nama file.
	ext := strings.ToLower(filepath.Ext(filename))

	if ext == "" {
		return "", fmt.Errorf("file extension is required")
	}

	// Validasi extension yang diperbolehkan.
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		// allowed
	default:
		return "",
			fmt.Errorf(
				"unsupported file type: %s",
				ext,
			)
	}

	// Buat folder tujuan.
	targetDir := filepath.Join(
		s.BasePath,
		folder,
	)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "",
			fmt.Errorf(
				"failed to create storage directory: %w",
				err,
			)
	}

	// Generate nama file unik.
	newFilename := uuid.New().String() + ext

	targetPath := filepath.Join(
		targetDir,
		newFilename,
	)

	// Simpan byte ke file.
	if err := os.WriteFile(
		targetPath,
		data,
		0644,
	); err != nil {
		return "",
			fmt.Errorf(
				"failed to save file: %w",
				err,
			)
	}

	// Return path yang bisa digunakan oleh frontend.
	return "/" + strings.Trim(
		filepath.ToSlash(
			filepath.Join(
				s.BaseURL,
				folder,
				newFilename,
			),
		),
		"/",
	), nil
}

// Move memindahkan file dari satu folder storage ke folder lainnya.
// Fungsi ini mempertahankan nama file yang sama.
func (s *LocalStorage) Move(
	filePath string,
	targetFolder string,
) (string, error) {
	if strings.TrimSpace(filePath) == "" {
		return "", fmt.Errorf("file path is required")
	}

	if strings.TrimSpace(targetFolder) == "" {
		return "", fmt.Errorf("target folder is required")
	}

	// Ambil path relatif dari URL storage.
	cleanPath := strings.TrimPrefix(
		filePath,
		s.BaseURL,
	)

	cleanPath = strings.TrimPrefix(
		cleanPath,
		"/",
	)

	// Hilangkan query string jika ada.
	if index := strings.Index(cleanPath, "?"); index >= 0 {
		cleanPath = cleanPath[:index]
	}

	// Pastikan path tidak kosong.
	if strings.TrimSpace(cleanPath) == "" {
		return "", fmt.Errorf("invalid storage file path")
	}

	// Resolve BasePath.
	basePath, err := filepath.Abs(s.BasePath)
	if err != nil {
		return "",
			fmt.Errorf(
				"failed to resolve storage path: %w",
				err,
			)
	}

	// Resolve source file.
	sourcePath := filepath.Join(
		s.BasePath,
		filepath.FromSlash(cleanPath),
	)

	sourcePathAbs, err := filepath.Abs(sourcePath)
	if err != nil {
		return "",
			fmt.Errorf(
				"failed to resolve source file path: %w",
				err,
			)
	}

	// Pastikan source masih berada di dalam BasePath.
	if !strings.HasPrefix(
		sourcePathAbs,
		basePath+string(os.PathSeparator),
	) {
		return "",
			fmt.Errorf("invalid source storage file path")
	}

	// Pastikan source file memang ada.
	info, err := os.Stat(sourcePathAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return "",
				fmt.Errorf("source file not found")
		}

		return "",
			fmt.Errorf(
				"failed to access source file: %w",
				err,
			)
	}

	if info.IsDir() {
		return "",
			fmt.Errorf("source path is a directory")
	}

	// Buat folder tujuan.
	targetDir := filepath.Join(
		s.BasePath,
		targetFolder,
	)

	targetDirAbs, err := filepath.Abs(targetDir)
	if err != nil {
		return "",
			fmt.Errorf(
				"failed to resolve target directory: %w",
				err,
			)
	}

	// Pastikan target folder masih berada di dalam BasePath.
	if !strings.HasPrefix(
		targetDirAbs,
		basePath+string(os.PathSeparator),
	) {
		return "",
			fmt.Errorf("invalid target storage folder")
	}

	if err := os.MkdirAll(
		targetDirAbs,
		0755,
	); err != nil {
		return "",
			fmt.Errorf(
				"failed to create target directory: %w",
				err,
			)
	}

	// Pertahankan nama file yang sama.
	filename := filepath.Base(sourcePathAbs)

	targetPath := filepath.Join(
		targetDirAbs,
		filename,
	)

	// Pindahkan file.
	if err := os.Rename(
		sourcePathAbs,
		targetPath,
	); err != nil {
		return "",
			fmt.Errorf(
				"failed to move file: %w",
				err,
			)
	}

	// Return URL baru.
	return "/" + strings.Trim(
		filepath.ToSlash(
			filepath.Join(
				s.BaseURL,
				targetFolder,
				filename,
			),
		),
		"/",
	), nil
}

// Delete menghapus file dari local storage.
func (s *LocalStorage) Delete(filePath string) error {
	if strings.TrimSpace(filePath) == "" {
		return nil
	}

	// Jika yang diberikan adalah absolute URL,
	// ambil bagian path-nya saja.
	if parsedURL, err := url.Parse(filePath); err == nil {
		if parsedURL.IsAbs() {
			filePath = parsedURL.Path
		}
	}

	// Ambil bagian path setelah BaseURL.
	cleanPath := strings.TrimPrefix(
		filePath,
		s.BaseURL,
	)

	// Hilangkan "/" di awal.
	cleanPath = strings.TrimPrefix(
		cleanPath,
		"/",
	)

	// Hilangkan query string jika ada.
	if index := strings.Index(cleanPath, "?"); index >= 0 {
		cleanPath = cleanPath[:index]
	}

	// Pastikan path tidak kosong.
	if strings.TrimSpace(cleanPath) == "" {
		return nil
	}

	// Resolve BasePath menjadi absolute path.
	basePath, err := filepath.Abs(s.BasePath)
	if err != nil {
		return fmt.Errorf(
			"failed to resolve storage path: %w",
			err,
		)
	}

	// Buat target file berdasarkan BasePath.
	targetPath := filepath.Join(
		basePath,
		filepath.FromSlash(cleanPath),
	)

	// Resolve target menjadi absolute path.
	targetPathAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf(
			"failed to resolve file path: %w",
			err,
		)
	}

	// Pastikan file masih berada di dalam BasePath.
	if !strings.HasPrefix(
		targetPathAbs,
		basePath+string(os.PathSeparator),
	) {
		return fmt.Errorf(
			"invalid storage file path",
		)
	}

	// Hapus file.
	if err := os.Remove(targetPathAbs); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf(
			"failed to delete file: %w",
			err,
		)
	}

	return nil
}

// CleanupTempFiles menghapus file di folder temp
// yang sudah lebih dari maxAge.
func (s *LocalStorage) CleanupTempFiles(maxAge time.Duration) error {
	tempPath := filepath.Join(s.BasePath, "temp")

	entries, err := os.ReadDir(tempPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	now := time.Now()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(tempPath, entry.Name())

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > maxAge {
			if err := os.Remove(filePath); err != nil {
				continue
			}
		}
	}

	return nil
}
