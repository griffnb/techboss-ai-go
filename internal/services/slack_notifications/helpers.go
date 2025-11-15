package slacknotifications

import (
	"context"
	"fmt"

	"strings"

	"github.com/CrowdShield/go-core/lib/log/trace"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
)

func MarkdownLink(text, url string) string {
	return fmt.Sprintf("<%s|%s>", url, text)
}

func AccountLink(accountObj *account.AccountJoined) string {
	baseAdminURL := environment.GetConfig().Server.AdminURL

	return MarkdownLink(
		accountObj.GetName(),
		fmt.Sprintf("%s/accounts/details/%s", baseAdminURL, accountObj.ID()),
	)
}

func mobileIcon(ctx context.Context) string {
	if isMobile(ctx) {
		return "📱"
	}
	return "🖥️"
}

func isMobile(ctx context.Context) bool {
	mobileKeywords := []string{"Mobile", "Android", "iPhone", "iPad", "iPod", "Windows Phone", "BlackBerry"}
	trace := trace.GetTraceContext(ctx)
	if tools.Empty(trace) {
		return false
	}

	ua := trace.Headers.Get("User-Agent")
	if tools.Empty(ua) {
		return false
	}

	for _, keyword := range mobileKeywords {
		if strings.Contains(strings.ToLower(ua), strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
