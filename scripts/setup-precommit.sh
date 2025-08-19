#!/bin/bash
# Setup script for pre-commit hooks
# This script installs and configures pre-commit hooks for the GHCP Memory Context Server

set -e

echo "🔒 Setting up pre-commit security hooks for GHCP Memory Context Server..."

# Check if Python is available
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is required but not installed. Please install Python 3."
    exit 1
fi

# Check if pip is available
if ! command -v pip3 &> /dev/null; then
    echo "❌ pip3 is required but not installed. Please install pip3."
    exit 1
fi

# Install pre-commit if not already installed
if ! command -v pre-commit &> /dev/null; then
    echo "📦 Installing pre-commit..."
    pip3 install pre-commit
else
    echo "✅ pre-commit is already installed"
fi

# Install golangci-lint if not already installed
if ! command -v golangci-lint &> /dev/null; then
    echo "📦 Installing golangci-lint..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            brew install golangci-lint
        else
            echo "ℹ️  Please install golangci-lint manually: https://golangci-lint.run/usage/install/"
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
    else
        echo "ℹ️  Please install golangci-lint manually: https://golangci-lint.run/usage/install/"
    fi
else
    echo "✅ golangci-lint is already installed"
fi

# Install gitleaks if not already installed
if ! command -v gitleaks &> /dev/null; then
    echo "📦 Installing gitleaks..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            brew install gitleaks
        else
            echo "ℹ️  Please install gitleaks manually: https://github.com/zricethezav/gitleaks"
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux - download latest release
        wget -O gitleaks.tar.gz https://github.com/zricethezav/gitleaks/releases/download/v8.18.0/gitleaks_8.18.0_linux_x64.tar.gz
        tar -xf gitleaks.tar.gz
        sudo mv gitleaks /usr/local/bin/
        rm gitleaks.tar.gz
    else
        echo "ℹ️  Please install gitleaks manually: https://github.com/zricethezav/gitleaks"
    fi
else
    echo "✅ gitleaks is already installed"
fi

# Install pre-commit hooks
echo "🔧 Installing pre-commit hooks..."
pre-commit install

# Install commit-msg hook for conventional commits (optional)
pre-commit install --hook-type commit-msg

# Update hook repositories to latest versions
echo "🔄 Updating pre-commit hook repositories..."
pre-commit autoupdate

# Initialize secrets baseline if it doesn't exist or is empty
if [ ! -s .secrets.baseline ]; then
    echo "🔍 Initializing secrets baseline..."
    detect-secrets scan --baseline .secrets.baseline
fi

# Test the installation
echo "🧪 Testing pre-commit hooks..."
if pre-commit run --all-files; then
    echo "✅ Pre-commit hooks setup completed successfully!"
else
    echo "⚠️  Some pre-commit checks failed. Please review and fix the issues above."
    echo "💡 You can run 'pre-commit run --all-files' to test again."
fi

echo ""
echo "🎉 Pre-commit security hooks are now active!"
echo ""
echo "📚 Usage:"
echo "  • Hooks run automatically on each commit"
echo "  • Run manually: pre-commit run --all-files"
echo "  • Update hooks: pre-commit autoupdate"
echo "  • Skip hooks (emergency): git commit --no-verify"
echo ""
echo "🔒 Security features enabled:"
echo "  • Secret detection (detect-secrets, gitleaks)"
echo "  • Private key detection"
echo "  • Large file blocking"
echo "  • Go code quality checks (golangci-lint)"
echo "  • File format validation"
