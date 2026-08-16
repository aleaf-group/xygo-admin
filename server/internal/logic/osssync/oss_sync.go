package osssync

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"xygo/internal/dao"
	"xygo/internal/library/storager"
	"xygo/internal/model/do"
	"xygo/internal/model/entity"
	"xygo/internal/model/input/adminin"
	"xygo/internal/service"
	"xygo/utility"
)

type sOssSync struct{}

func init() {
	service.RegisterOssSync(New())
}

func New() *sOssSync {
	return &sOssSync{}
}

func (s *sOssSync) Preview(ctx context.Context, in *adminin.OssSyncPreviewInp) (*adminin.OssSyncPreviewModel, error) {
	driver, err := resolveSyncDriver(in.TargetDriver)
	if err != nil {
		return nil, err
	}

	localTotal, localBytes, err := countLocalAttachments(ctx)
	if err != nil {
		return nil, err
	}

	ready := storager.CanInitDriver(ctx, driver)
	reuploadTotal := 0
	if ready {
		reuploadTotal, err = countReuploadable(ctx, driver)
		if err != nil {
			return nil, err
		}
	}

	return &adminin.OssSyncPreviewModel{
		TargetDriver:  driver,
		TargetLabel:   storager.DriverLabel(driver),
		Ready:         ready,
		LocalTotal:    localTotal,
		ReuploadTotal: reuploadTotal,
		Total:         localTotal,
		TotalBytes:    localBytes,
	}, nil
}

func (s *sOssSync) Run(ctx context.Context, in *adminin.OssSyncRunInp) (*adminin.OssSyncRunModel, error) {
	driver, err := resolveSyncDriver(in.TargetDriver)
	if err != nil {
		return nil, err
	}
	if !storager.CanInitDriver(ctx, driver) {
		return nil, gerror.Newf("请先完善并保存%s配置后再同步", storager.DriverLabel(driver))
	}

	cfg := storager.GetConfig(ctx)
	target, err := storager.NewTargetDriver(ctx, driver)
	if err != nil {
		return nil, gerror.Wrapf(err, "初始化目标存储驱动失败")
	}

	size := in.PageSize
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	localTotal, _, _ := countLocalAttachments(ctx)
	reuploadTotal := 0
	if in.IncludeReupload {
		reuploadTotal, _ = countReuploadable(ctx, driver)
	}
	grandTotal := localTotal + reuploadTotal

	out := &adminin.OssSyncRunModel{
		Total:    grandTotal,
		Failures: make([]adminin.OssSyncFailureItem, 0),
	}

	localDrv := storager.NewLocal(cfg.Local)

	// 1) 本地 storage 记录（按 id 游标分批，失败也推进避免卡在同一批）
	localList, err := fetchLocalBatch(ctx, in.LastLocalId, size)
	if err != nil {
		return nil, err
	}
	lastLocalId := in.LastLocalId
	for _, item := range localList {
		out.Processed++
		lastLocalId = item.Id
		if err := s.syncOne(ctx, &item, driver, cfg, target, localDrv, in.DeleteLocal, in.DryRun, false); err != nil {
			out.Failed++
			out.Failures = append(out.Failures, adminin.OssSyncFailureItem{
				Id: item.Id, Name: item.Name, Url: item.Url, Error: err.Error(),
			})
			continue
		}
		out.Success++
	}
	out.LastLocalId = lastLocalId
	if len(localList) >= size {
		out.Done = false
		return out, nil
	}

	// 2) 重新上传（云端已删、本地仍在）
	if !in.IncludeReupload {
		out.Done = localTotal <= len(localList)
		if out.Total == 0 {
			out.Total = localTotal
		}
		return out, nil
	}

	reuploadList, lastReuploadId, err := fetchReuploadBatch(ctx, driver, in.LastReuploadId, size)
	if err != nil {
		return nil, err
	}
	out.LastReuploadId = lastReuploadId
	for _, item := range reuploadList {
		out.Processed++
		if err := s.syncOne(ctx, &item, driver, cfg, target, localDrv, in.DeleteLocal, in.DryRun, true); err != nil {
			out.Failed++
			out.Failures = append(out.Failures, adminin.OssSyncFailureItem{
				Id: item.Id, Name: item.Name, Url: item.Url, Error: err.Error(),
			})
			continue
		}
		out.Success++
	}

	out.Done = len(reuploadList) < size
	return out, nil
}

