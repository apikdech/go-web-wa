# WhatsApp Profile Image Fetcher - Implementation Plan

## Overview
Build a Go application that fetches WhatsApp profile images and sends them to Discord via webhook, designed to run as a scheduled job on Google Cloud Run.

## Architecture Components

### 1. Authentication & Session Management
**Challenge**: WhatsApp Web requires initial pairing (QR code or phone number pairing)
**Solution**: 
- Use file-based session storage for simplicity
- Implement automatic session restoration
- For Cloud Run: Use Google Cloud Storage or mount persistent volume for session data
- Initial setup: Run locally once to pair device, then deploy session data

### 2. Core Application Flow
```
1. Load environment variables
2. Initialize WhatsApp client with session store
3. Connect to WhatsApp (auto-restore session or fail if not paired)
4. Fetch target phone number's profile image
5. Download image to memory
6. Send image to Discord webhook
7. Clean up and exit
```

### 3. Technology Stack
- **Go**: Latest version (1.21+)
- **whatsmeow**: WhatsApp library
- **Discord Webhook**: For image delivery
- **Docker**: For Cloud Run deployment
- **Google Cloud Storage**: For session persistence (optional)

## Implementation Details

### 1. Project Structure
```
go-web-wa/
├── main.go                 # Main application entry point
├── pkg/
│   ├── whatsapp/          # WhatsApp client wrapper
│   │   ├── client.go      # WhatsApp client initialization
│   │   └── session.go     # Session management
│   ├── discord/           # Discord webhook integration
│   │   └── webhook.go     # Discord webhook sender
│   └── config/            # Configuration management
│       └── config.go      # Environment variables
├── sessions/              # Session storage directory
├── go.mod                 # Go dependencies
├── go.sum                 # Go dependencies checksum
├── Dockerfile             # Docker configuration
├── .env.example           # Environment variables template
└── README.md              # Documentation
```

### 2. Required Environment Variables
```bash
# WhatsApp Configuration
TARGET_PHONE_NUMBER=1234567890  # Phone number to fetch profile from
SESSION_FILE_PATH=./sessions/   # Session storage path

# Discord Configuration
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...

# Optional Cloud Storage (for Cloud Run)
GOOGLE_CLOUD_PROJECT=your-project
GOOGLE_CLOUD_BUCKET=your-bucket
```

### 3. Authentication Flow Options

#### Option A: Pre-paired Session (Recommended for Cloud Run)
1. Run application locally once to pair device
2. Upload session files to cloud storage
3. Cloud Run jobs download session data on startup
4. Automatic session restoration

#### Option B: Dynamic Pairing (Complex for scheduled jobs)
1. Generate pairing code programmatically
2. Manual intervention required for first-time setup
3. Not ideal for automated scheduled jobs

**Recommendation**: Use Option A for simplicity and reliability.

### 4. Session Storage Strategy

#### For Local Development:
- File-based storage in `./sessions/` directory
- SQLite database for session metadata

#### For Cloud Run:
- Google Cloud Storage bucket for session files
- Download on startup, upload on session updates
- Fallback to local storage if cloud storage fails

### 5. Error Handling & Logging
- Comprehensive error handling for network issues
- Structured logging for debugging
- Graceful handling of WhatsApp connection failures
- Retry mechanisms for transient failures

### 6. Docker Configuration
```dockerfile
# Multi-stage build for smaller image
FROM golang:1.21-alpine AS builder
# Build application

FROM alpine:latest
# Runtime dependencies
# Copy binary and run
```

### 7. Cloud Run Deployment
- Use Cloud Run Jobs for scheduled execution
- Cloud Scheduler to trigger jobs
- Environment variables via Cloud Run configuration
- Session data via Cloud Storage mounting

## Implementation Steps

### Phase 1: Core Application
1. [ ] Initialize Go project with dependencies
2. [ ] Create WhatsApp client wrapper
3. [ ] Implement basic session management
4. [ ] Add profile image fetching functionality
5. [ ] Create Discord webhook integration
6. [ ] Build main application flow

### Phase 2: Authentication Setup
1. [ ] Implement session file management
2. [ ] Add QR code pairing for initial setup
3. [ ] Test session restoration
4. [ ] Add error handling for authentication failures

### Phase 3: Cloud Integration
1. [ ] Create Dockerfile
2. [ ] Add Google Cloud Storage integration
3. [ ] Configure Cloud Run deployment
4. [ ] Set up Cloud Scheduler

### Phase 4: Testing & Deployment
1. [ ] Local testing with real WhatsApp account
2. [ ] Cloud Run testing
3. [ ] Scheduled job testing
4. [ ] Documentation and deployment guide

## Security Considerations

1. **Session Data Protection**
   - Encrypt session files at rest
   - Use secure Cloud Storage with proper IAM
   - Rotate session data periodically

2. **API Keys & Secrets**
   - Use Google Secret Manager for sensitive data
   - Never commit credentials to version control
   - Use service account authentication

3. **Network Security**
   - HTTPS for all external communications
   - Validate Discord webhook URLs
   - Rate limiting for API calls

## Potential Challenges & Solutions

1. **WhatsApp Rate Limiting**
   - Implement exponential backoff
   - Cache profile images to reduce API calls
   - Monitor for rate limit responses

2. **Session Expiration**
   - Automatic session renewal
   - Notification system for manual re-pairing
   - Fallback mechanisms

3. **Cloud Run Cold Starts**
   - Optimize Docker image size
   - Use Cloud Run min instances for critical jobs
   - Implement health checks

## Success Metrics

- [ ] Successful WhatsApp connection restoration
- [ ] Profile image fetch success rate > 95%
- [ ] Discord webhook delivery success rate > 99%
- [ ] Cold start time < 10 seconds
- [ ] End-to-end execution time < 30 seconds

## Questions for Clarification

1. **Target Phone Number**: Is this a single phone number or multiple numbers?
2. **Image Format**: Any specific image format/size requirements for Discord?
3. **Scheduling Frequency**: How often should this job run?
4. **Failure Handling**: What should happen if the target user has no profile image?
5. **Cloud Resources**: Do you have existing GCP project/resources to use?

## Next Steps

After reviewing this plan, I'll implement the solution step by step, starting with the core Go application and then adding cloud deployment capabilities. 
