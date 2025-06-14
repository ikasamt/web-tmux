#!/bin/bash

echo "Building Angular frontend..."
cd frontend
npm run build
cd ..

echo "Copying frontend dist to backend..."
rm -rf backend/static
cp -r frontend/dist backend/static

echo "Building Go backend..."
cd backend
go build -o ../web-terminal main.go
cd ..

echo "Cleaning up..."
rm -rf backend/static

echo "Build complete! Run ./web-terminal to start the server."
echo "Access at http://localhost:8080"