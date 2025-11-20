# Fix Tests TODO

## Summary
Total failing test packages: 24
- 23 packages failing due to missing config file: `/home/runner/work/techboss-ai-go/techboss-ai-go/.configs/unit_test.json`
- 1 package failing due to actual test logic issue

## Broken Tests

### Config File Missing Issues
All these packages fail with: `FATAL Read File Error: open /home/runner/work/techboss-ai-go/techboss-ai-go/.configs/unit_test.json: no such file or directory`

1. ❌ `github.com/griffnb/techboss-ai-go/internal/common`
2. ❌ `github.com/griffnb/techboss-ai-go/internal/cron/taskworker/delay_queue`
3. ❌ `github.com/griffnb/techboss-ai-go/internal/integrations/cloudflare`
4. ❌ `github.com/griffnb/techboss-ai-go/internal/integrations/sendpulse`
5. ❌ `github.com/griffnb/techboss-ai-go/internal/models/account`
6. ❌ `github.com/griffnb/techboss-ai-go/internal/models/admin`
7. ❌ `github.com/griffnb/techboss-ai-go/internal/models/agent`
8. ❌ `github.com/griffnb/techboss-ai-go/internal/models/agent_attribute`
9. ❌ `github.com/griffnb/techboss-ai-go/internal/models/ai_tool`
10. ❌ `github.com/griffnb/techboss-ai-go/internal/models/billing_plan`
11. ❌ `github.com/griffnb/techboss-ai-go/internal/models/category`
12. ❌ `github.com/griffnb/techboss-ai-go/internal/models/conversation`
13. ❌ `github.com/griffnb/techboss-ai-go/internal/models/document`
14. ❌ `github.com/griffnb/techboss-ai-go/internal/models/global_config`
15. ❌ `github.com/griffnb/techboss-ai-go/internal/models/lead`
16. ❌ `github.com/griffnb/techboss-ai-go/internal/models/message`
17. ❌ `github.com/griffnb/techboss-ai-go/internal/models/object_tag`
18. ❌ `github.com/griffnb/techboss-ai-go/internal/models/organization`
19. ❌ `github.com/griffnb/techboss-ai-go/internal/models/subscription`
20. ❌ `github.com/griffnb/techboss-ai-go/internal/models/tag`
21. ❌ `github.com/griffnb/techboss-ai-go/internal/services/dynamo_queue`
22. ❌ `github.com/griffnb/techboss-ai-go/internal/services/email_sender`
23. ❌ `github.com/griffnb/techboss-ai-go/internal/services/slack_notifications`

### Actual Test Failures

24. ❌ `github.com/griffnb/techboss-ai-go/internal/services/ai_proxies/openai` - TestProxyRequest_MockOpenAI
   - Error: `Expected /chat/completions path, got /responses`
   - Location: `client_test.go:69`

## Fix Priority

1. **First**: Create missing config file `.configs/unit_test.json` to resolve 23 test failures
2. **Second**: Fix the TestProxyRequest_MockOpenAI test logic issue
