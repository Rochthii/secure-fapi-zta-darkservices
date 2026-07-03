# ECC Install Log

**Nguồn**: https://github.com/affaan-m/everything-claude-code  
**Phiên bản**: v1.9.0  
**Ngày cài**: 2026-03-31  
**Target**: Antigravity IDE (`.agent/`)  

---

## Đã cài

### `.agent/rules/` — 15 file (flattened)

| File | Nguồn gốc |
|------|-----------|
| `common-agents.md` | `rules/common/agents.md` |
| `common-code-review.md` | `rules/common/code-review.md` |
| `common-coding-style.md` | `rules/common/coding-style.md` |
| `common-development-workflow.md` | `rules/common/development-workflow.md` |
| `common-git-workflow.md` | `rules/common/git-workflow.md` |
| `common-hooks.md` | `rules/common/hooks.md` |
| `common-patterns.md` | `rules/common/patterns.md` |
| `common-performance.md` | `rules/common/performance.md` |
| `common-security.md` | `rules/common/security.md` |
| `common-testing.md` | `rules/common/testing.md` |
| `ts-coding-style.md` | `rules/typescript/coding-style.md` |
| `ts-hooks.md` | `rules/typescript/hooks.md` |
| `ts-patterns.md` | `rules/typescript/patterns.md` |
| `ts-security.md` | `rules/typescript/security.md` |
| `ts-testing.md` | `rules/typescript/testing.md` |

### `.agent/skills/` — 24 file/dir

**Agent skills (từ `agents/`):**
| File | Vai trò |
|------|---------|
| `database-reviewer.md` | Review Supabase queries, RLS, migrations |
| `security-reviewer.md` | Phân tích bảo mật, OWASP Top 10 |
| `code-reviewer.md` | Review chất lượng code tổng quát |
| `typescript-reviewer.md` | Review TypeScript/JavaScript |
| `planner.md` | Lập kế hoạch implementation |
| `architect.md` | Thiết kế kiến trúc hệ thống |
| `tdd-guide.md` | TDD methodology (Red/Green/Refactor) |
| `build-error-resolver.md` | Sửa lỗi build tự động |
| `e2e-runner.md` | Playwright E2E testing |
| `refactor-cleaner.md` | Dọn dead code |
| `doc-updater.md` | Cập nhật tài liệu |
| `performance-optimizer.md` | Tối ưu hiệu năng |
| `frontend-design/SKILL.md` | Thiết kế UI (custom skill của dự án) |

**Skill workflow docs (từ `skills/*/SKILL.md`):**
| File | Nội dung |
|------|---------|
| `skill-tdd-workflow.md` | Quy trình TDD đầy đủ |
| `skill-security-review.md` | Checklist bảo mật chi tiết |
| `skill-backend-patterns.md` | API, DB, caching patterns |
| `skill-frontend-patterns.md` | React/Next.js patterns |
| `skill-database-migrations.md` | Migration patterns (Prisma, Drizzle) |
| `skill-api-design.md` | REST API design, pagination, errors |
| `skill-deployment-patterns.md` | CI/CD, Docker, rollbacks |
| `skill-e2e-testing.md` | Playwright Page Object Model |
| `skill-verification-loop.md` | Build/test/lint/typecheck loop |
| `skill-postgres-patterns.md` | PostgreSQL optimization |
| `skill-nextjs-turbopack.md` | Next.js + Turbopack patterns |
| `skill-search-first.md` | Research-before-coding workflow |
| `skill-strategic-compact.md` | Context window management |

### `.agent/workflows/` — 17 file (KHÔNG đụng, giữ nguyên)

Đây là workflows tự viết bằng tiếng Việt, phù hợp context dự án — **KHÔNG thay thế bằng ECC**.

---

## Không cài (lý do)

| Component | Lý do bỏ qua |
|-----------|-------------|
| `rules/golang/`, `rules/python/`, `rules/java/`, ... | Dự án 100% TypeScript |
| `skills/django-*`, `skills/laravel-*`, `skills/springboot-*` | Framework không dùng |
| `skills/investor-*`, `skills/article-writing`, `skills/market-research` | Không liên quan |
| `skills/swift-*`, `skills/kotlin-*`, `skills/perl-*`, `skills/cpp-*` | Ngôn ngữ không dùng |
| `commands/` (ECC slash commands) | Đã có workflows tiếng Việt tốt hơn |
| `hooks/` (ECC hooks) | Chưa cần, thêm sau nếu cần memory persistence |
| `mcp-configs/` | Đã có MCP config riêng |

---

## Cập nhật tương lai

Khi cần nâng cấp ECC:
```powershell
# Clone lại repo mới nhất
git clone https://github.com/affaan-m/everything-claude-code.git tmp\ecc --depth=1

# Xem có agent/skill mới nào hữu ích không
# Manually copy những gì cần
```
