---
name: meta-learning
description: This skill should be used when analyzing development sessions to extract learnings, determining what constitutes valuable knowledge to capture, creating skills from session insights, or evaluating whether patterns are worth codifying for reuse.
version: 1.0.0
---

# Meta-Learning: Capturing Knowledge from Development Sessions

## Overview

Meta-learning is the practice of extracting reusable knowledge from development sessions and codifying it into skills. Not every session contains valuable learnings—this skill provides criteria for identifying what's worth capturing and how to transform insights into effective skills.

**Key principle:** Capture patterns and insights that will genuinely help future work. Quality over quantity.

## Identifying Valuable Learnings

### What Makes a Good Learning

Valuable learnings have these characteristics:

1. **Reusable** - Applicable to future similar situations
2. **Non-obvious** - Not common knowledge or easily discoverable
3. **Pattern-based** - Demonstrates a technique, approach, or principle
4. **Context-rich** - Includes the "why" behind decisions, not just "what" was done
5. **Actionable** - Provides concrete steps or examples

### Categories of Valuable Learnings

**Debugging Patterns:**

- Systematic approaches that worked (e.g., "When API fails, check headers first, then payload format, then auth")
- Root cause identification techniques
- Tool usage for diagnosis (e.g., using curl to isolate API issues)

**Architecture Decisions:**

- Trade-offs considered and chosen approach with rationale
- Pattern selections (e.g., repository pattern for database access)
- Structure decisions that solve specific problems

**Framework/Library Patterns:**

- Non-obvious usage patterns discovered
- Common pitfalls and how to avoid them
- Integration approaches that work well together
- Configuration patterns for specific use cases

**Testing Strategies:**

- Effective test organization approaches
- Mocking patterns that proved useful
- Coverage strategies for specific scenarios

**Domain Knowledge:**

- Business logic patterns specific to the project
- Data model relationships and constraints
- API contracts and integration patterns

**Rules:**

- If something is global and better set as a rule, i.e. "use bun not npm" then add it to ./claude/rules/{rule}.md
- If it applies only for a specific file type, you can add a frontmatter with paths, <example>paths: \*.go</example>

**Sticking Points:**

- If the agent has trouble finding or executing something it should and does multiple tool calls to figure it out
- Wasted context or calls searching for things the wrong way, did they use code_tools? did they use it wrong?

### What to Skip

**Don't capture these:**

- Trivial operations (typo fixes, simple CRUD)
- Generic advice easily found in documentation
- One-off solutions without reusable patterns
- Purely conversational exchanges without technical work
- Simple reads of documentation without application

## Quality Thresholds

Apply these filters to sessions before capturing:

**Minimum Activity Level:**

- At least 5 tool calls (Read, Write, Edit, Bash, etc.)
- At least 2-3 minutes of technical work
- Some problem-solving or decision-making occurred

**Substantive Content:**

- Not just reading existing code
- Not just cosmetic changes
- Involves learning something new or applying a pattern

**Clarity:**

- The pattern or insight is clear enough to explain
- Steps can be articulated concretely
- Examples from the session illustrate the point

## Creating Skills from Learnings

### Skill Structure

Follow skill-development best practices:

**Frontmatter (required):**

```yaml
---
name: skill-topic-name
description: This skill should be used when [specific trigger conditions]. Be concrete with trigger phrases.
version: 1.0.0
---
```

**Body structure:**

1. **Overview** - Brief explanation of what this skill covers (2-3 sentences)
2. **Key Concepts** - Core ideas or patterns (bullet points or short sections)
3. **Patterns/Techniques** - Specific approaches learned, with examples from session
4. **Best Practices** - Distilled guidance from the experience
5. **Common Pitfalls** - Problems encountered and solutions (optional)
6. **Examples** - Concrete code or command examples from session

**Target length:** 1,500-2,000 words for body

**Writing style:** Imperative form (verb-first instructions), not second person

