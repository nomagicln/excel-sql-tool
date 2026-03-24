package main

import (
    "fmt"
    "github.com/xuri/excelize/v2"
)

func main() {
    f := excelize.NewFile()
    defer f.Close()
    
    // Sheet1: 销售明细
    f.SetSheetName("Sheet1", "销售明细")
    headers := []string{"代理商", "区域", "销售额", "日期"}
    for i, h := range headers {
        cell, _ := excelize.CoordinatesToCellName(i+1, 1)
        f.SetCellValue("销售明细", cell, h)
    }
    
    data := [][]interface{}{
        {"代理商A", "华东", 100000, "2026-01-15"},
        {"代理商B", "华东", 85000, "2026-01-20"},
        {"代理商C", "华南", 120000, "2026-02-01"},
        {"代理商D", "华南", 95000, "2026-02-10"},
        {"代理商E", "华北", 150000, "2026-02-15"},
        {"代理商F", "华北", 110000, "2026-03-01"},
        {"代理商G", "华东", 200000, "2026-03-10"},
        {"代理商H", "华南", 80000, "2026-03-15"},
    }
    
    for rowIdx, row := range data {
        for colIdx, val := range row {
            cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
            f.SetCellValue("销售明细", cell, val)
        }
    }
    
    f.NewSheet("产品表")
    f.SetCellValue("产品表", "A1", "产品")
    f.SetCellValue("产品表", "B1", "类别")
    f.SetCellValue("产品表", "A2", "产品X")
    f.SetCellValue("产品表", "B2", "A类")
    f.SetCellValue("产品表", "A3", "产品Y")
    f.SetCellValue("产品表", "B3", "B类")
    
    if err := f.SaveAs("tests/test_data.xlsx"); err != nil {
        fmt.Println(err)
    }
    fmt.Println("Created tests/test_data.xlsx")
}