func (s *sOssSync) syncOne(
	ctx context.Context,
	item *entity.SysAttachment,
	targetDriver string,
	cfg *storager.StorageConfig,
	target storager.Storager,
	localDrv storager.Storager,
	deleteLocal bool,
	dryRun bool,
	isReupload bool,
) error {
	if !isReupload && !isLocalStorage(item.Storage) {
		return gerror.New("非本地存储记录，已跳过")
	}

	objectKey := storager.CloudObjectKeyFromLocal(item.Url, targetDriver, cfg)
	if objectKey == "" {
		return gerror.New("无法解析云存储 object key")
	}

	if dryRun {
		return nil
	}

	data, err := storager.ReadAttachmentBytes(ctx, item.Storage, item.Url)
	if err != nil {
		return gerror.Wrap(err, "读取附件内容失败")
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(item.Name)), ".")
	if ext == "" {
		ext = strings.TrimPrefix(strings.ToLower(filepath.Ext(item.Url)), ".")
	}

	_, err = target.Upload(ctx, &storager.UploadFile{
		Data:      data,
		Filename:  item.Name,
		Ext:       ext,
		MimeType:  item.Mimetype,
		Size:      int64(len(data)),
		ObjectKey: objectKey,
	})
	if err != nil {
		return gerror.Wrap(err, "上传到云存储失败")
	}

	storedURL := storager.StoredURLFromObjectKey(objectKey)
	updateCtx := context.WithoutCancel(ctx)
	_, err = dao.SysAttachment.Ctx(updateCtx).
		Where(dao.SysAttachment.Columns().Id, item.Id).
		Data(do.SysAttachment{
			Url:        storedURL,
			Storage:    targetDriver,
			UpdateTime: uint(utility.NowUnix()),
		}).Update()
	if err != nil {
		return gerror.Wrap(err, "更新附件记录失败")
	}

	if deleteLocal {
		_ = localDrv.Delete(ctx, item.Url)
	}
	return nil
}

func countLocalAttachments(ctx context.Context) (int, uint64, error) {
	m := localAttachmentQuery(ctx)
	total, err := m.Count()
	if err != nil {
		return 0, 0, gerror.Wrap(err, "统计本地附件失败")
	}
	var sum struct {
		TotalSize uint64 `json:"totalSize"`
	}
	if total > 0 {
		_ = localAttachmentQuery(ctx).Fields("SUM(size) as totalSize").Scan(&sum)
	}
	return total, sum.TotalSize, nil
}

func countReuploadable(ctx context.Context, driver string) (int, error) {
	var items []entity.SysAttachment
	if err := cloudAttachmentQuery(ctx, driver).Scan(&items); err != nil {
		return 0, gerror.Wrap(err, "统计可重新上传附件失败")
	}
	n := 0
	for i := range items {
		if storager.AttachmentAvailable(ctx, items[i].Storage, items[i].Url) {
			n++
		}
	}
	return n, nil
}

func fetchLocalBatch(ctx context.Context, afterId uint64, size int) ([]entity.SysAttachment, error) {
	m := localAttachmentQuery(ctx).OrderAsc(dao.SysAttachment.Columns().Id)
	if afterId > 0 {
		m = m.WhereGT(dao.SysAttachment.Columns().Id, afterId)
	}
	var list []entity.SysAttachment
	err := m.Limit(size).Scan(&list)
	return list, err
}

func fetchReuploadBatch(ctx context.Context, driver string, afterId uint64, size int) ([]entity.SysAttachment, uint64, error) {
	m := cloudAttachmentQuery(ctx, driver).OrderAsc(dao.SysAttachment.Columns().Id)
	if afterId > 0 {
		m = m.WhereGT(dao.SysAttachment.Columns().Id, afterId)
	}
	var candidates []entity.SysAttachment
	// 多取一些候选，再按是否可读过滤
	if err := m.Limit(size * 10).Scan(&candidates); err != nil {
		return nil, afterId, err
	}
	list := make([]entity.SysAttachment, 0, size)
	lastId := afterId
	for _, item := range candidates {
		if !storager.AttachmentAvailable(ctx, item.Storage, item.Url) {
			continue
		}
		list = append(list, item)
		lastId = item.Id
		if len(list) >= size {
			break
		}
	}
	return list, lastId, nil
}

func localAttachmentQuery(ctx context.Context) *gdb.Model {
	return dao.SysAttachment.Ctx(ctx).WhereIn(dao.SysAttachment.Columns().Storage, []string{"local", ""})
}

func cloudAttachmentQuery(ctx context.Context, driver string) *gdb.Model {
	return dao.SysAttachment.Ctx(ctx).Where(dao.SysAttachment.Columns().Storage, driver)
}

func isLocalStorage(storage string) bool {
	s := strings.ToLower(strings.TrimSpace(storage))
	return s == "" || s == "local"
}

func resolveSyncDriver(raw string) (string, error) {
	driver := storager.NormalizeDriverInput(raw)
	if driver == "" || driver == "local" {
		return "", gerror.New("无效的目标云存储")
	}
	return driver, nil
}