### Naming Skills

Use kebab-case with descriptive names:

**Good names:**

- `api-authentication-debugging`
- `react-hooks-dependency-patterns`
- `postgres-query-optimization`
- `go-error-handling-patterns`

**Bad names:**

- `debugging` (too generic)
- `stuff-learned` (not descriptive)
- `session-notes` (not focused)

### Smart Merge Strategy

**Determine whether to create new or update existing:**

1. **Read existing skills** - Use Glob to find skills, Read to examine content
2. **Calculate topic overlap:**
   - > 60% overlap → Update existing skill
   - <60% overlap → Create new skill
3. **Overlap indicators:**
   - Same framework/library
   - Same type of problem (e.g., both about authentication)
   - Same technical domain (e.g., both about database queries)

**When updating existing skills:**

- Preserve existing structure and style
- Add new patterns to relevant sections
- Include new examples from session
- Update version (bump patch: 1.0.1 → 1.0.2)
- Add note in "Recent Additions" if significant

**When creating new skills:**

- Start with template structure above
- Use examples from session throughout
- Ensure description has strong trigger phrases
- Set version to 1.0.0

## Skill Quality Standards

Every generated skill must meet these standards:

**Description Quality:**

- Third-person format ("This skill should be used when...")
- Includes 3-5 specific trigger phrases users would say
- Concrete, not vague

**Content Quality:**

- Uses imperative/infinitive form throughout
- Specific examples from the session included
- Avoids generic advice
- Focused on actionable guidance

**Structure Quality:**

- Clear sections with headings
- Bullet points for lists
- Code examples formatted properly
- Consistent style with plugin's other skills

**Version Tracking:**

- New skills start at 1.0.0
- Updates bump patch version (1.0.0 → 1.0.1)

## Implementation Workflow

Follow this process when invoked to capture learnings:

1. **Read configuration** - Check `.claude/session-learner.local.md` for settings if it exists
2. **Apply quality threshold** - Verify session meets minimum criteria
3. **Analyze session** - Identify patterns, decisions, techniques learned
4. **Evaluate existing skills** - Search for related skills (smart merge decision)
5. **Generate or update skill:**
   - New: Create directory and SKILL.md with proper structure
   - Update: Use Edit to add content to existing skill
6. **Stage changes** - If `git add` modified skills
7. **Report results** - Summarize what was captured and where

## Examples of Good vs. Bad Learnings

For detailed examples showing what to capture and what to skip, see:

- **[`references/learning-examples.md`](./references/learning-examples.md)** - Extensive examples with commentary
- **[`examples/good-skill-example.md`](./examples/good-skill-example.md)** - Complete skill created from session
- **[`examples/bad-skill-example.md`](./examples/bad-skill-example.md)** - What NOT to create

## Best Practices

**DO:**

- ✅ Apply quality thresholds strictly
- ✅ Include concrete examples from session
- ✅ Use smart merge strategy
- ✅ Write in imperative form
- ✅ Create focused, specific skills
- ✅ Stage skills for PR when configured

**DON'T:**

- ❌ Capture trivial operations
- ❌ Create generic advice skills
- ❌ Use second person writing
- ❌ Create duplicate skills without checking existing
- ❌ Skip version tracking
- ❌ Force capture when session has no valuable learnings

## Additional Resources

### Reference Files

For detailed patterns and examples:

- **[`references/learning-examples.md`](./references/learning-examples.md)** - Good vs bad learning examples with detailed commentary

### Example Files

Working examples in `examples/`:

- **[`good-skill-example.md`](./examples/good-skill-example.md)** - Complete skill showing best practices
- **[`bad-skill-example.md`](./examples/bad-skill-example.md)** - Anti-pattern showing what to avoid

---

**Remember:** The goal is compounding knowledge—each captured learning should make future sessions easier. Be selective and focus on truly valuable patterns that warrant codification.
