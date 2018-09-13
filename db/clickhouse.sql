CREATE DATABASE homer;

CREATE TABLE `metrics` (
	`date`    Date DEFAULT toDate(ctime),
	`ctime`   DateTime,
	`street-n:temp` Nullable(Float32),
	`balcon:temp`   Nullable(Float32),
	`balcon:humd`   Nullable(Float32),
	`balcon:pres`   Nullable(Float32),
	`bedroom:temp`  Nullable(Float32),
	`bedroom:humd`  Nullable(Float32)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(`date`)
ORDER BY (`ctime`);


--	`year`    UInt16,
--	`month`   UInt8,
--	`day`     UInt8,
--	`weekday` UInt8,
--	`hour`    UInt8,
--	`minute`  UInt8,

-- SELECT `ctime`,  `bedroom:humd` FROM metrics WHERE `ctime` > now() - 86400 ORDER BY `ctime` LIMIT 0,5000
