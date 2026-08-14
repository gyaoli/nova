--
-- 账号数据
--
create table if not exists account (
    `acct_name` varchar(128) not null comment '帐号名',
    `password` varchar(128) not null comment '密码',
    `account` varchar(128) not null comment '玩家帐号ID，相当于UID',
    `platform` varchar(32) not null comment '注册时用的平台标识',
    `client_type` int(11) not null comment '客户端类型',
    `device_id` varchar(128) not null default '' comment '设备号',
    `device_name` varchar(128) not null default '' comment '设备名',
    `reg_ip` char(16) DEFAULT NULL COMMENT '注册IP',
    `reg_time` int(11) not null default '0' comment '注册时间',
    primary key (acct_name),
    unique acc (account),
    KEY `reg_time` (`reg_time`)
    ) engine=innodb default CHARSET=utf8 COLLATE=utf8_bin comment='帐号数据';


--
-- 角色数据（只保存简略信息）
--
create table if not exists role (
    `id` int(11) not null auto_increment,
    `platform` varchar(32) not null comment '平台标识',
    `zone_id` int(11) not null default '0' comment '区号',
    `account` varchar(128) not null comment '玩家帐号ID，相当于UID',
    `name` varchar(20) not null default '' comment '角色名',
    `lev` tinyint(4) not null default '0' comment '等级',
    `reg_ip` char(16) DEFAULT NULL COMMENT '注册IP',
    `reg_time` int(11) NOT NULL DEFAULT '0' COMMENT '注册时间',
    `data` mediumblob NOT NULL COMMENT '序列化后的角色数据',
    `last_update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
    `last_login_time` int(11) not null default '0' COMMENT '最近一次登录时间',
    `last_logout_time` int(11) not null default '0' COMMENT '最近一次登出时间',
    primary key (id, platform, zone_id),
    unique `acc` (account, platform, zone_id),
    unique `name` (`name`),
    KEY `reg_time` (`reg_time`),
    KEY `last_login_time` (`last_login_time`),
    KEY `last_logout_time` (`last_logout_time`)
    ) engine=innodb default CHARSET=utf8 COLLATE=utf8_bin comment='角色数据';
