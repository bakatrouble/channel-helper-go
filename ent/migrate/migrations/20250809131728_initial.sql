-- Create "posts" table
CREATE TABLE `posts`
(
    `id`         uuid     NOT NULL,
    `type`       text     NOT NULL,
    `file_id`    text     NOT NULL,
    `is_sent`    bool     NOT NULL DEFAULT false,
    `created_at` datetime NOT NULL,
    `sent_at`    datetime NULL,
    `image_hash` text NULL,
    PRIMARY KEY (`id`)
);
-- Create index "post_image_hash" to table: "posts"
CREATE INDEX `post_image_hash` ON `posts` (`image_hash`);
-- Create index "post_is_sent" to table: "posts"
CREATE INDEX `post_is_sent` ON `posts` (`is_sent`);
-- Create index "post_type" to table: "posts"
CREATE INDEX `post_type` ON `posts` (`type`);
-- Create "post_message_ids" table
CREATE TABLE `post_message_ids`
(
    `id`               integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    `chat_id`          integer NOT NULL,
    `message_id`       integer NOT NULL,
    `post_message_ids` uuid NULL,
    CONSTRAINT `post_message_ids_posts_message_ids` FOREIGN KEY (`post_message_ids`) REFERENCES `posts` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "postmessageid_chat_id_message_id" to table: "post_message_ids"
CREATE UNIQUE INDEX `postmessageid_chat_id_message_id` ON `post_message_ids` (`chat_id`, `message_id`);
-- Create "upload_tasks" table
CREATE TABLE `upload_tasks`
(
    `id`           uuid     NOT NULL,
    `type`         text     NOT NULL,
    `data`         blob NULL,
    `is_processed` bool     NOT NULL DEFAULT false,
    `created_at`   datetime NOT NULL,
    `sent_at`      datetime NULL,
    `image_hash`   text NULL,
    PRIMARY KEY (`id`)
);
-- Create index "uploadtask_image_hash" to table: "upload_tasks"
CREATE INDEX `uploadtask_image_hash` ON `upload_tasks` (`image_hash`);
-- Create index "uploadtask_is_processed" to table: "upload_tasks"
CREATE INDEX `uploadtask_is_processed` ON `upload_tasks` (`is_processed`);
-- Create index "uploadtask_type" to table: "upload_tasks"
CREATE INDEX `uploadtask_type` ON `upload_tasks` (`type`);
