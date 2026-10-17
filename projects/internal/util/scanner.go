package util

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"unicode/utf8"
	"workspaces/projects/internal/models"
)

const maxFileSize = 5 << 20 // 5 МБ: файлы больше не читаем и в БД не кладём

var skipDirs = map[string]struct{}{
	".git": {}, "node_modules": {}, "target": {}, "__pycache__": {},
}

// ScanDirectory обходит проект и возвращает файлы с путями ОТНОСИТЕЛЬНО root
// (в формате со слэшами, например "src/main.go").
func ScanDirectory(root string, projectID uint32) ([]models.File, error) {
	var files []models.File

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil // файл исчез посреди обхода
			}
			return err
		}

		if d.IsDir() {
			if _, skip := skipDirs[d.Name()]; skip && path != root {
				return filepath.SkipDir
			}
			return nil
		}

		// Только обычные файлы: без симлинков, сокетов, девайсов.
		if !d.Type().IsRegular() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if info.Size() > maxFileSize {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("rel path %s: %w", path, err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if !utf8.Valid(data) {
			return nil // бинарник, в текстовое поле не кладём
		}

		files = append(files, models.File{
			ProjectID: projectID,
			Path:      filepath.ToSlash(rel),
			Text:      string(data),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// SnapshotDirectory возвращает хэш от путей, размеров и mtime всех файлов проекта.
func SnapshotDirectory(root string) (string, error) {
	h := sha256.New()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil // файл исчез посреди обхода (например, tmp-файл)
			}
			return err
		}
		if d.IsDir() {
			if _, skip := skipDirs[d.Name()]; skip && path != root {
				return filepath.SkipDir
			}
		}

		info, err := d.Info()
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}

		rel, _ := filepath.Rel(root, path)
		// WalkDir обходит в лексическом порядке, сортировка не нужна
		fmt.Fprintf(h, "%s|%t|%d|%d\n", rel, d.IsDir(), info.Size(), info.ModTime().UnixNano())
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
