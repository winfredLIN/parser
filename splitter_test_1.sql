DELIMITER --good

USE `fwdb`--good
-- sasa
DROP PROCEDURE IF EXISTS `_GS_GM_Check` --good
CREATE DEFINER=`root`@`%` PROCEDURE `_GS_GM_Check`(vi_uid INT,vi_pwd VARCHAR(32),vi_ip VARCHAR(100),OUT vo_level INT,OUT vo_code INT)
BEGIN
	-- JinSQ 2016-04-08 CHECK IP ADD END
    END--good

DELIMITER ;
DELIMITER //
use test //
SET @sql = 'SELECT * FROM employees WHERE salary = 2321.21';
SET @result = @sql;
PREPARE stmt FROM @result;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
//
DELIMITER ;