package environment

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/CrowdShield/go-core/lib/email"
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/dbclient"
	"github.com/CrowdShield/go-core/lib/oauth"

	"github.com/CrowdShield/go-core/lib/awsconfig"
	"github.com/CrowdShield/go-core/lib/config"
	"github.com/CrowdShield/go-core/lib/dao"
	"github.com/CrowdShield/go-core/lib/dynamo"
	baseenv "github.com/CrowdShield/go-core/lib/environment"
	"github.com/CrowdShield/go-core/lib/localstore"
	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/queue"
	"github.com/CrowdShield/go-core/lib/s3"
	"github.com/CrowdShield/go-core/lib/secrets"
	"github.com/CrowdShield/go-core/lib/slack"
	"github.com/CrowdShield/go-core/lib/tools"
)

var (
	IS_CLOUD   = false
	AWS_CONFIG *aws.Config
)

func init() {
	IS_CLOUD = !tools.Empty(os.Getenv("CLOUD"))
}

// SysEnvironment an example environment
type SysEnvironment struct {
	baseenv.BaseEnvironment
	Region     string
	LogReader  *log.Reader
	RequestLog *log.CoreLogger
	S3         *s3.S3
	Oauth      *oauth.Authenticator
	Email      email.Emailer
}

// CreateEnvironment creates the system environment, should be the main function used
func CreateEnvironment() *SysEnvironment {
	envType := os.Getenv("SYS_ENV")
	if tools.Empty(envType) {
		fmt.Fprintf(os.Stderr, "Fatal SYS_ENV Required")
		os.Exit(1)

	}
	region := os.Getenv("REGION")
	if tools.Empty(region) {
		fmt.Fprintf(os.Stderr, "Fatal REGION Required")
		os.Exit(1)

	}
	configFile := os.Getenv("CONFIG_FILE")

	if tools.Empty(configFile) {

		if IS_CLOUD {
			awsConfig, err := awsconfig.New(region)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Fatal aws config failed %v", err.Error())
				os.Exit(1)
			}

			AWS_CONFIG = awsConfig
		}

		configData, configObj := getSecrets(envType, region)
		coreConfigObj := config.NewConfigFromMap(envType, configData)
		coreConfigObj.Config = configObj
		sysEnv, err := NewEnvironment(envType, region, coreConfigObj)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FATAL %v", err)
			os.Exit(1)

		}
		return sysEnv
	}

	sysEnv, err := NewEnvironmentFromConfigFile(envType, region, configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL %v", err)
		os.Exit(1)

	}
	return sysEnv
}

func getSecrets(environment, region string) (map[string]interface{}, *Config) {
	key := os.Getenv("SECRET_KEY")
	skey := os.Getenv("SECRET_SKEY")

	var secretsClient *secrets.SecretsClient
	var err error
	if tools.Empty(key) && tools.Empty(skey) {
		secretsClient, err = secrets.NewWithAWSConfig(AWS_CONFIG)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FATAL %v", err)
			os.Exit(1)
			return nil, nil
		}
	} else {
		secretsClient, err = secrets.NewWithKeys(region, key, skey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FATAL %v", err)
			os.Exit(1)
			return nil, nil
		}
	}

	secretValue, err := secretsClient.GetSecret(fmt.Sprintf("%v/config", environment))
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL %v", err)
		os.Exit(1)
		return nil, nil
	}
	if tools.Empty(secretValue) {
		fmt.Fprintf(os.Stderr, "FATAL No secrets Found")
		os.Exit(1)
		return nil, nil
	}
	secretMap := make(map[string]interface{})
	configObj := &Config{}
	err = json.Unmarshal([]byte(secretValue), &secretMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL %v", err)
		os.Exit(1)
		return nil, nil
	}

	err = json.Unmarshal([]byte(secretValue), configObj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL %v", err)
		os.Exit(1)
		return nil, nil
	}

	return secretMap, configObj
}

// NewEnvironmentFromConfigFile creates a new env from a config file
func NewEnvironmentFromConfigFile(
	environment, region, configFileLocation string,
) (*SysEnvironment, error) {
	log.Debug(fmt.Sprintf("Loading From Config %v", configFileLocation))

	configObj := &Config{}
	coreConfigObj, err := config.NewConfig(environment, configFileLocation, configObj)
	if err != nil {
		return nil, err
	}
	return NewEnvironment(environment, region, coreConfigObj)
}

