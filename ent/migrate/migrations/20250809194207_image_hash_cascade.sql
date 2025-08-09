-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_post_message_ids" table
CREATE TABLE `new_post_message_ids` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `chat_id` integer NOT NULL,
  `message_id` integer NOT NULL,
  `post_message_ids` uuid NULL,
  CONSTRAINT `post_message_ids_posts_message_ids` FOREIGN KEY (`post_message_ids`) REFERENCES `posts` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Copy rows from old table "post_message_ids" to new temporary table "new_post_message_ids"
INSERT INTO `new_post_message_ids` (`id`, `chat_id`, `message_id`, `post_message_ids`) SELECT `id`, `chat_id`, `message_id`, `post_message_ids` FROM `post_message_ids`;
-- Drop "post_message_ids" table after copying rows
DROP TABLE `post_message_ids`;
-- Rename temporary table "new_post_message_ids" to "post_message_ids"
ALTER TABLE `new_post_message_ids` RENAME TO `post_message_ids`;
-- Create index "postmessageid_chat_id_message_id" to table: "post_message_ids"
CREATE UNIQUE INDEX `postmessageid_chat_id_message_id` ON `post_message_ids` (`chat_id`, `message_id`);
-- Create "new_image_hashes" table
CREATE TABLE `new_image_hashes` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `image_hash` text NOT NULL,
  `post_image_hash` uuid NULL,
  `upload_task_image_hash` uuid NULL,
  CONSTRAINT `image_hashes_upload_tasks_image_hash` FOREIGN KEY (`upload_task_image_hash`) REFERENCES `upload_tasks` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `image_hashes_posts_image_hash` FOREIGN KEY (`post_image_hash`) REFERENCES `posts` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Copy rows from old table "image_hashes" to new temporary table "new_image_hashes"
INSERT INTO `new_image_hashes` (`id`, `image_hash`, `post_image_hash`, `upload_task_image_hash`) SELECT `id`, `image_hash`, `post_image_hash`, `upload_task_image_hash` FROM `image_hashes`;
-- Drop "image_hashes" table after copying rows
DROP TABLE `image_hashes`;
-- Rename temporary table "new_image_hashes" to "image_hashes"
ALTER TABLE `new_image_hashes` RENAME TO `image_hashes`;
-- Create index "image_hashes_post_image_hash_key" to table: "image_hashes"
CREATE UNIQUE INDEX `image_hashes_post_image_hash_key` ON `image_hashes` (`post_image_hash`);
-- Create index "image_hashes_upload_task_image_hash_key" to table: "image_hashes"
CREATE UNIQUE INDEX `image_hashes_upload_task_image_hash_key` ON `image_hashes` (`upload_task_image_hash`);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
