package slacknotifications

import "github.com/griffnb/techboss-ai-go/internal/environment"

var (
	DEV_CHANNEL = "dev-test"

	NICK_SLACK = "<@U06B5JK45H6>"
)

func channel(channel string) string {
	if !environment.IsProduction() {
		return DEV_CHANNEL
	}
	return channel
}
