package util

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const maxPathLen = 1024

// CleanRelPath проверяет путь от клиента или из БД: относительный, без "..",
// без NUL и обратных слэшей. Возвращает очищенный путь с "/" в качестве разделителя.
func CleanRelPath(p string) (string, error) {
	if p == "" || len(p) > maxPathLen || strings.ContainsRune(p, 0) || strings.Contains(p, `\`) {
		return "", fmt.Errorf("invalid path %q", p)
	}
	c := path.Clean(p)
	if path.IsAbs(c) || c == "." || c == ".." || strings.HasPrefix(c, "../") {
		return "", fmt.Errorf("invalid path %q", p)
	}
	return c, nil
}

// ProjectDir возвращает абсолютный путь к каталогу проекта: <root>/<projectPath>.
// projectPath (относительный путь из БД) проверяется, выйти за root он не может.
func ProjectDir(root, projectPath string) (string, error) {
	rel, err := CleanRelPath(projectPath)
	if err != nil {
		return "", fmt.Errorf("project path: %w", err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(absRoot, filepath.FromSlash(rel)), nil
}

func openProjectRoot(dir string) (*os.Root, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return os.OpenRoot(dir)
}

// CreateFileExclusive создаёт пустой файл (с каталогами) и возвращает ошибку,
// для которой errors.Is(err, fs.ErrExist) == true, если файл уже есть.
// Проверка и создание атомарны (O_EXCL).
func CreateFileExclusive(dir, rel string) error {
	root, err := openProjectRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()

	if d := path.Dir(rel); d != "." {
		if err := root.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	f, err := root.OpenFile(rel, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}

// WriteFileAtomic записывает файл через временный файл и rename, чтобы программа
// в контейнере не увидела наполовину записанное содержимое. Права существующего
// файла сохраняются (например, +x, выставленный в терминале).
func WriteFileAtomic(dir, rel string, data []byte) error {
	root, err := openProjectRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()

	if d := path.Dir(rel); d != "." {
		if err := root.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	mode := os.FileMode(0o644)
	if info, err := root.Stat(rel); err == nil {
		mode = info.Mode().Perm()
	}

	tmp := fmt.Sprintf("%s.%d.tmp", rel, time.Now().UnixNano())
	f, err := root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		_ = root.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = root.Remove(tmp)
		return err
	}
	if err := root.Rename(tmp, rel); err != nil {
		_ = root.Remove(tmp)
		return err
	}
	return nil
}

// RemoveFile удаляет файл внутри каталога проекта.
func RemoveFile(dir, rel string) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	return root.Remove(rel)
}
