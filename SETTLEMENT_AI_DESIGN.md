## go-shoppings 结算系统（AI能力融合）设计草案

### 一、服务边界与核心流程

1. 平台统一收款，账期结算给租户，支持多种结算方式（manual_settlement/wechat/alipay/bank）。
2. 账期结算流程：
   - 账期内所有归属租户的订单货款归集统计
   - 结算资料审核通过后，平台财务发起结算（可自动/手动）
   - 生成结算单，推送租户确认，完成放款
3. 分账/分润、分销佣金、积分等与结算周期解耦，分别归属各自业务表
4. 对账、异常检测、结算报告等为平台与租户双方提供透明可追溯的结算体验

### 二、AI能力切入点

1. 智能对账：自动比对平台流水、第三方支付账单、租户订单，发现差异并生成对账建议
2. 异常检测：基于历史数据，识别异常交易、可疑结算、重复/漏结算等风险
3. 预测分析：预测未来账期结算金额、租户活跃度、坏账风险等，辅助财务决策
4. 结算报告自动生成：自动汇总账期内结算明细、异常、建议，生成可导出报告
5. 智能问答/解释：租户可查询结算明细、异常原因、对账建议，AI自动生成解释

### 三、关键API与数据结构（草案）

- GET /api/v1/platform/settlements?page&size&tenant_id&status
  - 查询结算单列表，支持多条件筛选
- GET /api/v1/platform/settlements/{id}
  - 查询结算单详情，含AI对账结果、异常提示、结算明细
- POST /api/v1/platform/settlements/trigger
  - 触发账期结算（可选参数：账期、租户、自动/手动）
- GET /api/v1/platform/settlements/{id}/ai-report
  - 获取AI生成的结算报告、对账建议、异常解释
- POST /api/v1/platform/settlements/{id}/confirm
  - 租户确认结算单，平台放款
- GET /api/v1/platform/settlements/{id}/qa?question=xxx
  - 针对结算单明细/异常，AI问答解释

#### 结算单 Settlement
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 结算单ID |
| tenant_id | uint64 | 归属租户 |
| period_start | datetime | 账期起 |
| period_end | datetime | 账期止 |
| total_amount | decimal | 结算总金额 |
| status | int | 状态（待确认/已确认/已放款/异常）|
| ai_report | json | AI对账/异常分析结果 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### AI对账/异常分析结构（ai_report）
| 字段 | 类型 | 说明 |
|------|------|------|
| diff_items | array | 差异明细（如平台与支付账单不符）|
| anomaly_flags | array | 异常类型标记（如大额、频繁、重复等）|
| suggestions | array | 处理建议 |
| summary | string | AI生成摘要 |

### 四、验证与回滚机制

1. 所有结算/对账/AI分析结果均可追溯原始数据，支持人工复核与回滚
2. 结算单状态流转（待确认→已确认→已放款/异常）全程日志与操作记录
3. AI能力仅做辅助，关键放款/结算需人工确认，异常需平台/租户双向确认
4. 结算单/AI报告/对账明细均可导出，便于外部审计与合规

### 五、AI能力实现建议

1. 可用大模型/LLM（如 GPT-4/企业私有大模型）结合结构化数据分析与自然语言生成
2. 训练/微调数据建议来源于历史结算单、对账明细、异常案例、财务审核意见
3. 关键AI推理过程需保留原始输入、输出与解释链路，便于溯源与合规
4. 支持多租户隔离，AI分析仅基于本租户/本账期数据

---
本设计草案为 go-shoppings 结算系统 AI 能力融合的初步方案，后续可根据业务反馈与合规要求持续迭代。