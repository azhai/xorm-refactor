
SET NAMES utf8;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for t_access
-- ----------------------------
DROP TABLE IF EXISTS `t_access`;
CREATE TABLE `t_access` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `role_name` varchar(50) NOT NULL DEFAULT '' COMMENT '角色名',
  `resource_type` varchar(50) NOT NULL DEFAULT '' COMMENT '资源类型',
  `resource_args` varchar(255) NULL DEFAULT NULL COMMENT '资源参数',
  `perm_code` smallint(5) UNSIGNED NOT NULL DEFAULT 0 COMMENT '权限码',
  `actions` varchar(50) NOT NULL DEFAULT '' COMMENT '允许的操作',
  `granted_at` timestamp(0) NULL DEFAULT NULL COMMENT '授权时间',
  `revoked_at` timestamp(0) NULL DEFAULT NULL COMMENT '撤销时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_access_role_name`(`role_name`) USING BTREE,
  INDEX `idx_access_revoked_at`(`revoked_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '权限控制' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for t_cron_daily
-- ----------------------------
DROP TABLE IF EXISTS `t_cron_daily`;
CREATE TABLE `t_cron_daily` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `task_id` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '任务ID',
  `is_active` bit(1) NOT NULL DEFAULT b'0' COMMENT '有效',
  `workday` bit(1) NOT NULL DEFAULT b'0' COMMENT '工作日',
  `weekday` tinyint(3) UNSIGNED NOT NULL DEFAULT 0 COMMENT '周X|周Y...',
  `run_clock` char(8) NOT NULL DEFAULT '' COMMENT '具体时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_cron_daily_task_id`(`task_id`) USING BTREE,
  INDEX `idx_cron_daily_run_clock`(`run_clock`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '日常执行' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for t_cron_notice
-- ----------------------------
DROP TABLE IF EXISTS `t_cron_notice`;
CREATE TABLE `t_cron_notice` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_uid` char(16) NOT NULL DEFAULT '' COMMENT '用户ID',
  `task_id` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '任务ID',
  `is_active` bit(1) NOT NULL DEFAULT b'0' COMMENT '有效',
  `important` tinyint(3) UNSIGNED NOT NULL DEFAULT 0 COMMENT '重要程度',
  `message` text NULL COMMENT '消息内容',
  `read_time` datetime(0) NULL DEFAULT NULL COMMENT '阅读时间',
  `delay_start_time` datetime(0) NULL DEFAULT NULL COMMENT '推迟开始时间',
  `start_time` datetime(0) NULL DEFAULT NULL COMMENT '开始时间',
  `stop_time` datetime(0) NULL DEFAULT NULL COMMENT '结束时间',
  `start_clock` char(8) NULL DEFAULT NULL COMMENT '开始时刻',
  `stop_clock` char(8) NULL DEFAULT NULL COMMENT '结束时刻',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_cron_notice_task_id`(`task_id`) USING BTREE,
  INDEX `idx_cron_notice_user_uid`(`user_uid`) USING BTREE,
  INDEX `idx_cron_notice_read_time`(`read_time`) USING BTREE,
  INDEX `idx_cron_notice_delay_start_time`(`delay_start_time`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '消息提醒' ROW_FORMAT = DYNAMIC;

CREATE TABLE IF NOT EXISTS `t_cron_notice_{{CURR_MONTH}}` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_uid` char(16) NOT NULL DEFAULT '' COMMENT '用户ID',
  `task_id` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '任务ID',
  `is_active` bit(1) NOT NULL DEFAULT b'0' COMMENT '有效',
  `important` tinyint(3) UNSIGNED NOT NULL DEFAULT 0 COMMENT '重要程度',
  `message` text NULL COMMENT '消息内容',
  `read_time` datetime(0) NULL DEFAULT NULL COMMENT '阅读时间',
  `delay_start_time` datetime(0) NULL DEFAULT NULL COMMENT '推迟开始时间',
  `start_time` datetime(0) NULL DEFAULT NULL COMMENT '开始时间',
  `stop_time` datetime(0) NULL DEFAULT NULL COMMENT '结束时间',
  `start_clock` char(8) NULL DEFAULT NULL COMMENT '开始时刻',
  `stop_clock` char(8) NULL DEFAULT NULL COMMENT '结束时刻',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_cron_notice_task_id`(`task_id`) USING BTREE,
  INDEX `idx_cron_notice_user_uid`(`user_uid`) USING BTREE,
  INDEX `idx_cron_notice_read_time`(`read_time`) USING BTREE,
  INDEX `idx_cron_notice_delay_start_time`(`delay_start_time`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '消息提醒' ROW_FORMAT = DYNAMIC;

CREATE TABLE IF NOT EXISTS `t_cron_notice_{{PREV_MONTH}}` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_uid` char(16) NOT NULL DEFAULT '' COMMENT '用户ID',
  `task_id` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '任务ID',
  `is_active` bit(1) NOT NULL DEFAULT b'0' COMMENT '有效',
  `important` tinyint(3) UNSIGNED NOT NULL DEFAULT 0 COMMENT '重要程度',
  `message` text NULL COMMENT '消息内容',
  `read_time` datetime(0) NULL DEFAULT NULL COMMENT '阅读时间',
  `delay_start_time` datetime(0) NULL DEFAULT NULL COMMENT '推迟开始时间',
  `start_time` datetime(0) NULL DEFAULT NULL COMMENT '开始时间',
  `stop_time` datetime(0) NULL DEFAULT NULL COMMENT '结束时间',
  `start_clock` char(8) NULL DEFAULT NULL COMMENT '开始时刻',
  `stop_clock` char(8) NULL DEFAULT NULL COMMENT '结束时刻',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_cron_notice_task_id`(`task_id`) USING BTREE,
  INDEX `idx_cron_notice_user_uid`(`user_uid`) USING BTREE,
  INDEX `idx_cron_notice_read_time`(`read_time`) USING BTREE,
  INDEX `idx_cron_notice_delay_start_time`(`delay_start_time`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '消息提醒' ROW_FORMAT = DYNAMIC;

CREATE TABLE IF NOT EXISTS `t_cron_notice_{{EARLY_MONTH}}` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_uid` char(16) NOT NULL DEFAULT '' COMMENT '用户ID',
  `task_id` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '任务ID',
  `is_active` bit(1) NOT NULL DEFAULT b'0' COMMENT '有效',
  `important` tinyint(3) UNSIGNED NOT NULL DEFAULT 0 COMMENT '重要程度',
  `message` text NULL COMMENT '消息内容',
  `read_time` datetime(0) NULL DEFAULT NULL COMMENT '阅读时间',
  `delay_start_time` datetime(0) NULL DEFAULT NULL COMMENT '推迟开始时间',
  `start_time` datetime(0) NULL DEFAULT NULL COMMENT '开始时间',
  `stop_time` datetime(0) NULL DEFAULT NULL COMMENT '结束时间',
  `start_clock` char(8) NULL DEFAULT NULL COMMENT '开始时刻',
  `stop_clock` char(8) NULL DEFAULT NULL COMMENT '结束时刻',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_cron_notice_task_id`(`task_id`) USING BTREE,
  INDEX `idx_cron_notice_user_uid`(`user_uid`) USING BTREE,
  INDEX `idx_cron_notice_read_time`(`read_time`) USING BTREE,
  INDEX `idx_cron_notice_delay_start_time`(`delay_start_time`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '消息提醒' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for t_cron_task
-- ----------------------------
DROP TABLE IF EXISTS `t_cron_task`;
CREATE TABLE `t_cron_task` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_uid` char(16) NOT NULL DEFAULT '' COMMENT '用户ID',
  `refer_id` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联任务ID',
  `is_active` bit(1) NOT NULL DEFAULT b'0' COMMENT '有效',
  `behind` smallint(6) NOT NULL DEFAULT 0 COMMENT '相对推迟/提前多少分钟',
  `action_type` enum('command','message','http_get','http_post','function') NOT NULL DEFAULT 'command' COMMENT '动作类型',
  `cmd_url` varchar(500) NOT NULL DEFAULT '' COMMENT '指令或网址',
  `args_data` text NULL COMMENT '参数或消息体',
  `last_time` datetime(0) NULL DEFAULT NULL COMMENT '最后执行时间',
  `last_result` text NULL COMMENT '执行结果',
  `last_error` text NULL COMMENT '出错信息',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_cron_task_refer_id`(`refer_id`) USING BTREE,
  INDEX `idx_cron_task_user_uid`(`user_uid`) USING BTREE,
  INDEX `idx_cron_task_last_time`(`last_time`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '定时任务' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for t_cron_timer
-- ----------------------------
DROP TABLE IF EXISTS `t_cron_timer`;
CREATE TABLE `t_cron_timer` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `task_id` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '任务ID',
  `is_active` bit(1) NOT NULL DEFAULT b'0' COMMENT '有效',
  `run_date` date NULL DEFAULT NULL COMMENT '指定日期',
  `run_clock` char(8) NULL DEFAULT NULL COMMENT '具体时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_cron_timer_task_id`(`task_id`) USING BTREE,
  INDEX `idx_cron_timer_run_date`(`run_date`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '定时执行' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for t_group
-- ----------------------------
DROP TABLE IF EXISTS `t_group`;
CREATE TABLE `t_group` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `gid` char(16) NOT NULL DEFAULT '' COMMENT '唯一ID',
  `title` varchar(50) NOT NULL DEFAULT '' COMMENT '名称',
  `remark` text NULL COMMENT '说明备注',
  `created_at` timestamp(0) NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uix_group_gid`(`gid`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '用户组' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for t_menu
-- ----------------------------
DROP TABLE IF EXISTS `t_menu`;
CREATE TABLE `t_menu` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `lft` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '左边界',
  `rgt` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '右边界',
  `depth` tinyint(3) UNSIGNED NOT NULL DEFAULT 1 COMMENT '高度',
  `path` varchar(100) NOT NULL DEFAULT '' COMMENT '路径',
  `title` varchar(50) NOT NULL DEFAULT '' COMMENT '名称',
  `icon` varchar(30) NULL DEFAULT NULL COMMENT '图标',
  `remark` text NULL COMMENT '说明备注',
  `created_at` timestamp(0) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` timestamp(0) NULL DEFAULT NULL COMMENT '更新时间',
  `deleted_at` timestamp(0) NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_menu_rgt`(`rgt`) USING BTREE,
  INDEX `idx_menu_path`(`path`) USING BTREE,
  INDEX `idx_menu_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '菜单' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for t_role
-- ----------------------------
DROP TABLE IF EXISTS `t_role`;
CREATE TABLE `t_role` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL DEFAULT '' COMMENT '名称',
  `remark` text NULL COMMENT '说明备注',
  `created_at` timestamp(0) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` timestamp(0) NULL DEFAULT NULL COMMENT '更新时间',
  `deleted_at` timestamp(0) NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uix_role_name`(`name`) USING BTREE,
  INDEX `idx_role_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '角色' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for t_user
-- ----------------------------
DROP TABLE IF EXISTS `t_user`;
CREATE TABLE `t_user` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` char(16) NOT NULL DEFAULT '' COMMENT '唯一ID',
  `username` varchar(30) NOT NULL DEFAULT '' COMMENT '用户名',
  `password` varchar(60) NOT NULL DEFAULT '' COMMENT '密码',
  `realname` varchar(20) NULL DEFAULT NULL COMMENT '昵称/称呼',
  `mobile` varchar(20) NULL DEFAULT NULL COMMENT '手机号码',
  `email` varchar(50) NULL DEFAULT NULL COMMENT '电子邮箱',
  `prin_gid` char(16) NOT NULL DEFAULT '' COMMENT '主用户组',
  `vice_gid` char(16) NULL DEFAULT NULL COMMENT '次用户组',
  `avatar` varchar(100) NULL DEFAULT NULL COMMENT '头像',
  `introduction` varchar(500) NULL DEFAULT NULL COMMENT '介绍说明',
  `created_at` timestamp(0) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` timestamp(0) NULL DEFAULT NULL COMMENT '更新时间',
  `deleted_at` timestamp(0) NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uix_user_uid`(`uid`) USING BTREE,
  INDEX `idx_user_username`(`username`) USING BTREE,
  INDEX `idx_user_mobile`(`mobile`) USING BTREE,
  INDEX `idx_user_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '用户' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for t_user_role
-- ----------------------------
DROP TABLE IF EXISTS `t_user_role`;
CREATE TABLE `t_user_role` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_uid` char(16) NOT NULL DEFAULT '' COMMENT '用户ID',
  `role_name` varchar(50) NOT NULL DEFAULT '' COMMENT '角色名',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_user_role_user_uid`(`user_uid`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '用户角色' ROW_FORMAT = DYNAMIC;

SET FOREIGN_KEY_CHECKS = 1;
