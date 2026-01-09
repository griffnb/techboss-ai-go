# Good Skill Example

This is an example of a well-crafted skill created from session learnings.

---

```markdown
---
name: postgres-jsonb-querying
description: This skill should be used when working with PostgreSQL JSONB columns, querying JSON data in Postgres, using JSONB operators, or troubleshooting JSON query performance issues.
version: 1.0.0
---

# PostgreSQL JSONB Query Patterns

## Overview

PostgreSQL's JSONB type enables storing and querying JSON data efficiently. This skill covers query patterns, operator usage, and indexing strategies discovered through debugging slow JSON queries and learning optimal access patterns.

## Key Concepts

**JSONB vs JSON:**
- JSONB stores binary format (faster queries, slightly slower writes)
- JSON stores exact text (preserves formatting and key order)
- Use JSONB for querying, JSON only if exact format preservation needed

**Common Operators:**
- `->` - Get JSON object field as JSON
- `->>` - Get JSON object field as text
- `#>` - Get nested object at path as JSON
- `#>>` - Get nested object at path as text
- `@>` - Contains (left JSON contains right JSON)
- `?` - Exists (key exists)
- `?|` - Exists any (any of the keys exist)
- `?&` - Exists all (all keys exist)

## Query Patterns

### Accessing Nested Fields

**Get top-level field:**
```sql
SELECT data->>'name' FROM users;
-- Returns: "John Doe" (text)

SELECT data->'name' FROM users;
-- Returns: "John Doe" (json)
```

**Get nested field:**
```sql
SELECT data#>>'{address,city}' FROM users;
-- Path: data['address']['city']

SELECT data->'address'->>'city' FROM users;
-- Equivalent chained syntax
```

### Filtering by JSON Content

**Exact field match:**
```sql
SELECT * FROM users WHERE data->>'status' = 'active';
```

**Contains check (useful for arrays or nested objects):**
```sql
-- Check if user has specific role
SELECT * FROM users WHERE data->'roles' @> '"admin"';

-- Check if object contains sub-object
SELECT * FROM users
WHERE data @> '{"preferences": {"notifications": true}}';
```

**Key existence:**
```sql
-- Single key
SELECT * FROM users WHERE data ? 'email';

-- Any of these keys
SELECT * FROM users WHERE data ?| array['email', 'phone'];

-- All of these keys
SELECT * FROM users WHERE data ?& array['email', 'name'];
```

### Array Operations

**Check array contains value:**
```sql
SELECT * FROM posts
WHERE data->'tags' @> '["postgresql"]';
```

**Array length:**
```sql
SELECT jsonb_array_length(data->'tags') FROM posts;
```

**Expand array to rows:**
```sql
SELECT jsonb_array_elements(data->'tags') FROM posts;
```

## Indexing for Performance

### When to Index

Index JSONB columns when:
- Querying specific fields frequently
- Filtering by JSON content in WHERE clauses
- JSON column is large (>1KB average)
- Table has >10,000 rows

### GIN Index (Recommended)

**Create GIN index for general querying:**
```sql
CREATE INDEX idx_users_data ON users USING GIN (data);
```

**Benefits:**
- Supports `@>`, `?`, `?|`, `?&` operators
- Good for general-purpose JSON querying
- Relatively small index size

**Use when:**
- Querying various fields
- Using contains (`@>`) operator
- Checking key existence

### Expression Index for Specific Fields

**Create index on extracted field:**
```sql
CREATE INDEX idx_users_email ON users ((data->>'email'));
```

**Benefits:**
- Faster for equality checks on specific field
- Smaller index size than GIN
- Works with standard comparison operators

**Use when:**
- Querying one field frequently
- Need exact match or range queries
- Performance critical query

### Index Usage Example

**Without index (slow):**
```sql
EXPLAIN ANALYZE
SELECT * FROM users WHERE data->>'email' = 'user@example.com';
-- Seq Scan (2.3ms for 10k rows)
```

**With expression index (fast):**
```sql
CREATE INDEX idx_users_email ON users ((data->>'email'));

EXPLAIN ANALYZE
SELECT * FROM users WHERE data->>'email' = 'user@example.com';
-- Index Scan (0.08ms)
```

## Common Pitfalls

### Pitfall 1: Using -> Instead of ->>

**Problem:**
```sql
-- Wrong: Returns JSON, not text
WHERE data->'status' = 'active'  -- Won't match!
```

**Solution:**
```sql
-- Correct: Returns text
WHERE data->>'status' = 'active'  -- Matches

-- Or compare JSON
WHERE data->'status' = '"active"'  -- Note quotes
```

### Pitfall 2: Not Using Indexes

**Problem:**
Query is slow despite small result set because full table scan occurs.

**Solution:**
- Add GIN index for general querying
- Add expression index for frequently queried fields
- Check EXPLAIN ANALYZE to verify index usage

### Pitfall 3: Incorrect Path Syntax

**Problem:**
```sql
-- Wrong: String instead of array
data#>>address.city

-- Wrong: Quotes around path
data#>>'{"address","city"}'
```

**Solution:**
```sql
-- Correct: Array syntax with braces
data#>>'{address,city}'

-- Correct: Chained operator syntax
data->'address'->>'city'
```

### Pitfall 4: Forgetting Type Coercion

**Problem:**
```sql
-- Won't work: comparing text to integer
WHERE data->>'age' > 25
```

**Solution:**
```sql
-- Cast to integer for comparison
WHERE (data->>'age')::integer > 25
```

## Best Practices

1. **Use JSONB over JSON** for querying use cases
2. **Index based on query patterns** - GIN for general, expression for specific
3. **Use ->> for text extraction** in WHERE clauses
4. **Validate JSON before insert** to avoid malformed data
5. **Keep JSON structure consistent** across rows for easier querying
6. **Extract to columns** if querying field very frequently (>50% of queries)
7. **Monitor index size** - GIN indexes on large JSONB can grow significantly

## Example from Session

During debugging of slow user search, discovered:

**Original slow query (no index):**
```sql
SELECT * FROM users
WHERE data->>'organization_id' = '123'
ORDER BY data->>'created_at' DESC;
-- Execution time: 2,340ms
```

**Optimized with expression indexes:**
```sql
CREATE INDEX idx_users_org_id ON users ((data->>'organization_id'));
CREATE INDEX idx_users_created ON users ((data->>'created_at'));

SELECT * FROM users
WHERE data->>'organization_id' = '123'
ORDER BY data->>'created_at' DESC;
-- Execution time: 18ms
```

**Key insight:** Expression indexes dramatically improve performance for specific field queries, even better than GIN index for this use case.

---

*This skill was created from session learnings about optimizing PostgreSQL JSONB queries.*
```

---

## Why This is a Good Skill

**Strong trigger description:**
- Third-person format
- Specific phrases: "working with JSONB", "querying JSON data", "JSONB operators"
- Covers multiple entry points

**Well-structured content:**
- Clear overview
- Key concepts explained
- Multiple pattern sections
- Real examples from session
- Common pitfalls identified
- Best practices distilled

**Actionable and specific:**
- Concrete SQL examples
- Performance comparisons
- Shows before/after optimization
- Explains when to use each approach

**Appropriate length:**
- ~1,800 words (in target range)
- Focused on JSONB querying
- Could expand to references/ if needed for advanced topics

**Imperative form:**
- "Use JSONB over JSON"
- "Index based on query patterns"
- "Validate JSON before insert"
- No "you should" language
