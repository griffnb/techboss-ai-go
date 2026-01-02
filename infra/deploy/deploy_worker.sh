# Default values
APP="${APP:-}"
IMAGE="${IMAGE:-}"
LOCAL_ACCOUNT_ID="${LOCAL_ACCOUNT_ID:-}"
ENVIRONMENT="${ENVIRONMENT:-}"
PORT="${PORT:-8080}"
SYSTEM_MEMORY="${SYSTEM_MEMORY:-1024}"
SYSTEM_CPU="${SYSTEM_CPU:-512}"
TASK_COUNT="${TASK_COUNT:-1}"
PUBLIC_IP="${PUBLIC_IP:-ENABLED}"


if [[ -z "$APP" ]]; then
    echo "Error: App is required"
    exit 1
fi
if [[ -z "$IMAGE" ]]; then
    echo "Error: Image is required"
    exit 1
fi

if [[ -z "$ENVIRONMENT" ]]; then
    echo "Error: Environment is required"
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
        aws configure set sso_start_url https://xxxx.awsapps.com/start
        aws configure set sso_region us-east-1
        aws sso login
    fi
fi




ACCOUNT=$( aws sts get-caller-identity --query Account --output text )
REGION=us-east-1

aws cloudformation deploy \
    --no-fail-on-empty-changeset \
    --stack-name "$APP-deployment" \
    --template-file "./infra/deploy/stack_worker.yaml" \
    --parameter-overrides "Image=$IMAGE" \
                          "Architecture=$ARCHITECTURE" \
                          "App=$APP" \
                          "ContainerPort=$PORT" \
                          "Environment=$ENVIRONMENT" \
                          "SystemMemory=$SYSTEM_MEMORY" \
                          "SystemCPU=$SYSTEM_CPU" \
                          "TaskCount=$TASK_COUNT" \
                          "AssignPublicIp=$PUBLIC_IP" \


