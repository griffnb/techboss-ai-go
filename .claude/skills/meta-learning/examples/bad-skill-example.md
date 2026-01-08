# Bad Skill Example

This shows what NOT to create - a poorly crafted skill that shouldn't exist.

---

```markdown
---
name: coding-tips
description: Use this when you need help with coding.
version: 1.0.0
---

# General Coding Tips

## Overview

You should follow good coding practices when writing code. This skill helps you write better code.

## Things You Should Do

When you're coding, you should:

- Write clean code
- Use good variable names
- Add comments
- Test your code
- Follow style guides
- Don't repeat yourself (DRY)
- Keep it simple (KISS)
- Use version control

## Writing Functions

You should write functions that do one thing. Functions should have good names that describe what they do. You should add docstrings to your functions.

Example:
```python
def calculate(x, y):
    return x + y
```

## Using Git

You should use Git for version control. You can commit your changes with git commit. You should write good commit messages.

Commands:
- `git add .`
- `git commit -m "message"`
- `git push`

## Debugging

When you have a bug, you should try to debug it. You can use print statements or a debugger. You should figure out what's causing the bug and fix it.

## Best Practices

You should always follow best practices when coding. This includes:
- Writing tests
- Reviewing code
- Documenting your work
- Learning new things

Remember that good code is important!
```

---

## Why This is a Bad Skill

### Problem 1: Weak Description

**Current:**
```yaml
description: Use this when you need help with coding.
```

**Issues:**
- Not third-person
- Extremely vague
- No specific trigger phrases
- Could apply to anything

**Should be:**
```yaml
description: This skill should be used when working with [specific framework], debugging [specific problem type], or implementing [specific pattern].
```

### Problem 2: Generic Content

Everything in this skill is:
- Common knowledge
- Available in any coding tutorial
- Not specific to any framework or language
- No novel insights
- Nothing learned from actual session

### Problem 3: Wrong Writing Style

**Uses second person throughout:**
- "You should follow..."
- "You can commit..."
- "You should write..."

**Should use imperative:**
- "Follow good coding practices"
- "Commit changes with git commit"
- "Write functions that do one thing"

### Problem 4: No Concrete Examples

The examples provided are trivial:
- `calculate(x, y)` - too simple to be useful
- Basic git commands - available in any git tutorial
- "Use print statements" - no specific technique

**Should have:**
- Real examples from session
- Specific patterns discovered
- Before/after comparisons
- Actual problems solved

### Problem 5: Overly Broad Scope

This skill tries to cover:
- General coding practices
- Functions
- Git usage
- Debugging
- Best practices

**Should:**
- Focus on one specific area
- Go deep rather than broad
- Be applicable to specific situations

### Problem 6: No Session Context

Nothing indicates this came from actual work:
- No specific problem solved
- No "discovered through debugging"
- No performance improvements
- No patterns learned

**Should:**
- Reference the session problem
- Show what was tried
- Explain what worked and why

### Problem 7: No Reusable Patterns

This doesn't provide:
- Systematic approach to anything
- Diagnostic steps
- Decision criteria
- Specific techniques

**Should:**
- Show step-by-step workflows
- Provide decision trees
- Explain when to use each approach

## What Should Have Been Done Instead

**Option 1: Don't create this skill**
- Content is too generic
- No value added beyond common knowledge
- Nothing specific learned

**Option 2: If session had valuable learnings, create focused skill**

Example: If session involved debugging Python decorators, create:

```yaml
---
name: python-decorator-debugging
description: This skill should be used when debugging Python decorators, troubleshooting decorator order issues, or understanding decorator execution flow.
---

# Python Decorator Debugging Patterns

[Specific patterns discovered during actual debugging session]
[Concrete examples from the code worked on]
[Systematic approach that proved effective]
```

## Red Flags Checklist

If a skill has these issues, it should NOT be created:

- [ ] ❌ Vague description without specific triggers
- [ ] ❌ Generic advice from any tutorial
- [ ] ❌ Uses second person ("you should")
- [ ] ❌ Overly broad scope (covers multiple unrelated topics)
- [ ] ❌ No connection to actual session work
- [ ] ❌ Trivial examples (Hello World level)
- [ ] ❌ No novel patterns or insights
- [ ] ❌ Common knowledge easily found in docs

If 3 or more boxes checked, DON'T create the skill.

## Summary

**This bad skill fails because:**
1. Too generic and vague
2. No specific learnings from session
3. Wrong writing style (second person)
4. Overly broad scope
5. No reusable patterns
6. Common knowledge only
7. Trivial examples

**Remember:** Only create skills that capture specific, valuable, reusable patterns discovered through actual development work. Quality over quantity!
