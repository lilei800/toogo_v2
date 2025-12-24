@echo off
chcp 65001 >nul
mysql -h 127.0.0.1 -P 3306 -u root -proot hotgo --default-character-set=utf8mb4 < fix_chinese_encoding.sql

