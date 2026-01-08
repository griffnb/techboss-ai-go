# Learning Examples: Good vs. Bad

This reference provides extensive examples of learnings worth capturing vs. those that should be skipped.

## Good Learning Examples

These examples show valuable patterns worth codifying into skills.

### Example 1: API Authentication Debugging Pattern

**Session context:**
User was debugging API authentication failures. They tried several approaches before discovering the root cause.

**Learning captured:**
```
When debugging API authentication failures:

1. Verify request is reaching the API (check logs/network tab)
2. Examine request headers - especially Authorization header format
3. Check token/credential validity (expiration, format)
4. Validate request payload structure matches API contract
5. Compare working request (Postman/curl) with app request

Common issues:
- Missing "Bearer " prefix in Authorization header
- Expired tokens not being refreshed
- Header name case sensitivity (Authorization vs authorization)
```

**Why this is good:**
- Systematic approach applicable to future API debugging
- Specific steps with rationale
- Identifies common pitfalls discovered
- Reusable pattern

### Example 2: React Hooks Dependency Array Pattern

**Session context:**
User struggled with infinite render loops when using useEffect. Discovered the correct dependency array pattern.

**Learning captured:**
```
React useEffect dependency array rules:

1. Include all values from component scope used inside effect
2. For functions called in effect, either:
   - Wrap function in useCallback
   - Define function inside effect
   - Use functional updates for setState

Infinite loop indicators:
- Effect updates state that's in dependency array
- Object/array reference changes every render
- Missing eslint-plugin-react-hooks warnings

Solution pattern:
- Use functional setState: setCount(c => c + 1)
- Memoize objects/arrays with useMemo
- Extract non-reactive values outside component
```

**Why this is good:**
- Framework-specific pattern with clear rules
- Identifies problem symptoms
- Provides concrete solutions
- Addresses common React pitfall

### Example 3: Database Query Optimization Pattern

**Session context:**
User's query was slow (3s). Through profiling, discovered N+1 query problem and fixed with joins.

**Learning captured:**
```
Identifying and fixing N+1 query problems:

Symptoms:
- Query executes quickly, but page load slow
- Database logs show many similar queries
- Response time scales linearly with result count

Diagnosis:
1. Enable query logging
2. Count queries per request (look for >10 for simple pages)
3. Identify repeated queries with only ID changing

Solution patterns:
- Use JOIN or INCLUDE to eager load associations
- Batch queries with WHERE id IN (...)
- Add database indexes on foreign keys
- Consider denormalization for read-heavy paths

Example (Rails):
# Bad: N+1
users = User.all
users.each { |u| u.posts.count }  # Query per user

# Good: Eager loading
users = User.includes(:posts)
users.each { |u| u.posts.count }  # Single query
```

**Why this is good:**
- Identifies specific performance problem pattern
- Shows diagnostic steps
- Provides multiple solution approaches
- Includes concrete code example from session

### Example 4: Git Merge Conflict Resolution Strategy

**Session context:**
User had complex merge conflict across multiple files. Developed systematic approach.

**Learning captured:**
```
Resolving complex multi-file merge conflicts:

Process:
1. Understand change intent (read both branch HEADs)
2. Categorize conflicts:
   - Logic changes (need careful merge)
   - Formatting/style (pick one consistently)
   - Additions (usually can keep both)
   - Deletions (verify not still needed)

3. Resolve in order of risk:
   - High risk: Business logic, data models
   - Medium risk: API contracts, configurations
   - Low risk: Tests, documentation, formatting

4. After resolution:
   - Run full test suite
   - Verify application starts
   - Check git diff against both branches
   - Test the originally conflicting functionality

Tools:
- `git diff HEAD...MERGE_HEAD` - See what changed
- `git show :1:file` - See base version
- `git checkout --ours/--theirs file` - Take one side
```

**Why this is good:**
- Systematic workflow for complex problem
- Prioritization strategy
- Multiple concrete commands
- Post-resolution verification steps

### Example 5: Test Data Setup Pattern

**Session context:**
User's tests were flaky due to shared test data. Discovered factory pattern with traits.

**Learning captured:**
```
Reliable test data with factories and traits:

Problem: Shared fixtures cause test interdependence and flakiness

Solution pattern (using factory_bot):

1. Define base factory with required fields:
```ruby
factory :user do
  email { Faker::Internet.email }
  name { Faker::Name.name }
end
```

2. Create traits for variations:
```ruby
factory :user do
  # base fields...

  trait :admin do
    role { 'admin' }
  end

  trait :with_posts do
    after(:create) do |user|
      create_list(:post, 3, user: user)
    end
  end
end
```

3. Use in tests:
```ruby
it 'allows admins to delete' do
  admin = create(:user, :admin)
  post = create(:post)

  expect { admin.delete(post) }.not_to raise_error
