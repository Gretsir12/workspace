package util

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/moby/moby/client"
)

const (
	// Зафиксируйте конкретный тег (проверьте актуальный на Docker Hub).
	nixBaseImage = "nixos/nix:2.24.9"

	maxDeps      = 50
	pullTimeout  = 10 * time.Minute
	buildTimeout = 15 * time.Minute

	labelManaged = "ide.managed"
)

var (
	// Атрибут nixpkgs, например: python313, nodejs_22, python313Packages.requests.
	// Строгая проверка обязательна: имя попадает в shell-команду внутри Dockerfile.
	depRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.+-]{0,99}$`)
	// Полный SHA коммита nixpkgs: фиксирует версии пакетов.
	revRe = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

// Docker — обёртка над клиентом Docker Engine API.
// Создавайте один раз при старте приложения и переиспользуйте.
type Docker struct {
	api        *client.Client
	nixpkgsRev string
}

// NewDocker создаёт клиент из окружения (DOCKER_HOST и т.д.).
// nixpkgsRev — 40-символьный SHA коммита nixpkgs, из которого берутся пакеты.
func NewDocker(nixpkgsRev string) (*Docker, error) {
	if !revRe.MatchString(nixpkgsRev) {
		return nil, errors.New("nixpkgsRev must be a 40-char lowercase hex commit SHA")
	}
	api, err := client.New(
		client.FromEnv,
		client.WithUserAgent("ide-backend/1.0"),
	)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	return &Docker{api: api, nixpkgsRev: nixpkgsRev}, nil
}

func (d *Docker) Close() error { return d.api.Close() }

// PullBaseImage заранее скачивает базовый образ (удобно вызвать при старте,
// чтобы первая пользовательская сборка не ждала pull).
func (d *Docker) PullBaseImage(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, pullTimeout)
	defer cancel()

	resp, err := d.api.ImagePull(ctx, nixBaseImage, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("pull %s: %w", nixBaseImage, err)
	}
	defer resp.Close()

	// Pull асинхронный: без дочитывания потока он не завершится.
	// Wait читает поток до конца и возвращает ошибку, если pull не удался.
	if err := resp.Wait(ctx); err != nil {
		return fmt.Errorf("pull %s: %w", nixBaseImage, err)
	}
	return nil
}

// BuildImageWithDeps собирает образ с указанными пакетами nixpkgs и возвращает его тег.
// Одинаковый набор deps даёт тот же тег; повторная сборка отработает из кэша слоёв.
// Лог сборки пишется в logs (может быть nil).
func (d *Docker) BuildImageWithDeps(ctx context.Context, deps []string, logs io.Writer) (string, error) {
	deps, err := normalizeDeps(deps)
	if err != nil {
		return "", err
	}
	tag := d.imageTag(deps)

	ctx, cancel := context.WithTimeout(ctx, buildTimeout)
	defer cancel()

	buildCtx, err := tarDockerfile(d.dockerfile(deps))
	if err != nil {
		return "", fmt.Errorf("make build context: %w", err)
	}

	res, err := d.api.ImageBuild(ctx, buildCtx, client.ImageBuildOptions{
		Tags:        []string{tag},
		Dockerfile:  "Dockerfile",
		Remove:      true,
		ForceRemove: true,
		Labels:      map[string]string{labelManaged: "true"},
	})
	if err != nil {
		return "", fmt.Errorf("image build: %w", err)
	}
	defer res.Body.Close()

	if err := readBuildStream(res.Body, logs); err != nil {
		return "", err
	}
	return tag, nil
}

// normalizeDeps валидирует, убирает дубли и сортирует пакеты.
func normalizeDeps(deps []string) ([]string, error) {
	seen := make(map[string]struct{}, len(deps))
	out := make([]string, 0, len(deps))
	for _, p := range deps {
		p = strings.TrimSpace(p)
		if !depRe.MatchString(p) || strings.Contains(p, "..") {
			return nil, fmt.Errorf("invalid dependency name %q", p)
		}
		if _, dup := seen[p]; dup {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	if len(out) > maxDeps {
		return nil, fmt.Errorf("too many dependencies: %d (max %d)", len(out), maxDeps)
	}
	sort.Strings(out)
	return out, nil
}

// imageTag — детерминированный тег: хэш от базового образа, ревизии nixpkgs и пакетов.
func (d *Docker) imageTag(deps []string) string {
	sum := sha256.Sum256([]byte(nixBaseImage + "\n" + d.nixpkgsRev + "\n" + strings.Join(deps, "\n")))
	return "ide-env:" + hex.EncodeToString(sum[:8])
}

// dockerfile генерирует Dockerfile. Пакеты ставятся в профиль /opt/env,
// а не в профиль root, чтобы код можно было запускать не от root.
func (d *Docker) dockerfile(deps []string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "FROM %s\n", nixBaseImage)
	if len(deps) > 0 {
		fmt.Fprintf(&b,
			"RUN mkdir -p /opt && nix-env --profile /opt/env "+
				"-f https://github.com/NixOS/nixpkgs/archive/%s.tar.gz -iA %s "+
				"&& nix-collect-garbage -d\n",
			d.nixpkgsRev, strings.Join(deps, " "))
	}
	b.WriteString("ENV PATH=/opt/env/bin:$PATH\n")
	b.WriteString("RUN /opt/env/bin/rustc --version && /opt/env/bin/cargo --version\n")
	return []byte(b.String())
}

// tarDockerfile упаковывает Dockerfile в tar (контекст сборки в памяти).
func tarDockerfile(df []byte) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{Name: "Dockerfile", Mode: 0o644, Size: int64(len(df))}); err != nil {
		return nil, err
	}
	if _, err := tw.Write(df); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return &buf, nil
}

// readBuildStream читает JSON-поток сборки до конца. Ошибка сборки приходит
// внутри потока (HTTP 200), поэтому поле "error" проверять обязательно.
func readBuildStream(r io.Reader, logs io.Writer) error {
	dec := json.NewDecoder(r)
	for {
		var msg struct {
			Stream string `json:"stream"`
			Error  string `json:"error"`
		}
		if err := dec.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("read build stream: %w", err)
		}
		if msg.Error != "" {
			return fmt.Errorf("build failed: %s", msg.Error)
		}
		if logs != nil && msg.Stream != "" {
			_, _ = io.WriteString(logs, msg.Stream)
		}
	}
}
