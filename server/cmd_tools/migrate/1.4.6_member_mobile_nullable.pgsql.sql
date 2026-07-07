-- ============================================================
-- v1.4.6 会员手机号允许为空(NULL) —— PostgreSQL 版
-- 背景：mobile 原为 NOT NULL DEFAULT ''，且 PG 侧缺失唯一索引。
-- 方案：改为可空、去掉默认空串，历史空串归一为 NULL，
--       补齐手机号唯一索引（PG 唯一索引允许多个 NULL）。
-- ============================================================

-- 1) 放宽为可空并去掉默认空串
ALTER TABLE xy_member ALTER COLUMN mobile DROP NOT NULL;
ALTER TABLE xy_member ALTER COLUMN mobile DROP DEFAULT;

-- 2) 历史空字符串归一为 NULL
UPDATE xy_member SET mobile = NULL WHERE mobile = '';

-- 3) 补齐手机号唯一索引（幂等）
CREATE UNIQUE INDEX IF NOT EXISTS uk_member_mobile ON xy_member (mobile);
