#!/bin/bash
echo "Building GNSS applications..."

# Set variables
OUTPUT_DIR="../build"
GNSS_PKG="../cmd/gnss"
NTRIP_CLIENT_PKG="../cmd/ntrip-client"
GNSS_APP_NAME="gnss_receiver"
NTRIP_CLIENT_APP_NAME="ntrip-client"

# Create output directory if it doesn't exist
mkdir -p $OUTPUT_DIR

# Build GNSS application
echo "Building GNSS application..."
go build -o $OUTPUT_DIR/$GNSS_APP_NAME $GNSS_PKG

if [ $? -ne 0 ]; then
    echo "GNSS application build failed!"
    exit 1
fi

# Build NTRIP client application
echo "Building NTRIP client application..."
go build -o $OUTPUT_DIR/$NTRIP_CLIENT_APP_NAME $NTRIP_CLIENT_PKG

if [ $? -ne 0 ]; then
    echo "NTRIP client application build failed!"
    exit 1
fi

echo "Build completed successfully!"
echo "Executable locations:"
echo "- GNSS application: $OUTPUT_DIR/$GNSS_APP_NAME"
echo "- NTRIP client: $OUTPUT_DIR/$NTRIP_CLIENT_APP_NAME"