// NewEnvironment creates a new Global Environment
func NewEnvironment(
	environment string,
	region string,
	coreConfig *config.Config,
) (*SysEnvironment, error) {
	var err error
	if !tools.Empty(baseenv.Env()) {
		return baseenv.Env().(*SysEnvironment), nil
	}

	env := &SysEnvironment{}
	env.Databases = make(map[string]dbclient.DBClient)
	env.Environment = environment
	env.Region = region
	env.LocalDev = true
	env.Config = coreConfig

	configObj := coreConfig.Config.(*Config)

	if environment == "production" || environment == "staging" {
		env.LocalDev = false
	}

	if !tools.Empty(IS_CLOUD) {
		env.LocalDev = false
	}

	// Load Slack

	slack.BOT_CLIENT.Load(configObj.Slack)

	err = withLogs(env, configObj, region)
	if err != nil {
		return nil, err
	}
	err = withDatabases(env, configObj)
	if err != nil {
		return nil, err
	}
	err = withS3(env, configObj)
	if err != nil {
		return nil, err
	}

	// Setup Local Store
	env.LocalStore = localstore.NewLocalStore()
	err = withOauth(env, configObj)
	if err != nil {
		return nil, err
	}
	err = withQueues(env, configObj, region)
	if err != nil {
		return nil, err
	}
	err = withDynamo(env, configObj, region)
	if err != nil {
		return nil, err
	}

	err = withEmail(env, configObj)
	if err != nil {
		return nil, err
	}

	baseenv.SetSystemEnvironment(env)

	return env, nil
}

func withQueues(env *SysEnvironment, configObj *Config, region string) error {
	// Sets up queues
	if configObj.SQS.Enabled {
		if !tools.Empty(configObj.SQS.LocalEndpoint) {
			log.Debug("Loading SQS Localstack")
			sqsQueue, err := queue.NewSQSQueue(&queue.SQSConfig{
				Region:                 region,
				Endpoint:               configObj.SQS.LocalEndpoint,
				URLs:                   configObj.SQS.Queues,
				Key:                    configObj.SQS.Key,
				Skey:                   configObj.SQS.Skey,
				MessageDeduplicationID: true,
			})
			if err != nil {
				return err
			}

			err = sqsQueue.BuildQueues(context.Background())
			if err != nil {
				return err
			}

			env.Queue = sqsQueue

		} else if !tools.Empty(configObj.SQS.Key) {
			sqsQueue, err := queue.NewSQSQueue(&queue.SQSConfig{
				Region:                 region,
				Key:                    configObj.SQS.Key,
				Skey:                   configObj.SQS.Skey,
				URLs:                   configObj.SQS.Queues,
				MessageDeduplicationID: true,
			})
			if err != nil {
				return err
			}

			err = sqsQueue.BuildQueues(context.Background())
			if err != nil {
				return err
			}

			env.Queue = sqsQueue
		} else {
			sqsQueue, err := queue.NewSQSQueue(&queue.SQSConfig{
				AWSConfig:              AWS_CONFIG,
				URLs:                   configObj.SQS.Queues,
				MessageDeduplicationID: true,
			})
			if err != nil {
				return err
			}

			err = sqsQueue.BuildQueues(context.Background())
			if err != nil {
				return err
			}

			env.Queue = sqsQueue
		}
	} else {
		env.Queue = queue.NewTestQueue()
	}

	return nil
}

