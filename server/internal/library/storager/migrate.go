// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package storager

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/gfile"
)

const localPublicRoot = "resource/public"

// LocalPhysicalPath 根据附件表 url 解析本地磁盘路径。
func LocalPhysicalPath(storedURL string) string {
	key := CdnNormalizeKey(storedURL)
	if key == "" {
		return ""
	}
	candidates := []string{
		gfile.Join(localPublicRoot, key),
	}
	if strings.HasPrefix(key, "upload/") {
		candidates = append(candidates, gfile.Join(localPublicRoot, "attachment", key))
	}
	if strings.HasPrefix(key, "attachment/upload/") {
		rest := strings.TrimPrefix(key, "attachment/upload/")
		candidates = append(candidates, gfile.Join(localPublicRoot, "attachment/upload", rest))
	}
	for _, p := range candidates {
		if gfile.Exists(p) {
			return p
		}
	}
	return candidates[0]
}

// ReadAttachmentBytes 读取附件二进制：优先本地磁盘，其次通过可访问 URL HTTP 拉取。
// 线上环境常见只有 URL/静态服务可访问、磁盘无实体文件的情况。
func ReadAttachmentBytes(ctx context.Context, storage, storedURL string) ([]byte, error) {
	if physical := LocalPhysicalPath(storedURL); physical != "" && gfile.Exists(physical) {
		return gfile.GetBytes(physical), nil
	}
	urls := buildAttachmentFetchURLs(ctx, storage, storedURL)
	if len(urls) == 0 {
		return nil, gerror.New("本地文件不存在且无法解析拉取地址")
	}
	client := g.Client().Timeout(60 * time.Second)
	var lastErr error
	for _, fetchURL := range urls {
		data, err := httpGetAttachment(client, ctx, fetchURL)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, gerror.New("本地文件不存在且无法解析拉取地址")
}

func httpGetAttachment(client *gclient.Client, ctx context.Context, fetchURL string) ([]byte, error) {
	resp, err := client.Get(ctx, fetchURL)
	if err != nil {
		return nil, gerror.Wrapf(err, "HTTP 拉取附件失败")
	}
	defer resp.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, gerror.Newf("HTTP 拉取附件失败：状态码 %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, gerror.Wrap(err, "读取 HTTP 响应失败")
	}
	if len(data) == 0 {
		return nil, gerror.New("HTTP 拉取到的附件为空")
	}
	return data, nil
}

// AttachmentAvailable 判断附件是否可读（本地磁盘或 HTTP 可拉取）。
func AttachmentAvailable(ctx context.Context, storage, storedURL string) bool {
	if physical := LocalPhysicalPath(storedURL); physical != "" && gfile.Exists(physical) {
		return true
	}
	return len(buildAttachmentFetchURLs(ctx, storage, storedURL)) > 0
}

// buildAttachmentFetchURLs 构造 HTTP 拉取候选地址。
// 库中 url 不含 /api；Nuxt 等部署对外暴露 /api/attachment/...，需一并尝试。
func buildAttachmentFetchURLs(ctx context.Context, storage, storedURL string) []string {
	raw := strings.TrimSpace(storedURL)
	if raw == "" {
		return nil
	}
	if isAbsoluteURL(raw) {
		return []string{raw}
	}
	accessURL := CdnUrl(ctx, storage, raw)
	if isAbsoluteURL(accessURL) {
		return []string{accessURL}
	}
	base := localFetchBaseURL(ctx)
	if base == "" {
		return nil
	}
	path := strings.TrimLeft(accessURL, "/")
	if path == "" {
		return nil
	}
	base = strings.TrimRight(base, "/")
	urls := []string{base + "/" + path}
	if strings.HasPrefix(path, "attachment/") {
		urls = append(urls, base+"/api/"+path)
	}
	return urls
}

func localFetchBaseURL(ctx context.Context) string {
	if r := g.RequestFromCtx(ctx); r != nil {
		scheme := "http"
		if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
			scheme = "https"
		}
		host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
		if host == "" {
			host = strings.TrimSpace(r.Host)
		}
		if host != "" {
			return scheme + "://" + host
		}
	}
	addr := strings.TrimSpace(g.Cfg().MustGet(ctx, "server.address").String())
	if addr == "" {
		return "http://127.0.0.1:8000"
	}
	if strings.HasPrefix(addr, ":") {
		return "http://127.0.0.1" + addr
	}
	if strings.Contains(addr, "://") {
		return strings.TrimRight(addr, "/")
	}
	return "http://" + strings.TrimRight(addr, "/")
}

// CloudObjectKeyFromLocal 将本地附件 url 映射为目标云存储 object key。
func CloudObjectKeyFromLocal(storedURL string, targetDriver string, cfg *StorageConfig) string {
	key := CdnNormalizeKey(storedURL)
	key = strings.TrimPrefix(key, "attachment/upload/")
	prefix := strings.Trim(strings.TrimSpace(driverObjectPrefix(cfg, targetDriver)), "/")
	if prefix == "" {
		return key
	}
	if key == prefix || strings.HasPrefix(key, prefix+"/") {
		return key
	}
	return prefix + "/" + key
}

func driverObjectPrefix(cfg *StorageConfig, driver string) string {
	switch normalizeDriverName(driver) {
	case "qiniu", "qn":
		return cfg.Qiniu.Prefix
	case "aliyun-oss", "aliyun", "oss":
		return cfg.Aliyun.Prefix
	case "tencent-cos", "tencent", "cos":
		return cfg.Tencent.Prefix
	default:
		return "upload/"
	}
}

// StoredURLFromObjectKey 附件表入库格式（前导 /）。
func StoredURLFromObjectKey(objectKey string) string {
	key := CdnNormalizeKey(objectKey)
	if key == "" {
		return ""
	}
	return "/" + key
}

// NewTargetDriver 按指定驱动名创建存储实例（不依赖全局单例，用于迁移上传）。
func NewTargetDriver(ctx context.Context, driver string) (Storager, error) {
	cfg := GetConfig(ctx)
	cfg.Driver = normalizeDriverName(driver)
	return NewDriver(cfg)
}

// CanInitDriver 判断指定云存储驱动配置是否完整可用。
func CanInitDriver(ctx context.Context, driver string) bool {
	_, err := NewTargetDriver(ctx, driver)
	return err == nil
}

// NormalizeDriverInput 统一前端/配置传入的驱动名。
func NormalizeDriverInput(raw string) string {
	return normalizeDriverName(strings.TrimSpace(raw))
}

// TargetDriverFromConfig 读取 oss 配置中的目标云驱动（local 时返回空）。
func TargetDriverFromConfig(ctx context.Context) (string, *StorageConfig, error) {
	cfg := GetConfig(ctx)
	driver := normalizeDriverName(cfg.Driver)
	if driver == "" || driver == "local" {
		return "", cfg, nil
	}
	return driver, cfg, nil
}

// DriverLabel 驱动中文名（配置页展示用）。
func DriverLabel(driver string) string {
	switch normalizeDriverName(driver) {
	case "qiniu", "qn":
		return "七牛云"
	case "aliyun-oss", "aliyun", "oss":
		return "阿里云 OSS"
	case "tencent-cos", "tencent", "cos":
		return "腾讯云 COS"
	default:
		return driver
	}
}
