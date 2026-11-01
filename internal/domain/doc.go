// Package domain 放业务模型、纯规则和领域错误。
//
// 约定：
//   - 只依赖标准库，不 import HTTP、SQL、pgx、service 或 repository。
//   - 纯规则用构造函数、Normalize/Validate 方法或纯函数表达。
//   - 需要数据库状态或事务的规则留在 repository，service 只做编排。
//   - JSON tag 保留，供响应复用；HTTP 请求体仍用 httpapi 的 DTO。
//   - 移入本包的纯规则必须有单元测试。
package domain
