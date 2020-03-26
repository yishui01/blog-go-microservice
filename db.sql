
CREATE DATABASE IF NOT EXISTS `micro_blog` DEFAULT CHARACTER SET utf8mb4;

USE `blog`;

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
