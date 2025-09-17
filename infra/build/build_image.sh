# Default values
APP="${APP:-}"
COMMAND_PATH="${COMMAND_PATH:-}"
LOCAL_ACCOUNT_ID="${LOCAL_ACCOUNT_ID:-}"


if [[ -z "$APP" ]]; then
    echo "Error: App is required"
    exit 1
fi
if [[ -z "$COMMAND_PATH" ]]; then
    echo "Error: Command path is required"
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
REPO=$ACCOUNT.dkr.ecr.$REGION.amazonaws.com/$APP
TAG=${GITHUB_SHA:-$(date +%Y%m%d%H%M%S)}

aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin $ACCOUNT.dkr.ecr.$REGION.amazonaws.com

#go env -w GOSUMDB="off"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o app "./$COMMAND_PATH"
#CGO_ENABLED=0 go build -o app "./$COMMAND_PATH"

docker build --tag $APP .
docker tag $APP $REPO:$TAG
docker push "$REPO:$TAG"

echo "$REPO:$TAG"