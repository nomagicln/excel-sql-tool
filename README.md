# Excel-SQL-Tool

让大模型通过 SQL 查询 Excel 文件的 CLI 工具。

## 背景

传统方式：LLM 全量读取 Excel → 成本高、速度慢、context 容易爆

新方式：LLM 生成 SQL → 精准查询只返回需要的数据

灵感来源：[x2bool/xlite](https://github.com/x2bool/xlite)

## 安装

```bash
go install github.com/nomagicln/excel-sql-tool@latest
```

或源码编译：

```bash
git clone git@github.com:nomagicln/excel-sql-tool.git
cd excel-sql-tool
go build -o excel-sql-tool ./cmd/cli/
```

## 使用方式

### 1. inspect - 审视 Excel，提取元数据

```bash
./excel-sql-tool inspect data.xlsx --output metadata.json
```

可选：`--config config.yaml` 指定 sheet 配置

### 2. generate - 生成 SKILL.md

```bash
./excel-sql-tool generate config.yaml metadata.json --output SKILL.md
```

### 3. query - 执行 SQL 查询

```bash
./excel-sql-tool query data.xlsx "SELECT * FROM Sheet1 LIMIT 10" --sheet Sheet1
```

### 4. server - 启动后台服务

```bash
./excel-sql-tool server --port 8080
```

## 配置示例 (config.yaml)

```yaml
name: "华东销售查询"
description: "华东区代理商销售数据查询工具"
examples:
  - "Q3销售额前10的代理商"
  - "同比增长超过20%的产品"
domain:
  - "代理商等级: S/A/B/C"
  - "区域编码: HD01-华东, HD02-华南"
sheets:
  - name: "销售明细"
    header_row: 1
    data_start_row: 2
```

## Excel 约束

- 不能有合并单元格（不支持）
- 第一行是表头（或在配置中指定）
- 不支持公式，只读值
- 支持 .xlsx 和 .xls 格式

## 技术栈

- Go 1.23+
- [excelize](https://github.com/xuri/excelize/v2) - Excel 解析
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) - SQLite 引擎

## License

MIT
