CREATE TABLE IF NOT EXISTS `mc_article` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `sn` varchar(32) NOT NULL DEFAULT '' COMMENT '文章序号，程序生成',
  `title` varchar(127) NOT NULL DEFAULT '' COMMENT '文章标题',
  `tags` varchar(127) DEFAULT '' COMMENT '文章 tag，逗号分隔',
  `img` varchar(255) NOT NULL DEFAULT '' COMMENT '文章封面图',
  `content` longtext NOT NULL COMMENT '内容，markdown 格式',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '状态：0-草稿;1-已上线;2-下线',
  `ord` int(3)  DEFAULT 0 COMMENT '排序权重，越大越靠前',
  `deleted_at` timestamp DEFAULT NULL COMMENT '删除时间，软删除',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `sn` (`sn`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章表';


CREATE TABLE IF NOT EXISTS `mc_metas` (
  `article_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '文章ID',
  `sn`  varchar(32) NOT NULL DEFAULT '' COMMENT '冗余字段',
  `view_count` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '浏览数',
  `cm_count` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '评论数',
  `laud_count` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '赞数',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT article_id FOREIGN KEY (article_id) references mc_article(id) on DELETE CASCADE on update CASCADE,
  PRIMARY KEY (`article_id`),
  UNIQUE KEY `sn` (`sn`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章扩展表';


CREATE TABLE IF NOT EXISTS `mc_tag`(
    `id` mediumint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `name` varchar(30) NOT NULL,
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` timestamp DEFAULT NULL COMMENT '删除时间，软删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `name` (`name`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章tags表';


CREATE TABLE IF NOT EXISTS `mc_article_tag`(
    `article_id` int unsigned NOT NULL,
    `tag_id` int unsigned NOT NULL,
    `tag_name` varchar (30) NOT NULL COMMENT '冗余字段',
    CONSTRAINT article_tag_id  FOREIGN KEY (article_id) references mc_article(id) on DELETE CASCADE on update CASCADE,
     `deleted_at` timestamp DEFAULT NULL COMMENT '删除时间，软删除',
     KEY (`article_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '文章-tag关联表';

/** 信息（配置）表：包括音乐、背景图、友链等都存这里面，通过webkey来找出对应的配置，webkey可以重复，比如webkey为MUSIC的代表音乐**/
CREATE TABLE IF NOT EXISTS `mc_web_info`(
     `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
     `sn`  varchar(32) NOT NULL COMMENT 'sn',
     `web_key`  varchar(60) NOT NULL COMMENT 'key，配置名称，可重复，但是key+value不能重复',
     `unique_val`  varchar(150) NOT NULL COMMENT '值的唯一标识，string格式，用于检测该key下的值是否重复，比如音乐的url就可以作为一个uniqueval',
     `web_val`  varchar(2000) NOT NULL COMMENT 'value的完整值，json格式',
     `status` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '状态：0-下线;1-已上线',
     `ord`     varchar(20)  NOT NULL default "0" COMMENT '权重排序(asc序)，越大越靠前',
     `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
     `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
     `deleted_at` timestamp DEFAULT NULL COMMENT '删除时间，软删除',
     PRIMARY KEY `id` (id),
     UNIQUE KEY `sn` (sn),
     UNIQUE KEY `key-val` (web_key,unique_val) -- 同一个key下禁止重复配置
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '网站配置表';


CREATE TABLE IF NOT EXISTS `mc_user`(
    `id` int  unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `sn` varchar(32) NOT NULL COMMENT 'SN',
    `username` varchar(20)  COMMENT '用户名（登录名，授权登陆的无用户名）',
    `password` varchar(500) COMMENT '密码（授权登陆的无密码）',
    `password_token` varchar(500) NOT NULL COMMENT '密码对应令牌，每次修改密码都会重新生成',
    `nickname` varchar (20) NOT NULL COMMENT '用户昵称',
    `avatar` varchar(150) COMMENT '头像',
    `desc` varchar(150) COMMENT '简介',
    `email` varchar (20) COMMENT '邮箱',
    `phone` varchar (20) COMMENT '手机号',
    `cate` varchar(20) NOT NULL DEFAULT 'BACK' COMMENT '注册类型,BACK（后台添加）,OPEN（第三方登录）,EMAIL（邮箱注册），PHONE(手机）',
    `open_cate` varchar(50) COMMENT '第三方登录类型',
    `openid` varchar(100) COMMENT '第三方OPENID',
    `openinfo` varchar(5000) COMMENT '第三方OPEN信息',
    `status` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '用户状态：0-正常;1-禁用',
    `is_super` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '是否为管理员，0-否;1-是',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` timestamp DEFAULT NULL COMMENT '删除时间，软删除',
    PRIMARY KEY `id` (id),
    UNIQUE KEY `sn` (sn),
    UNIQUE KEY `username` (username)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '用户表';


INSERT INTO `mc_user` values(DEFAULT,"deedkuhlzxagysfdyks","admin",
"$2a$10$cE2unOi2AzfToRfX8SOVkOLRsC0n1uQCmusZYdwW/PKqtwKa6kNy6",
"jeyJpdiI6Inl1WSsyMld5d09yMW1QdDlrdlFIOVE9PSIsInZhbHVXSmZORVVodmZqUTJvOYyI6ImVlNzI4NDU0NzE0OT4M2YwY2MifQ",
"initAdmin","avatar","desc","email","phone",DEFAULT,"","","",0,1,DEFAULT,DEFAULT,DEFAULT);


CREATE TABLE IF NOT EXISTS `mc_upload` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `name` varchar(255) NOT NULL COMMENT '文件名',
    `url` varchar(255) NOT NULL COMMENT '文件路径',
    `tag` varchar(255) NOT NULL COMMENT '文件后缀tag',
    `uploadkey` varchar(255) NOT NULL COMMENT '文件key（外链host后面的那段）',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` timestamp DEFAULT NULL COMMENT '删除时间，软删除',
    PRIMARY KEY `id` (id),
    UNIQUE KEY `uploadkey` (uploadkey)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '文件上传表';

