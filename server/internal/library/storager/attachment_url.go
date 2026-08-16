package storager

import (
	"context"
	"net/url"
	"strings"
)

// CdnUrl 根据 storage + 库中相对路径，返回浏览器可访问 URL。
// - local：/attachment/upload/... 相对路径（库中 object key，不含 /api 前缀）
// - 云存储：读取当前 sys_config 对应驱动的 domain，拼完整 CDN URL
// 注意：Nuxt 前台通过 resolveResourceUrl 在展示时加 /api 前缀，不入库。
func CdnUrl(ctx context.Context, storage, rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return rawURL
	}
	if isAbsoluteURL(rawURL) {
		return rawURL
	}
	key := CdnNormalizeKey(rawURL)
	if key == "" {
		return rawURL
	}
	return cdnResolveWithStorage(ctx, normalizeDriverName(storage), key)
}

// CdnUrlByPath 无 storage 字段时的兜底（如 CMS cover 只存了 path）。
// 路径像 attachment/upload → 本地；若全局 oss_driver 已为云存储则映射到 CDN。
func CdnUrlByPath(ctx context.Context, rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" || isAbsoluteURL(rawURL) {
		return rawURL
	}
	key := CdnNormalizeKey(rawURL)
	if key == "" {
		return rawURL
	}
	cfg := GetConfig(ctx)
	driver := normalizeDriverName(cfg.Driver)
	if isLocalObjectKey(key) {
		if driver == "" || driver == "local" {
			return toLocalAccessURL(key)
		}
		cloudKey := CloudObjectKeyFromLocal("/"+key, driver, cfg)
		if full := cdnDriverFullURL(ctx, driver, cloudKey); full != "" {
			return full
		}
		return toLocalAccessURL(key)
	}
	return cdnResolveWithStorage(ctx, cfg.Driver, key)
}

// CdnObjectKey 删除/底层存储使用的 object key（不含域名、不含前导 /）。
func CdnObjectKey(storage, rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return rawURL
	}
	if isAbsoluteURL(rawURL) {
		if u, err := url.Parse(rawURL); err == nil && u.Path != "" {
			return strings.TrimLeft(u.Path, "/")
		}
	}
	return CdnNormalizeKey(rawURL)
}

// CdnNormalizeKey 统一 object key（无前导 /）。
func CdnNormalizeKey(rawURL string) string {
	return strings.TrimLeft(strings.TrimSpace(rawURL), "/")
}

func isAbsoluteURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func isLocalObjectKey(key string) bool {
	return strings.HasPrefix(key, "attachment/upload/") || strings.HasPrefix(key, "attachment/")
}

func toLocalAccessURL(key string) string {
	if strings.HasPrefix(key, "/") {
		return key
	}
	return "/" + key
}

func cdnResolveWithStorage(ctx context.Context, storage, key string) string {
	storage = normalizeDriverName(storage)
	if storage == "" || storage == "local" {
		return toLocalAccessURL(key)
	}
	if full := cdnDriverFullURL(ctx, storage, key); full != "" {
		return full
	}
	return toLocalAccessURL(key)
}

func cdnDriverFullURL(ctx context.Context, storage, key string) string {
	cfg := GetConfig(ctx)
	switch normalizeDriverName(storage) {
	case "qiniu", "qn":
		if d, err := NewQiniu(cfg.Qiniu); err == nil {
			return d.GetFullUrl(key)
		}
	case "aliyun-oss", "aliyun", "oss":
		if d, err := NewAliyunOSS(cfg.Aliyun); err == nil {
			return d.GetFullUrl(key)
		}
	case "tencent-cos", "tencent", "cos":
		if d, err := NewTencentCOS(cfg.Tencent); err == nil {
			return d.GetFullUrl(key)
		}
	}
	return ""
}
