SET FOREIGN_KEY_CHECKS = 0;
SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

-- --------------------------------------------------------
-- 1. Table structure for table `sources`
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `sources` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `domain` VARCHAR(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_domain` (`domain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- 2. Table structure for table `series`
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `series` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `source_id` INT UNSIGNED NOT NULL,
  `slug` VARCHAR(255) NOT NULL,
  `title` VARCHAR(500) NOT NULL,
  `alternate_titles` TEXT,
  `description` TEXT,
  `thumbnail_url` VARCHAR(1024) DEFAULT NULL,
  `author` VARCHAR(255) DEFAULT NULL,
  `genre` TEXT,
  `status` VARCHAR(50) DEFAULT NULL,
  `release_year` YEAR DEFAULT NULL,
  `created_at` DATETIME DEFAULT NULL,
  `updated_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_slug` (`slug`),
  KEY `idx_source_id` (`source_id`),
  -- Custom Indexes
  KEY `idx_status_source` (`status`, `source_id`),
  KEY `idx_title` (`title`(255)),
  KEY `idx_author` (`author`),
  KEY `idx_updated_at` (`updated_at` DESC),
  KEY `idx_release_year` (`release_year`),
  KEY `idx_genre_status` (`genre`(100), `status`),
  KEY `idx_status_updated` (`status`, `updated_at` DESC),
  KEY `idx_source_updated` (`source_id`, `updated_at` DESC),
  KEY `idx_year_status` (`release_year`, `status`),
  FULLTEXT KEY `idx_search` (`title`, `alternate_titles`, `description`),
  CONSTRAINT `fk_series_source` FOREIGN KEY (`source_id`) REFERENCES `sources` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------
-- 3. Table structure for table `chapters`
-- --------------------------------------------------------
CREATE TABLE IF NOT EXISTS `chapters` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `series_id` BIGINT UNSIGNED NOT NULL,
  `chapter_number` DECIMAL(10,2) NOT NULL,
  
  -- Novel Specific Columns
  `title` VARCHAR(255) DEFAULT NULL COMMENT 'Optional chapter title',
  `content` LONGTEXT COMMENT 'The actual text body of the chapter',
  
  `created_at` DATETIME DEFAULT NULL,
  `updated_at` DATETIME DEFAULT NULL,
  
  PRIMARY KEY (`id`),
  KEY `idx_series_id` (`series_id`),
  
  -- Unique index: A series cannot have two Chapter 1.0s
  UNIQUE KEY `idx_series_chapter` (`series_id`, `chapter_number`),
  
  -- Helpful indexes for sorting
  KEY `idx_chapter_number` (`chapter_number`),
  KEY `idx_created_at` (`created_at` DESC),
  
  CONSTRAINT `fk_chapters_series` FOREIGN KEY (`series_id`) REFERENCES `series` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET FOREIGN_KEY_CHECKS = 1;