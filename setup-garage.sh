#!/bin/bash
# setup-garage.sh - Initialize Garage layout, buckets, and keys

set -e

echo "Waiting for Garage to start..."
until docker compose exec garage /garage status &>/dev/null; do
    echo "Garage not ready..."
    sleep 2
done

echo "Garage is up. Configuring layout..."

# Extract Node ID from status output (skip header row)
NODE_ID=$(docker compose exec garage /garage status | grep -A2 "HEALTHY NODES" | tail -n1 | awk '{print $1}')

if [ -z "$NODE_ID" ]; then
    echo "Error: Could not determine Garage Node ID"
    docker compose exec garage /garage status
    exit 1
fi

echo "Found Node ID: $NODE_ID"

# Initialize layout if version is 0
LAYOUT_VERSION=$(docker compose exec garage /garage layout show 2>/dev/null | grep "Current cluster layout version:" | awk '{print $NF}')
if [ "$LAYOUT_VERSION" = "0" ] || [ -z "$LAYOUT_VERSION" ]; then
    echo "Initializing layout..."
    docker compose exec garage /garage layout assign -z dc1 -c 100M "$NODE_ID"
    docker compose exec garage /garage layout apply --version 1
else
    echo "Layout already configured (version $LAYOUT_VERSION)."
fi

# Create bucket
echo "Creating bucket 'videos'..."
if ! docker compose exec garage /garage bucket list | grep -q "videos"; then
    docker compose exec garage /garage bucket create videos
else
    echo "Bucket 'videos' already exists."
fi

# Create API key and update .env
echo "Creating API Key..."
if ! docker compose exec garage /garage key list | grep -q "app-key"; then
    KEY_INFO=$(docker compose exec garage /garage key create app-key)
    
    KEY_ID=$(echo "$KEY_INFO" | grep "Key ID" | awk '{print $3}')
    SECRET_KEY=$(echo "$KEY_INFO" | grep "Secret key" | awk '{print $3}')
    
    echo ""
    echo "✅ Garage Initialized Successfully!"
    echo "---------------------------------------------------"
    
    # Update .env file
    touch .env
    sed -i '/^GARAGE_ACCESS_KEY=/d' .env
    sed -i '/^GARAGE_SECRET_KEY=/d' .env
    echo "GARAGE_ACCESS_KEY=$KEY_ID" >> .env
    echo "GARAGE_SECRET_KEY=$SECRET_KEY" >> .env
    
    echo "Keys added to .env"
    
    # Grant bucket permissions to the key
    echo "Granting bucket permissions..."
    docker compose exec garage /garage bucket allow videos --read --write --owner --key app-key
    
    # Configure CORS for browser uploads using AWS CLI
    echo "Configuring CORS for browser uploads..."
    
    # Create CORS configuration file
    cat > /tmp/cors.json << 'CORS_EOF'
{
  "CORSRules": [
    {
      "AllowedHeaders": ["*"],
      "AllowedMethods": ["GET", "PUT", "POST", "DELETE", "HEAD"],
      "AllowedOrigins": ["*"],
      "MaxAgeSeconds": 3000
    }
  ]
}
CORS_EOF

    # Get garage container ID and apply CORS using AWS CLI
    CONTAINER_ID=$(docker compose ps -q garage)
    
    docker run --rm \
        -v /tmp/cors.json:/aws/cors.json \
        --network container:$CONTAINER_ID \
        -e AWS_ACCESS_KEY_ID="$KEY_ID" \
        -e AWS_SECRET_ACCESS_KEY="$SECRET_KEY" \
        -e AWS_DEFAULT_REGION="us-east-1" \
        amazon/aws-cli:latest \
        --endpoint-url http://localhost:3900 \
        s3api put-bucket-cors --bucket videos --cors-configuration file:///aws/cors.json
    
    rm /tmp/cors.json
    echo "CORS configured successfully!"
    
    echo "---------------------------------------------------"
else
    echo "Key 'app-key' already exists."
    if ! grep -q "GARAGE_ACCESS_KEY" .env 2>/dev/null; then
        echo "⚠️  Key exists in Garage but not in .env."
        echo "To regenerate: docker compose exec garage /garage key delete app-key"
    fi
fi
