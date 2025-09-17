# Default values
APP="${APP:-}"

if [[ -z "$APP" ]]; then
    echo "Error: App is required"
    exit 1
fi



export AWS_DEFAULT_REGION=us-east-1
export ARCHITECTURE=X86_64


if [[ -z "$CI" ]]; then
    if ! (aws sts get-caller-identity); then
        aws configure set sso_account_id $LOCAL_ACCOUNT_ID
        aws configure set sso_role_name AdministratorAccess
        aws configure set region us-east-1
        aws configure set output json
        aws configure set sso_start_url https://xxxxx.awsapps.com/start
        aws configure set sso_region us-east-1
        aws sso login
    fi
fi




ACCOUNT=$( aws sts get-caller-identity --query Account --output text )
REGION=us-east-1


SERVICE=$(aws cloudformation list-exports \
  --query "Exports[?Name=='$APP-fargate-service'].Value" \
  --output text)

aws ecs update-service --no-cli-pager --cluster $APP-cluster --service $SERVICE --force-new-deployment