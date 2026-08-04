-- 消息队列 Topic 运行配置（Worker 数 / 重试策略）

CREATE TABLE IF NOT EXISTS `xy_sys_queue_config` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `topic` varchar(64) NOT NULL DEFAULT '' COMMENT 'Topic 标识（唯一）',
  `title` varchar(128) NOT NULL DEFAULT '' COMMENT '显示名称',
  `workers` int(11) NOT NULL DEFAULT 1 COMMENT '并行 Worker 数',
  `max_retry` int(11) NOT NULL DEFAULT 3 COMMENT '最大重试次数',
  `retry_delay_sec` int(11) NOT NULL DEFAULT 0 COMMENT '重试间隔（秒，0=立即重试）',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '状态:0禁用,1启用',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `sort` int(11) NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
  `updated_at` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_topic` (`topic`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='消息队列Topic配置';

INSERT INTO `xy_sys_queue_config` (`topic`, `title`, `workers`, `max_retry`, `retry_delay_sec`, `status`, `remark`, `sort`, `created_at`, `updated_at`) VALUES
('login_log', '登录日志', 1, 3, 0, 1, '内置消费者', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('notice_push', '通知推送', 1, 3, 0, 1, '内置消费者', 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('operation_log', '操作日志', 1, 3, 0, 1, '内置消费者', 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('demo_task', '演示任务', 1, 3, 0, 1, '示例消费者', 4, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

UPDATE `xy_admin_menu` SET `perms` = '["GET /admin/queue/stats","GET /admin/queue/topics"]' WHERE `id` = 250;

INSERT INTO `xy_admin_menu` (`id`, `parent_id`, `type`, `title`, `name`, `path`, `component`, `icon`, `is_frame`, `is_cache`, `redirect`, `query`, `perms`, `always_show`, `breadcrumb`, `affix`, `active_menu`, `redirect_name`, `hidden`, `sort`, `status`, `remark`, `created_by`, `updated_by`, `created_at`, `updated_at`) VALUES
(812, 250, 3, '查看', 'view', '', '', '', 0, 0, '', '', '["GET /admin/queue/stats","GET /admin/queue/topics"]', 0, 0, 0, '', '', 0, 0, 0, 1, '', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(813, 250, 3, '编辑配置', 'edit', '', '', '', 0, 0, '', '', '["POST /admin/queue/configSave"]', 0, 0, 0, '', '', 0, 0, 0, 1, '', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `perms` = VALUES(`perms`), `updated_at` = UNIX_TIMESTAMP();
