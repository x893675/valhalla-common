/*
Copyright 2024 x893675.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logger

// Options 日志配置选项
type Options struct {
	// Level 日志级别: debug, info, warn, error
	Level string `json:"level" yaml:"level" toml:"level"`
	// Format 输出格式: console, json
	Format string `json:"format" yaml:"format" toml:"format"`
	// Output 输出目标: stdout（仅标准输出）或文件路径（标准输出+文件，如 /var/log/app.log）
	Output string `json:"output" yaml:"output" toml:"output"`
	// File 文件轮转配置（仅当 Output 为文件路径时有效）
	File *FileOptions `json:"file,omitempty" yaml:"file,omitempty" toml:"file,omitempty"`
}

// FileOptions 日志文件轮转配置
type FileOptions struct {
	// MaxSizeMB 单个日志文件最大大小（MB）
	MaxSizeMB int `json:"maxSizeMB" yaml:"maxSizeMB" toml:"maxSizeMB"`
	// MaxBackups 最大备份文件数量
	MaxBackups int `json:"maxBackups" yaml:"maxBackups" toml:"maxBackups"`
	// MaxAgeDays 日志文件最大保留天数
	MaxAgeDays int `json:"maxAgeDays" yaml:"maxAgeDays" toml:"maxAgeDays"`
	// Compress 是否压缩归档的日志文件
	Compress bool `json:"compress" yaml:"compress" toml:"compress"`
}

// NewLogOptions 创建默认日志配置
func NewLogOptions() *Options {
	return &Options{
		Level:  "info",
		Format: "console",
		Output: "stdout",
		File: &FileOptions{
			MaxSizeMB:  100,
			MaxBackups: 5,
			MaxAgeDays: 30,
			Compress:   false,
		},
	}
}

// IsFile 判断是否配置了文件输出
func (o *Options) IsFile() bool {
	return o.Output != "" && o.Output != "stdout"
}

// GetFileOptions 获取文件配置（带默认值）
func (o *Options) GetFileOptions() *FileOptions {
	if o.File == nil {
		return &FileOptions{
			MaxSizeMB:  100,
			MaxBackups: 5,
			MaxAgeDays: 30,
			Compress:   false,
		}
	}
	return o.File
}
