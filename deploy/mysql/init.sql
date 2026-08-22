-- 创建秒杀库
CREATE DATABASE IF NOT EXISTS marketing_seckill;
USE marketing_seckill;

CREATE TABLE seckill_activity (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    activity_id VARCHAR(32) NOT NULL COMMENT '活动ID',
    activity_name VARCHAR(64) NOT NULL COMMENT '活动名称',
    sku_id VARCHAR(32) NOT NULL COMMENT '商品ID',
    total_count INT NOT NULL DEFAULT 0 COMMENT '总库存',
    limit_count INT NOT NULL DEFAULT 1 COMMENT '限购数量',
    activity_state TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0初始 1开启 2关闭',
    start_time DATETIME NOT NULL COMMENT '开始时间',
    end_time DATETIME NOT NULL COMMENT '结束时间',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_activity_id (activity_id)
) COMMENT '秒杀活动表';

CREATE TABLE seckill_order (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    order_id VARCHAR(32) NOT NULL COMMENT '订单ID',
    activity_id VARCHAR(32) NOT NULL COMMENT '活动ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    sku_id VARCHAR(32) NOT NULL COMMENT '商品ID',
    order_state TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0待支付 1已支付 2已取消 3已超时',
    order_time DATETIME NOT NULL COMMENT '下单时间',
    pay_time DATETIME COMMENT '支付时间',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_order_id (order_id),
    UNIQUE KEY uk_user_activity (user_id, activity_id)
) COMMENT '秒杀订单表';

-- 创建拼团库
CREATE DATABASE IF NOT EXISTS marketing_groupbuy;
USE marketing_groupbuy;

CREATE TABLE groupbuy_activity (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    activity_id VARCHAR(32) NOT NULL COMMENT '活动ID',
    activity_name VARCHAR(64) NOT NULL COMMENT '活动名称',
    discount_id VARCHAR(32) NOT NULL COMMENT '折扣ID',
    group_type TINYINT NOT NULL DEFAULT 0 COMMENT '成团类型：0自动 1达标',
    target_count INT NOT NULL DEFAULT 2 COMMENT '成团人数',
    valid_time INT NOT NULL DEFAULT 24 COMMENT '有效时长(小时)',
    tag_id VARCHAR(32) COMMENT '人群标签ID',
    activity_state TINYINT NOT NULL DEFAULT 0 COMMENT '状态',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_activity_id (activity_id)
) COMMENT '拼团活动表';

CREATE TABLE groupbuy_discount (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    discount_id VARCHAR(32) NOT NULL COMMENT '折扣ID',
    market_plan VARCHAR(8) NOT NULL COMMENT '营销方案：ZJ直减/MJ满减/N元购/ZK折扣',
    market_expr VARCHAR(64) NOT NULL COMMENT '营销表达式',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_discount_id (discount_id)
) COMMENT '折扣配置表';

CREATE TABLE groupbuy_order (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    order_id VARCHAR(32) NOT NULL COMMENT '订单ID',
    team_id VARCHAR(32) NOT NULL COMMENT '团队ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    activity_id VARCHAR(32) NOT NULL COMMENT '活动ID',
    biz_id VARCHAR(64) NOT NULL COMMENT '幂等键',
    order_state TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0锁定 1完成 2退单',
    create_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_order_id (order_id),
    UNIQUE KEY uk_biz_id (biz_id)
) COMMENT '拼团订单表';

CREATE TABLE groupbuy_team (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    team_id VARCHAR(32) NOT NULL COMMENT '团队ID',
    activity_id VARCHAR(32) NOT NULL COMMENT '活动ID',
    target_count INT NOT NULL DEFAULT 2 COMMENT '目标人数',
    complete_count INT NOT NULL DEFAULT 0 COMMENT '完成人数',
    lock_count INT NOT NULL DEFAULT 0 COMMENT '锁定人数',
    team_state TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0拼单中 1成功 2失败',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_team_id (team_id)
) COMMENT '拼团团队表';

CREATE TABLE notify_task (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_id VARCHAR(32) NOT NULL COMMENT '任务ID',
    notify_type VARCHAR(8) NOT NULL COMMENT '通知类型：HTTP/MQ',
    notify_status TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0初始 1完成 2重试 3失败',
    notify_url VARCHAR(256) COMMENT '通知地址',
    notify_data TEXT COMMENT '通知数据',
    uuid VARCHAR(64) NOT NULL COMMENT '幂等键',
    retry_count INT NOT NULL DEFAULT 0 COMMENT '重试次数',
    max_retry INT NOT NULL DEFAULT 3 COMMENT '最大重试',
    next_time BIGINT NOT NULL DEFAULT 0 COMMENT '下次重试时间',
    create_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_task_id (task_id),
    UNIQUE KEY uk_uuid (uuid)
) COMMENT '回调通知任务表';

-- 创建抽奖库
CREATE DATABASE IF NOT EXISTS marketing_lottery;
USE marketing_lottery;

CREATE TABLE lottery_activity (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    activity_id VARCHAR(32) NOT NULL COMMENT '活动ID',
    activity_name VARCHAR(64) NOT NULL COMMENT '活动名称',
    strategy_id VARCHAR(32) NOT NULL COMMENT '策略ID',
    activity_state TINYINT NOT NULL DEFAULT 0 COMMENT '状态',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_activity_id (activity_id)
) COMMENT '抽奖活动表';

CREATE TABLE lottery_strategy (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    strategy_id VARCHAR(32) NOT NULL COMMENT '策略ID',
    rule_models VARCHAR(256) COMMENT '规则模型',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_strategy_id (strategy_id)
) COMMENT '抽奖策略表';

CREATE TABLE strategy_award (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    strategy_id VARCHAR(32) NOT NULL COMMENT '策略ID',
    award_id VARCHAR(32) NOT NULL COMMENT '奖品ID',
    award_name VARCHAR(64) NOT NULL COMMENT '奖品名称',
    award_rate DECIMAL(8,4) NOT NULL COMMENT '中奖概率',
    award_count INT NOT NULL DEFAULT 0 COMMENT '奖品数量',
    rule_models VARCHAR(256) COMMENT '规则模型',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_strategy_award (strategy_id, award_id)
) COMMENT '策略奖品表';

CREATE TABLE lottery_order (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    order_id VARCHAR(32) NOT NULL COMMENT '订单ID',
    activity_id VARCHAR(32) NOT NULL COMMENT '活动ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    award_id VARCHAR(32) COMMENT '奖品ID',
    award_state TINYINT NOT NULL DEFAULT 0 COMMENT '状态',
    award_time DATETIME COMMENT '中奖时间',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_order_id (order_id)
) COMMENT '抽奖订单表';
