#!/bin/bash
# HotGo 日志清理脚本
# 用于清理过期的日志文件，释放磁盘空间
# 使用方法: bash cleanup_logs.sh

set -e

LOG_DIR="/opt/hotgo/hotgo_v2/server/logs"
DAYS_TO_KEEP=7  # 保留7天的日志
MIN_DISK_FREE_GB=5  # 最小保留5GB磁盘空间

echo "=========================================="
echo "  HotGo 日志清理脚本"
echo "=========================================="
echo ""

# 检查日志目录是否存在
if [ ! -d "$LOG_DIR" ]; then
    echo "错误: 日志目录不存在: $LOG_DIR"
    exit 1
fi

# 显示清理前的磁盘使用情况
echo "清理前的磁盘使用情况:"
df -h $LOG_DIR | tail -1
echo ""

# 计算日志目录大小
BEFORE_SIZE=$(du -sh $LOG_DIR | cut -f1)
echo "日志目录当前大小: $BEFORE_SIZE"
echo ""

# 清理超过指定天数的日志文件
echo "正在清理超过 ${DAYS_TO_KEEP} 天的日志文件..."
find $LOG_DIR -type f -name "*.log" -mtime +$DAYS_TO_KEEP -delete
find $LOG_DIR -type f -name "*.log.*" -mtime +$DAYS_TO_KEEP -delete  # 压缩的日志文件

# 清理空的日志目录
find $LOG_DIR -type d -empty -delete

# 清理压缩的日志文件（保留最近3个）
echo "清理旧的压缩日志文件（仅保留最近3个）..."
for subdir in $(find $LOG_DIR -type d); do
    # 查找压缩的日志文件，按时间排序，删除除了最近3个之外的所有文件
    find "$subdir" -type f -name "*.log.*.gz" -printf '%T@ %p\n' | \
        sort -rn | \
        tail -n +4 | \
        cut -d' ' -f2- | \
        xargs -r rm -f
done

# 显示清理后的磁盘使用情况
AFTER_SIZE=$(du -sh $LOG_DIR | cut -f1)
echo ""
echo "清理后的磁盘使用情况:"
df -h $LOG_DIR | tail -1
echo ""
echo "日志目录清理后大小: $AFTER_SIZE"
echo ""

# 计算释放的空间
BEFORE_BYTES=$(du -sb $LOG_DIR | cut -f1)
# 这里简化处理，实际可以更精确计算
echo "✅ 日志清理完成！"
echo ""

# 检查磁盘空间
DISK_FREE=$(df -BG $LOG_DIR | tail -1 | awk '{print $4}' | sed 's/G//')
if [ "$DISK_FREE" -lt "$MIN_DISK_FREE_GB" ]; then
    echo "⚠️  警告: 磁盘剩余空间不足 ${MIN_DISK_FREE_GB}GB，当前剩余: ${DISK_FREE}GB"
    echo "建议: 进一步清理日志或增加磁盘空间"
else
    echo "✅ 磁盘空间充足: 剩余 ${DISK_FREE}GB"
fi

echo ""
echo "=========================================="