end
```

Benefits:
- Each test creates own data (isolation)
- Traits make intent clear
- No database state dependencies
- Faker provides unique values
```

**Why this is good:**
- Solves specific testing problem
- Concrete code pattern with framework
- Shows progression: problem → solution → benefits
- Directly applicable to future tests

## Bad Learning Examples

These examples show sessions that don't warrant skill creation.

### Bad Example 1: Typo Fix

**Session:**
```
User: "Fix the typo in README.md, it says 'instalation' instead of 'installation'"
Assistant: *fixes typo*
```

**Why skip:**
- Trivial operation
- No reusable pattern
- No problem-solving involved
- Common knowledge

### Bad Example 2: Generic Advice

**Potential "learning":**
```
When writing code:
- Use descriptive variable names
- Add comments for complex logic
- Test your code
- Follow PEP 8 style guide
```

**Why skip:**
- Generic advice from any coding guide
- Nothing specific from session
- No novel insight
- Easily found in documentation

### Bad Example 3: Reading Documentation

**Session:**
User asked Claude to explain React documentation, Claude read and summarized.

**Why skip:**
- No application of knowledge
- Just information transfer
- No patterns discovered
- No problem solved

### Bad Example 4: Single Use Solution

**Session:**
User needed to convert one specific CSV file to JSON with custom transformations.

**Potential "learning":**
```
To convert users.csv to users.json:
1. Read CSV with pandas
2. Drop the 'legacy_id' column
3. Rename 'email_address' to 'email'
4. Write as JSON
```

**Why skip:**
- Too specific to one file
- No reusable pattern
- One-off data transformation
- Not applicable to future sessions

### Bad Example 5: Standard CRUD Operations

**Session:**
User asked to create a new model with standard CRUD endpoints.

**Why skip:**
- Standard framework usage
- No novel patterns
- Common operation covered in framework docs
- No debugging or problem-solving

### Bad Example 6: Vague "Best Practice"

**Potential "learning":**
```
It's good to use environment variables for configuration because it's more secure and flexible.
```

**Why skip:**
- Too vague and general
- No specific implementation pattern
- No examples from session
- Common knowledge

## Distinguishing Valuable from Trivial

### Questions to Ask

**Is this reusable?**
- ✅ Yes: "API debugging checklist applicable to any REST API"
- ❌ No: "Fix this specific typo in this specific file"

**Is this non-obvious?**
- ✅ Yes: "React hooks dependency rules and infinite loop patterns"
- ❌ No: "Use descriptive variable names"

**Did problem-solving occur?**
- ✅ Yes: "Debugged N+1 queries through profiling and query log analysis"
- ❌ No: "Read documentation and understood concept"

**Is there a concrete pattern?**
- ✅ Yes: "Factory pattern with traits for test isolation"
- ❌ No: "Testing is important"

**Will this help future work?**
- ✅ Yes: "Merge conflict resolution strategy for multi-file conflicts"
- ❌ No: "Converted this one CSV file to JSON"

## Skill Creation Decision Tree

```
Session ends
    ↓
Was code written or debugging done?
    No → Skip
    Yes → Continue
    ↓
Are there patterns or insights?
    No → Skip
    Yes → Continue
    ↓
Is it reusable in future sessions?
    No → Skip
    Yes → Continue
    ↓
Is it non-obvious or framework-specific?
    No → Skip
    Yes → Continue
    ↓
Can it be explained concretely?
    No → Skip
    Yes → Create/update skill! ✅
```

## Edge Cases

### Borderline Case: Configuration Pattern

**Session:** User configured webpack for the first time, following official guide with some customization.

**Decision:** **Capture if** the customization addressed project-specific needs not in docs.
**Skip if** just following the standard setup guide.

### Borderline Case: Debugging Session

**Session:** User spent 30 minutes debugging, tried many things, eventually found typo.

**Decision:** **Capture if** the debugging process revealed a systematic approach or common pitfall.
**Skip if** just trial and error with no transferable technique.

### Borderline Case: Framework Learning

**Session:** User learned basic React hooks (useState, useEffect) from examples.

**Decision:** **Capture if** discovered non-obvious patterns or gotchas through experience.
**Skip if** just learned basic usage covered in official tutorial.

## Summary

**Capture learnings that are:**
- Reusable patterns
- Non-obvious techniques
- Framework-specific gotchas
- Systematic approaches
- Problem-solving workflows
- Architecture decisions with rationale

**Skip learnings that are:**
- Trivial operations
- Generic advice
- One-off solutions
- Standard operations
- Pure information transfer
- Common knowledge

**When in doubt:** Ask "Will this specific pattern help me 3 months from now on a different but related problem?" If yes, capture it. If no, skip it.
