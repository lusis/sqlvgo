CREATE TABLE `testdata` (
`counter` INT UNSIGNED NOT NULL AUTO_INCREMENT,
`id` VARCHAR(191) NOT NULL,
`name` VARCHAR(191) NOT NULL,
`rtype` TINYINT UNSIGNED NOT NULL,
`rstate` TINYINT UNSIGNED NOT NULL,
`created_at` DATETIME(6) NOT NULL DEFAULT NOW(6),
`updated_at` DATETIME(6) NOT NULL DEFAULT NOW(6) ON UPDATE NOW(6),
PRIMARY KEY (`counter`),
UNIQUE INDEX `id` (`id`),
INDEX `state` (`rstate`),
INDEX `type` (`rtype`),
INDEX `state_and_type` (`rstate`,`rtype`)
)
CHARSET='utf8mb4'
ENGINE=InnoDB;