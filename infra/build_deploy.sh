# Default values
APP=""
COMMAND_PATH=""
LOCAL_ACCOUNT_ID=""
ENVIRONMENT=""
PORT=8080
TYPE="web"

while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --app)
      APP="$2"
      shift # past argument
      shift # past value
      ;;
    --image)
      COMMAND_PATH="$2"
      shift
      shift
      ;;
    --local-account-id)
      LOCAL_ACCOUNT_ID="$2"
      shift
      shift
      ;;
    --environment)
      ENVIRONMENT="$2"
      shift
      shift
      ;;
    --type)
      TYPE="$2"
      shift
      shift
      ;;
    *)
      echo "Unknown option $1"
      exit 1
      ;;
  esac
done



if [[ -z "$APP" ]]; then
    echo "Error: --app is required"
    exit 1
fi
if [[ -z "$COMMAND_PATH" ]]; then
    echo "Error: --command-path is required"
    exit 1
fi

if [[ -z "$ENVIRONMENT" ]]; then
    echo "Error: --environment is required"
    exit 1
fi
if [[ -z "$TYPE" ]]; then
    echo "Error: --type is required"
    exit 1
fi
if [[ "$TYPE" != "web" && "$TYPE" != "worker" ]]; then
    echo "Error: --type must be either 'web' or 'worker'"
    exit 1
fi


IMAGE_TAG=$(bash -ex build/build_image.sh --app "$APP" --command-path "$COMMAND_PATH" --environment "$ENVIRONMENT" --local-account-id "$LOCAL_ACCOUNT_ID")

if [[ "$TYPE" == "web" ]]; then
    bash -ex build/deploy_web.sh --app "$APP" --image "$IMAGE_TAG" --environment "$ENVIRONMENT" --local-account-id "$LOCAL_ACCOUNT_ID"
else
    bash -ex build/deploy_worker.sh --app "$APP" --image "$IMAGE_TAG" --environment "$ENVIRONMENT" --local-account-id "$LOCAL_ACCOUNT_ID"
fi

