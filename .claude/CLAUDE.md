## ⚠️ CRITICAL: ALWAYS FOLLOW DOCUMENTATION AND PRD
**MANDATORY REQUIREMENT**: Before making ANY changes to this codebase, you MUST:

1. **Read the PRD first if it exists**: All requirements and decisions are based on @docs/PRD.md - this is the single source of truth
2. **Follow the documentation**: All implementation details are documented in Instructions for models are in
@docs/MODELS.md

Instructions for controllers are in
@docs/CONTROLLERS.md

3. **Maintain consistency**: Any new features, APIs, or changes must align with existing patterns
4. **Verify against PRD**: Every implementation decision should trace back to a requirement in the PRD
5. **UPDATE CHECKLISTS**: ALWAYS update `/docs/{FEATURE}_TODO.md` when completing phases or major features
6. If go docs are missing from a function or package, and you learn something important about it, ADD TO YOUR TODO LIST THAT YOU NEED TO UPDATE THAT GO DOC WITH WHAT YOU LEARNED
7. **VERY IMPORTANT** Do not make large files with lots of functionality.  Group functions together into files that relate them together.  This makes it easier to find grouped functions and their associated tests.  **LARGE FILES ARE BAD**


## 🔄 CHECKLIST UPDATE POLICY

**NEVER FORGET**: When you complete any phase, feature, or major milestone:

1. **IMMEDIATELY** update `/docs/{FEATURE}_TODO.md` to mark items as completed
2. **ADD NEW PHASES** to the checklist as they are planned and implemented  
3. **KEEP DOCUMENTATION CURRENT** - the checklist should always reflect the actual project state
4. **UPDATE STATUS** for any infrastructure, integrations, or features that are now working

This ensures the checklist remains an accurate reflection of project progress and helps future development sessions.

**When implementing new features**:
1. Check if it exists in PRD requirements
2. Follow established patterns and conventions
3. Update documentation if adding new patterns


**IMPORTANT Before you begin, always launch the context-fetcher sub agent to gather the information required for the task.**