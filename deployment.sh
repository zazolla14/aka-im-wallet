#!/bin/bash
set -e

IMAGES_DIR="./build/images"
AWS_ACCOUNT_ID="${AWS_ACCOUNT_ID}"
AWS_REGION="${AWS_REGION}"
ENVIRONMENT="${ENVIRONMENT}"
VERSION="${VERSION}"
BITBUCKET_LOGIN="${BITBUCKET_LOGIN}"
BITBUCKET_TOKEN="${BITBUCKET_TOKEN}"
KUBECONFIG_BASE64="${KUBECONFIG_BASE64}"

images=(
  "wallet-api"
  "msgtransfer"
  "publisher"
)

services=(
  "wallet-api"
  "msgtransfer"
)

checkEnv() {
  if [[ -z "$ENVIRONMENT" || -z "$VERSION" ]]; then
    echo "❌ Missing required ENVIRONMENT or VERSION"
    echo "Usage: ENVIRONMENT=staging VERSION=abc123 ./cicd.sh build|deploy|all"
    exit 1
  fi
}

configure_kube_for_rancher() {
  echo "🔐 Loading kubeconfig from Bitbucket secret..."
  mkdir -p ~/.kube
  echo "$KUBECONFIG_BASE64" | base64 -d > ~/.kube/config
}

configure_aws_and_kube() {
  echo "🔐 Configuring AWS CLI and Kubernetes context..."
  aws configure set aws_access_key_id "$AWS_ACCESS_KEY_ID"
  aws configure set aws_secret_access_key "$AWS_SECRET_ACCESS_KEY"
  aws configure set region "$AWS_REGION"

  aws eks update-kubeconfig --region "$AWS_REGION" --name "$EKS_CLUSTER_NAME"
}

updateConfigMapAndIngress() {
  echo "📦 Updating im wallet configMap YAML..."
  kubectl apply -f deployments/deploy/wallet-config.yaml -n "$ENVIRONMENT"

  echo "📦 Updating im wallet traefik ingress route YAML..."
  shorted="${ENVIRONMENT#aka-}"
  kubectl apply \
    -f "deployments/traefik/${shorted}/middleware-strip.yaml" \
    -f "deployments/traefik/${shorted}/ingressroute.yaml" \
    -n "$ENVIRONMENT"
}

applyCronjobs() {
  echo "📦 Applying CronJobs with image version and namespace: $ENVIRONMENT..."
  SHORT_ENV="${ENVIRONMENT#aka-}"

  for cron in deployments/cronjobs/*.yaml; do
    if [[ -f "$cron" ]]; then
      CRON_NAME=$(basename "$cron" .yaml)
      IMAGE_NAME="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/akachat-v2/im-wallet:${SHORT_ENV}-publisher-${VERSION}"

      echo "🔧 Rendering $cron with image: $IMAGE_NAME"
      RENDERED_FILE="$(mktemp)"
      sed \
        -e "s|\${IMAGE}|${IMAGE_NAME}|g" \
        "$cron" > "$RENDERED_FILE"
      kubectl apply -f "$RENDERED_FILE" -n "$ENVIRONMENT"
      echo "✅ Applied: $CRON_NAME"

      rm -f "$RENDERED_FILE"
    else
      echo "⚠️  No cronjob YAMLs found in deployments/cronjobs/, skipping..."
    fi
  done
}


buildAndPush() {
  echo "🔐 Logging in to AWS ECR..."
  aws ecr get-login-password --region "$AWS_REGION" \
    | docker login --username AWS --password-stdin "${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"

  echo "🚀 Building and pushing Docker images to ECR..."
  SHORT_ENV="${ENVIRONMENT#aka-}"
  for service in "${images[@]}"; do
    SHORT_NAME="${service#openim-}"
    IMAGE_NAME="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/akachat-v2/im-wallet:${SHORT_ENV}-${SHORT_NAME}-${VERSION}"
    DOCKERFILE_PATH="${IMAGES_DIR}/${service}/Dockerfile"
    CMD_SERVICE_DIR="${service}"

    [[ "$CMD_SERVICE_DIR" == openim-rpc-* ]] && CMD_SERVICE_DIR="openim-rpc/${CMD_SERVICE_DIR}"

    if [[ -f "$DOCKERFILE_PATH" ]]; then
      echo "🔨 Building $IMAGE_NAME from $DOCKERFILE_PATH"
      docker build \
        --platform linux/amd64 \
        --build-arg SERVICE_NAME="${service}" \
        --build-arg CMD_SERVICE_DIR="${CMD_SERVICE_DIR}" \
        -f "$DOCKERFILE_PATH" \
        -t "$IMAGE_NAME" .
      echo "📤 Pushing $IMAGE_NAME to ECR..."
      docker push "$IMAGE_NAME"
      echo "✅ $IMAGE_NAME pushed successfully!"
    else
      echo "⚠️  Dockerfile not found for $service, skipping..."
    fi
  done
  echo "🎉 All images built and pushed to AWS ECR!"
}

deploy() {
  # tmp disable during migration to rancher
  # configure_aws_and_kube

  configure_kube_for_rancher
  updateConfigMapAndIngress

  echo "📦 Updating deployment YAMLs with image version and applying to EKS..."
  SHORT_ENV="${ENVIRONMENT#aka-}"
  for service in "${services[@]}"; do
    SHORT_NAME="${service#openim-}"
    IMAGE_NAME="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/akachat-v2/im-wallet:${SHORT_ENV}-${SHORT_NAME}-${VERSION}"
    YAML_FILE="deployments/deploy/${service}-deployment.yaml"
    SER_FILE="deployments/deploy/${service}-service.yaml"

    if [[ -f "$YAML_FILE" ]]; then
      echo "🔧 Rendering $YAML_FILE with image: $IMAGE_NAME and namespace: $ENVIRONMENT"
      RENDERED_FILE="$(mktemp)"
      sed \
        -e "s|\${IMAGE}|${IMAGE_NAME}|g" \
        "$YAML_FILE" > "$RENDERED_FILE"
      kubectl apply -f "$RENDERED_FILE" -n "$ENVIRONMENT"
      echo "✅ Applied: $service"

      rm -f "$RENDERED_FILE"
    else
      echo "⚠️  YAML not found for $service, skipping..."
    fi
  done

  applyCronjobs

  echo "🚀 All services & cronjobs deployed to k8s Rancher!"
}

checkEnv
case "$1" in
  build)
    buildAndPush
    ;;
  deploy)
    deploy
    ;;
  all)
    buildAndPush
    deploy
    ;;
  *)
    echo "Usage: $0 build|deploy|all"
    exit 1
    ;;
esac
