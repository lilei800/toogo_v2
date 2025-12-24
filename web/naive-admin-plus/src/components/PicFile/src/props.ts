import type { ExtractPublicPropTypes, PropType } from 'vue';

// 可按需调整 通常用不了那么多类型
type FileType =
  | 'jpg' // 常用的图片格式
  | 'png' // 常用的图片格式
  | 'gif' // 动画图片格式
  | 'jpeg' // 有损压缩图片格式
  | 'bmp' // 位图格式
  | 'tiff' // 高质量图片格式
  | 'webp' // 现代高效的图片格式
  | 'svg' // 矢量图格式
  | 'txt' // 文本文件
  | 'docx' // Microsoft Word 文档
  | 'doc' // 旧版 Microsoft Word 文档
  | 'pdf' // 便携式文档格式
  | 'xlsx' // Microsoft Excel 工作簿
  | 'xls' // 旧版 Microsoft Excel 工作簿
  | 'pptx' // Microsoft PowerPoint 幻灯片
  | 'ppt' // 旧版 Microsoft PowerPoint 幻灯片
  | 'mp3' // 常用音频格式
  | 'wav' // WAV 音频格式
  | 'aac' // 高效音频格式
  | 'flac' // 无损音频格式
  | 'ogg' // 开源音频格式
  | 'mp4' // 常用视频格式
  | 'avi' // 音频视频交错格式
  | 'mkv' // Matroska 视频容器
  | 'mov' // Apple QuickTime 视频
  | 'wmv' // Windows Media Video
  | 'zip' // 常用压缩格式
  | 'rar' // RAR 压缩格式
  | '7z' // 7-Zip 压缩格式
  | 'gz' // GZIP 压缩格式
  | 'tar' // Unix 压缩归档格式
  | 'bz2' // Bzip2 压缩格式
  | 'csv' // 逗号分隔值
  | 'bmp' // Windows 位图格式
  | 'tiff' // 标记图像文件格式
  | 'webp' // 现代图片格式
  | 'svg' // 可缩放矢量图
  | 'txt' // 纯文本文件
  | 'docx' // Microsoft Word 文档
  | 'doc' // Microsoft Word 文档
  | 'pdf' // 便携式文档格式
  | 'xls' // Microsoft Excel 工作表
  | 'xlsx' // Microsoft Excel 工作表
  | 'ppt' // Microsoft PowerPoint 幻灯片
  | 'pptx' // Microsoft PowerPoint 幻灯片
  | 'mp3' // MPEG 音频层 3
  | 'wav' // WAV 音频格式
  | 'aac' // 高级音频编码
  | 'flac' // 无损音频编码
  | 'ogg' // Ogg Vorbis 音频
  | 'mp4' // MPEG-4 Part 14 视频
  | 'avi' // Audio Video Interleave
  | 'mkv' // Matroska Video
  | 'mov' // Apple QuickTime
  | 'wmv' // Windows Media Video
  | 'webm' // WebM 视频
  | 'm4v' // MPEG-4 视频
  | 'ogv' // Ogg 视频
  | 'zip' // ZIP 压缩格式
  | 'rar'; // RAR 压缩格式

export type IdKeys = string[] | number[];

export interface PicFileItem {
  id: string | number;
  name: string;
  url: string;
  type: FileType;
  size: string;
  createDate: string;
}

export interface GroupItem {
  id: string | number;
  name: string;
  quantity: number;
}

export const picFileProps = {
  // 文件列表
  fileList: {
    type: Array as PropType<PicFileItem[]>,
    default: [],
  },
  // 分组列表
  groupList: {
    type: Array as PropType<GroupItem[]>,
    default: [],
  },
  // 显示分组
  showGroup: {
    type: Boolean,
    default: true,
  },
  // 提示
  tips: {
    type: String,
    default: '支持常见扩展名，单个文件不能超过50M',
  },
  loading: {
    type: Boolean,
    default: false,
  },
};

export type PicFilesProps = ExtractPublicPropTypes<typeof picFileProps>;
