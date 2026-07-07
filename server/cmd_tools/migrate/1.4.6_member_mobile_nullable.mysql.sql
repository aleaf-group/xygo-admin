-- ============================================================
-- v1.4.6 会员手机号允许为空(NULL) —— MySQL 版
-- 背景：mobile 原为 NOT NULL DEFAULT '' 且有唯一索引 uk_mobile，
--       多个无手机号会员写入空字符串会触发 uk_mobile 唯一键冲突。
-- 方案：改为可空，空值统一存 NULL（MySQL 唯一索引允许多个 NULL）。
-- ============================================================

-- 1) 列放宽为可空
ALTER TABLE `xy_member` MODIFY COLUMN `mobile` varchar(20) DEFAULT NULL COMMENT '手机号';

-- 2) 历史空字符串归一为 NULL，消除已有冲突数据
UPDATE `xy_member` SET `mobile` = NULL WHERE `mobile` = '';

-- 唯一索引 uk_mobile 保持不变（NULL 不参与唯一约束，无需重建）