func withLogs(env *SysEnvironment, configObj *Config, region string) error {
	cloudWatchConfig := configObj.CloudWatch
	// Setup Logger
	if env.LocalDev {
		log.SetupLocal()
		env.RequestLog = log.NewLogger(os.Stdout, nil, "", true, env.Environment)

	} else {
		groupName := configObj.CloudWatch.GroupName

		log.SetupWithAWSConfig(AWS_CONFIG, groupName, "Default-Errors", slack.BOT_CLIENT, configObj.CloudWatch.ErrorChannel, env.Environment)
		{
			writer, err := log.NewWriterWithAWSConfig(AWS_CONFIG, cloudWatchConfig.RequestGroupName, "Request-Log")
			if err != nil {
				return err
			}
			env.RequestLog = log.NewLogger(writer, slack.BOT_CLIENT, configObj.CloudWatch.ErrorChannel, false, env.Environment)
		}
	}

	// Load Cloudwatch reader
	if !tools.Empty(cloudWatchConfig.Key) {
		reader, err := log.NewReaderWithKeys(region, cloudWatchConfig.Key, cloudWatchConfig.Skey)
		if err != nil {
			return err
		}
		env.LogReader = reader
	} else {
		if IS_CLOUD {
			reader, err := log.NewReaderWithAWSConfig(AWS_CONFIG)
			if err != nil {
				return err
			}
			env.LogReader = reader
		}
	}

	return nil
}

func withDatabases(env *SysEnvironment, configObj *Config) error {
	for key, config := range configObj.Databases {
		db, err := dao.NewDataStore(1, config)
		if err != nil {
			return err
		}

		if key == "default" {
			defaultClient := model.NewClient(key, db)
			env.SetDefaultDatabase(defaultClient)
			model.SetDefaultClient(defaultClient)
		} else {
			env.Databases[key] = model.NewClient(key, db)
		}

	}

	return nil
}

func withDynamo(env *SysEnvironment, configObj *Config, region string) error {
	// Sets up dynamo
	var err error

	if env.LocalDev {
		env.Dynamo, err = dynamo.NewDynamoWithEndpoint(
			region,
			configObj.Dynamo.Key,
			configObj.Dynamo.Skey,
			configObj.Dynamo.Endpoint,
		)
		if err != nil {
			return err
		}
	} else {
		if !tools.Empty(configObj.Dynamo.Key) {
			env.Dynamo, err = dynamo.NewDynamoWithKeys(region, configObj.Dynamo.Key, configObj.Dynamo.Skey)
			if err != nil {
				return err
			}
		} else {
			env.Dynamo, err = dynamo.NewDynamoWithAWSConfig(AWS_CONFIG)
			if err != nil {
				return err
			}
		}
	}

	env.Dynamo.SetupSession(
		configObj.Dynamo.SessionTable,
		configObj.Dynamo.SessionKey,
		configObj.Dynamo.SessionData,
		configObj.Dynamo.SessionExpirationDays,
	)

	// use dynamo for sessions, its cheaper
	env.SessionStore = env.Dynamo

	dynamoCache := dynamo.NewDynamoCacheWithConnection("generic_cache", env.Dynamo)

	err = dynamoCache.CreateTable()
	if err != nil {
		return err
	}
	env.DataCache = dynamoCache

	return nil
}

func withS3(env *SysEnvironment, configObj *Config) error {
	var err error
	if !tools.Empty(configObj.S3Config) {
		if !IS_CLOUD {
			env.S3, err = s3.NewWithKeys(
				configObj.S3Config.Region,
				configObj.S3Config.Key,
				configObj.S3Config.Skey,
			)
			if err != nil {
				return err
			}
		} else {
			env.S3, err = s3.NewWithAWSConfig(AWS_CONFIG)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func withOauth(env *SysEnvironment, configObj *Config) error {
	var err error

	if !tools.Empty(configObj.Oauth) {
		env.Oauth, err = oauth.New(configObj.Oauth)
		if err != nil {
			return err
		}
	}

	return nil
}

func withEmail(env *SysEnvironment, configObj *Config) error {
	var err error
	// Sets up Emailer
	if !tools.Empty(configObj.Email) {
		switch configObj.Email.Provider {
		case "ses":
			if !IS_CLOUD {
				env.Email, err = email.NewWithKeys(
					configObj.Email.SES.Region,
					configObj.Email.SES.Key,
					configObj.Email.SES.Skey,
				)
				if err != nil {
					return err
				}
			} else {
				env.Email, err = email.NewWithAWSConfig(AWS_CONFIG)
				if err != nil {
					return err
				}
			}

		case "smtp":
			env.Email = email.NewSMTP(
				configObj.Email.SMTP.UserName,
				configObj.Email.SMTP.Password,
				configObj.Email.SMTP.Host,
				configObj.Email.SMTP.Port,
			)

		}
	}

	return nil
}
