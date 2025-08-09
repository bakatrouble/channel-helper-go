-- Disable the enforcement of foreign-keys constraints
PRAGMA
foreign_keys = off;

-- Create "image_hashes" table
CREATE TABLE `image_hashes`
(
    `id`                     integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    `image_hash`             text    NOT NULL,
    `post_image_hash`        uuid NULL,
    `upload_task_image_hash` uuid NULL,
    CONSTRAINT `image_hashes_upload_tasks_image_hash` FOREIGN KEY (`upload_task_image_hash`) REFERENCES `upload_tasks` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
    CONSTRAINT `image_hashes_posts_image_hash` FOREIGN KEY (`post_image_hash`) REFERENCES `posts` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "image_hashes_post_image_hash_key" to table: "image_hashes"
CREATE UNIQUE INDEX `image_hashes_post_image_hash_key` ON `image_hashes` (`post_image_hash`);
-- Create index "image_hashes_upload_task_image_hash_key" to table: "image_hashes"
CREATE UNIQUE INDEX `image_hashes_upload_task_image_hash_key` ON `image_hashes` (`upload_task_image_hash`);

-- Migrate image hashes from "posts" and "upload_tasks" to "image_hashes"
INSERT INTO `image_hashes`
    SELECT NULL, `image_hash`, `id`, NULL
    FROM `posts`
    WHERE `image_hash` IS NOT NULL;
INSERT INTO `image_hashes`
    SELECT NULL, `image_hash`, NULL, `id`
    FROM `upload_tasks`
    WHERE `image_hash` IS NOT NULL;

-- Create "new_posts" table
CREATE TABLE `new_posts`
(
    `id`         uuid     NOT NULL,
    `type`       text     NOT NULL,
    `file_id`    text     NOT NULL,
    `is_sent`    bool     NOT NULL DEFAULT false,
    `created_at` datetime NOT NULL,
    `sent_at`    datetime NULL,
    PRIMARY KEY (`id`)
);
-- Copy rows from old table "posts" to new temporary table "new_posts"
INSERT INTO `new_posts` (`id`, `type`, `file_id`, `is_sent`, `created_at`, `sent_at`)
SELECT `id`, `type`, `file_id`, `is_sent`, `created_at`, `sent_at`
FROM `posts`;
-- Drop "posts" table after copying rows
DROP TABLE `posts`;
-- Rename temporary table "new_posts" to "posts"
ALTER TABLE `new_posts` RENAME TO `posts`;
-- Create index "post_is_sent" to table: "posts"
CREATE INDEX `post_is_sent` ON `posts` (`is_sent`);
-- Create index "post_type" to table: "posts"
CREATE INDEX `post_type` ON `posts` (`type`);
-- Create "new_upload_tasks" table
CREATE TABLE `new_upload_tasks`
(
    `id`           uuid     NOT NULL,
    `type`         text     NOT NULL,
    `data`         blob NULL,
    `is_processed` bool     NOT NULL DEFAULT false,
    `created_at`   datetime NOT NULL,
    `sent_at`      datetime NULL,
    PRIMARY KEY (`id`)
);
-- Copy rows from old table "upload_tasks" to new temporary table "new_upload_tasks"
INSERT INTO `new_upload_tasks` (`id`, `type`, `data`, `is_processed`, `created_at`, `sent_at`)
SELECT `id`, `type`, `data`, `is_processed`, `created_at`, `sent_at`
FROM `upload_tasks`;
-- Drop "upload_tasks" table after copying rows
DROP TABLE `upload_tasks`;
-- Rename temporary table "new_upload_tasks" to "upload_tasks"
ALTER TABLE `new_upload_tasks` RENAME TO `upload_tasks`;
-- Create index "uploadtask_is_processed" to table: "upload_tasks"
CREATE INDEX `uploadtask_is_processed` ON `upload_tasks` (`is_processed`);
-- Create index "uploadtask_type" to table: "upload_tasks"
CREATE INDEX `uploadtask_type` ON `upload_tasks` (`type`);

-- Enable back the enforcement of foreign-keys constraints
PRAGMA
foreign_keys = on;
