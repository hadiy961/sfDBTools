#!/bin/bash

# Set environment variables for sfDBTools
export sfDBTools_SOURCE_USER="dbaDO"
export sfDBTools_SOURCE_PASSWORD="DataOn24!!"
export sfDBTools_SOURCE_HOST="localhost"
export sfDBTools_SOURCE_PORT="3306"

export sfDBTools_TARGET_USER="dbaDO"
export sfDBTools_TARGET_PASSWORD="DataOn24!!"
export sfDBTools_TARGET_HOST="localhost"
export sfDBTools_TARGET_PORT="3306"

echo "Environment variables set:"
echo "  sfDBTools_SOURCE_USER=$sfDBTools_SOURCE_USER"
echo "  sfDBTools_SOURCE_PASSWORD=******"
echo "  sfDBTools_SOURCE_HOST=$sfDBTools_SOURCE_HOST"
echo "  sfDBTools_SOURCE_PORT=$sfDBTools_SOURCE_PORT"
echo "  sfDBTools_TARGET_USER=$sfDBTools_TARGET_USER"
echo "  sfDBTools_TARGET_PASSWORD=******"
echo "  sfDBTools_TARGET_HOST=$sfDBTools_TARGET_HOST"
echo "  sfDBTools_TARGET_PORT=$sfDBTools_TARGET_PORT"
