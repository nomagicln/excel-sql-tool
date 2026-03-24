---
name: excel-query-销售数据查询
description: 代理商销售数据查询工具，支持区域统计、代理商排名等
---

# 销售数据查询

## 支持的查询类型

- 华东区域销售总额

- 销售额前3的代理商

- 3月份销售数据

## 数据 Schema

### Sheet: 销售明细
| 列名 | 类型 | 说明 |
|------|------|------|
| 代理商 | string |  |
| 区域 | string |  |
| 销售额 | number |  |
| 日期 | date |  |

### Sheet: 产品表
| 列名 | 类型 | 说明 |
|------|------|------|
| 产品 | string |  |
| 类别 | string |  |

## 领域知识

- 代理商等级: S级>100万, A级>50万, B级>20万, C级<20万

- 区域: 华东、华南、华北、华中、西部

## SQL 示例

```sql
SELECT * FROM "销售明细" LIMIT 100;
SELECT "代理商", COUNT(*) FROM "销售明细" GROUP BY "代理商";
SELECT * FROM "产品表" LIMIT 100;
SELECT "产品", COUNT(*) FROM "产品表" GROUP BY "产品";
```
