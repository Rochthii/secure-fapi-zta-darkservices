# Documentation & Sync Workflow

Run this workflow whenever a major architectural change or feature is completed.

1. **Identify Scope**: List all files, database tables, RLS policies, or APIs changed.
2. **Update Type-Specific Docs**:
   - *Architecture/Routing*: Update `docs/ARCHITECTURE.md` or relative design docs.
   - *Database Schema/Migration/RLS*: Update RLS docs or schemas.
   - *Standalone PDP Integration*: Update `README.md` and benchmark files.
3. **Roadmap & Master Plan Sync**:
   - Check off completed items (`[x]`) in `docs/13_IMPLEMENTATION_ROADMAP.md` and `docs/16_FINAL_MASTER_PLAN.md`.
4. **Consistency check**: Ensure obsolete configurations/features are deleted from docs.
5. **Changelog Entry**: Add new entries to `CHANGELOG.md` under `## [YYYY-MM-DD] - Feature Title`.
6. **Commit Docs**: Commit documentation changes separately with prefix `docs: ...`.
