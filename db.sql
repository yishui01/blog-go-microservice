
CREATE DATABASE IF NOT EXISTS `micro_blog` DEFAULT CHARACTER SET utf8mb4;

USE `micro_blog`;


CREATE TABLE IF NOT EXISTS `mc_article` (
  `aid` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `sn` varchar(32) NOT NULL DEFAULT '' COMMENT '文章序号，程序生成',
  `title` varchar(127) NOT NULL DEFAULT '' COMMENT '文章标题',
  `img` varchar(255) NOT NULL DEFAULT '' COMMENT '文章封面图',
  `content` longtext NOT NULL COMMENT '内容，markdown 格式',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '状态：0-草稿;1-已上线;2-下线',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间，软删除',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`aid`),
  UNIQUE KEY `sn` (`sn`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章表';