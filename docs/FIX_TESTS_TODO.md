# Fix Tests TODO

## Summary
Total failing test packages: 10 (down from 24)
- ✅ Config file issue fixed (23 packages now passing)
- ❌ 10 packages with actual test failures remaining

## Fixed Issues

✅ **Config File Missing** - Created `.configs/unit_test.json` from `infra/unit_test.json`
- All 23 packages that were failing due to missing config now pass

## Remaining Test Failures

### 1. ✅ `github.com/griffnb/techboss-ai-go/internal/models/account` - TestSave
   - Error: `Didnt Save` - value not persisted properly
   - Location: `account_test.go:46`

### 2. ✅ `github.com/griffnb/techboss-ai-go/internal/models/admin` - 2 tests
   - **TestSave**: `Didnt Save` - value not persisted properly (admin_test.go:49)
   - **TestDupe**: Duplicate key constraint violation not properly handled (admin_test.go:199)

### 3. ✅ `github.com/griffnb/techboss-ai-go/internal/models/agent_attribute` - 3 tests
   - **TestSave**: `Table doesnt have any properties, be sure to AddTableToProperties agent_attributes`
   - **TestFindAll**: Same property error
   - **TestFindFirst**: Same property error

### 4. ✅ `github.com/griffnb/techboss-ai-go/internal/models/conversation` - 3 tests
   - **TestSave**: `Table doesnt have any properties, be sure to AddTableToProperties conversations`
   - **TestFindAll**: Same property error
   - **TestFindFirst**: Same property error

### 5. ✅ `github.com/griffnb/techboss-ai-go/internal/models/document` - 3 tests
   - **TestSave**: `Table doesnt have any properties, be sure to AddTableToProperties documents`
   - **TestFindAll**: Same property error
   - **TestFindFirst**: Same property error

### 6. ✅ `github.com/griffnb/techboss-ai-go/internal/models/global_config` - 3 tests
   - **TestFindAll**: `unable to encode 54871 into text format for text (OID 25)`
   - **TestFindFirst**: Same encoding error
   - **TestGet**: Same encoding error

### 7. ✅ `github.com/griffnb/techboss-ai-go/internal/models/message` - TestGetMessagesByConversationID
   - Error: `Cannot do operations on a non-existent table` (DynamoDB table missing)
   - Location: `message_test.go:31`

### 8. ✅ `github.com/griffnb/techboss-ai-go/internal/models/subscription` - TestSave
   - Error: `Didnt Save` - value not persisted properly
   - Location: `subscription_test.go:46`

### 9. ✅ `github.com/griffnb/techboss-ai-go/internal/services/ai_proxies/openai` - TestProxyRequest_MockOpenAI
   - Error: `Expected /chat/completions path, got /responses`
   - Location: `client_test.go:69`

### 10. ✅ `github.com/griffnb/techboss-ai-go/internal/services/dynamo_queue` - TestReprocessDynamoThrottle
   - Error: `Cannot do operations on a non-existent table` (DynamoDB table missing)
   - Location: `reprocess_test.go:62`
