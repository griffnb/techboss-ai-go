package environment

import (
	"os"

	"github.com/griffnb/core/lib/cache"
	"github.com/griffnb/core/lib/dynamo"
	"github.com/griffnb/core/lib/email"
	env "github.com/griffnb/core/lib/environment"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/oauth"
	"github.com/griffnb/core/lib/queue"
	"github.com/griffnb/core/lib/s3"
	"github.com/griffnb/core/lib/tools"
)

const (
	CLIENT_DEFAULT = "default"
)

func IsCI() bool {
	return !tools.Empty(os.Getenv("GITHUB_ACTIONS"))
}

func IsUnitTest() bool {
	return env.Env().GetEnvironment() == "unit_test"
}

func IsProduction() bool {
	return env.Env().IsProduction()
}

func IsLocalDev() bool {
	return env.Env().IsLocalDev()
}

func GetDBClient(name string) *model.Client {
	if tools.Empty(env.Env()) {
		return nil
	}
	if tools.Empty(env.Env().(*SysEnvironment).Databases[name]) {
		return nil
	}

	return env.Env().(*SysEnvironment).Databases[name].(*model.Client)
}

func DB() *model.Client {
	return env.Env().DB().(*model.Client)
}

func GetOauth() *oauth.Authenticator {
	return env.Env().(*SysEnvironment).Oauth
}

func GetLogReader() *log.Reader {
	return env.Env().(*SysEnvironment).LogReader
}

func GetCache() cache.Cache {
	return env.Env().GetCache()
}

func GetConfig() *Config {
	return env.Env().GetConfig().Config.(*Config)
}

func GetEmailer() email.Emailer {
	return env.Env().(*SysEnvironment).Email
}

func GetCachePrefix() string {
	// return env.Env().GetConfig().Config.Cache.Prefix
	return "V1"
}

func GetS3() *s3.S3 {
	return env.Env().(*SysEnvironment).S3
}

func GetQueue() queue.Queue {
	return env.Env().GetQueue()
}

func GetCiphers() map[string]string {
	return GetConfig().Encryption.Ciphers
}

func GetCurrentCipher() string {
	return GetConfig().Encryption.CurrentCipher
}

func GetDynamo() *dynamo.Dynamo {
	return env.Env().GetDynamo()
}

func GetDynamoCacheTable() string {
	return GetConfig().Dynamo.CacheTable
}
