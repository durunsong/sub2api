## 1. Investigation
- [x] 固定官方 tag、当前 HEAD、增量与定制重叠范围。
- [x] 阅读规则、Fork 清单、调用路径和迁移。

## 2. Implementation
- [x] 上游行为测试先行，观察预期失败。
- [x] 三方合入官方增量并保留定制，版本设为 0.2.3。
- [x] 更新 Fork 清单与版本记录。

## 3. Verification
- [x] Go 默认/unit、vet/build 与前端回归/typecheck。
- [x] 核对官方增量、Fork 定制和历史迁移。
- [x] 编码、冲突标记、最终差异及验证记录。
