-- 消息队列 Topic 运行配置（Worker 数 / 重试策略）

CREATE TABLE IF NOT EXISTS public.xy_sys_queue_config (
  id BIGSERIAL PRIMARY KEY,
  topic VARCHAR(64) NOT NULL DEFAULT '',
  title VARCHAR(128) NOT NULL DEFAULT '',
  workers INT NOT NULL DEFAULT 1,
  max_retry INT NOT NULL DEFAULT 3,
  retry_delay_sec INT NOT NULL DEFAULT 0,
  status SMALLINT NOT NULL DEFAULT 1,
  remark VARCHAR(255) NOT NULL DEFAULT '',
  sort INT NOT NULL DEFAULT 0,
  created_at BIGINT NOT NULL DEFAULT 0,
  updated_at BIGINT NOT NULL DEFAULT 0
);
COMMENT ON TABLE public.xy_sys_queue_config IS '消息队列Topic配置';
CREATE UNIQUE INDEX IF NOT EXISTS uk_xy_sys_queue_config_topic ON public.xy_sys_queue_config (topic);

INSERT INTO public.xy_sys_queue_config (topic, title, workers, max_retry, retry_delay_sec, status, remark, sort, created_at, updated_at) VALUES
('login_log', '登录日志', 1, 3, 0, 1, '内置消费者', 1, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint),
('notice_push', '通知推送', 1, 3, 0, 1, '内置消费者', 2, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint),
('operation_log', '操作日志', 1, 3, 0, 1, '内置消费者', 3, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint),
('demo_task', '演示任务', 1, 3, 0, 1, '示例消费者', 4, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint)
ON CONFLICT (topic) DO UPDATE SET updated_at = EXTRACT(EPOCH FROM NOW())::bigint;

UPDATE public.xy_admin_menu SET perms = '["GET /admin/queue/stats","GET /admin/queue/topics"]' WHERE id = 250;

INSERT INTO public.xy_admin_menu (id, parent_id, type, title, name, path, component, icon, is_frame, is_cache, redirect, query, perms, always_show, breadcrumb, affix, active_menu, redirect_name, hidden, sort, status, remark, created_by, updated_by, created_at, updated_at) VALUES
(812, 250, 3, '查看', 'view', '', '', '', 0, 0, '', '', '["GET /admin/queue/stats","GET /admin/queue/topics"]', 0, 0, 0, '', '', 0, 0, 0, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint),
(813, 250, 3, '编辑配置', 'edit', '', '', '', 0, 0, '', '', '["POST /admin/queue/configSave"]', 0, 0, 0, '', '', 0, 0, 0, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint)
ON CONFLICT (id) DO UPDATE SET perms = EXCLUDED.perms, updated_at = EXTRACT(EPOCH FROM NOW())::bigint;
