if $1; then
    IMG_TAG=$1
else
    # Fetch current hour into a variable
    current_hour=$(date +"%H")
    current_minute=$(date +"%M")
    IMG_TAG="$current_hour-$current_minute"
fi

IMAGE_NAME="tsmsap/lifecycle-manager:$IMG_TAG"

read -p "Do you want to build and deploy the image $IMAGE_NAME? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    exit 1
fi

# Build the image
make docker-build IMG=$IMAGE_NAME
make docker-push IMG=$IMAGE_NAME
make local-deploy-with-watcher IMG=$IMAGE_NAME
