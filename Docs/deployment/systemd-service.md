# 🔧 Systemd Service Setup

Complete guide for setting up the NAS File Manager as a Linux systemd service for automatic startup and management.

## 📋 Table of Contents

- [Overview](#overview)
- [Service Installation](#service-installation)
- [Service Configuration](#service-configuration)
- [Service Management](#service-management)
- [Logging and Monitoring](#logging-and-monitoring)
- [Troubleshooting](#troubleshooting)
- [Advanced Configuration](#advanced-configuration)

## Overview

The NAS File Manager includes a systemd service file (`nas-app.service`) that enables:
- **Automatic startup** on system boot
- **Automatic restart** on application crashes
- **System integration** with Linux service management
- **Centralized logging** through journald
- **Process monitoring** and control

### Benefits of Systemd Service

| Feature | Benefit |
|---------|---------|
| Auto-start | Starts automatically after reboot |
| Auto-restart | Recovers from crashes automatically |
| Process management | Proper daemon behavior |
| Logging integration | Centralized log management |
| Security isolation | Runs under specific user |
| Resource control | Memory/CPU limits (optional) |

## Service Installation

### 1. Prerequisites

Ensure the application is properly installed and configured:

```bash
# Navigate to application directory
cd /home/heesung/NAS

# Verify application can run
npm test  # Should start successfully

# Verify environment configuration
ls -la .env  # Should exist with proper configuration
```

### 2. Install Service File

The application includes a pre-configured service file:

```bash
# Copy service file to systemd directory
sudo cp nas-app.service /etc/systemd/system/

# Reload systemd to recognize new service
sudo systemctl daemon-reload

# Verify service file is recognized
sudo systemctl list-unit-files | grep nas-app
```

### 3. Enable Auto-Start

```bash
# Enable service to start on boot
sudo systemctl enable nas-app.service

# Verify enabled status
sudo systemctl is-enabled nas-app.service
# Should output: enabled
```

### 4. Start Service

```bash
# Start the service immediately
sudo systemctl start nas-app.service

# Verify service is running
sudo systemctl status nas-app.service
```

## Service Configuration

### Default Service File

The included `nas-app.service` file:

```ini
[Unit]
Description=NAS File Management System
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=heesung
Environment=NODE_ENV=production
WorkingDirectory=/home/heesung/NAS
ExecStart=/usr/bin/npm start
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=nas-app

[Install]
WantedBy=multi-user.target
```

### Configuration Sections

#### [Unit] Section
- **Description**: Human-readable service description
- **After**: Service starts after network is available
- **StartLimitIntervalSec=0**: No limit on restart frequency

#### [Service] Section
- **Type=simple**: Service runs in foreground
- **Restart=always**: Automatically restart on failure
- **RestartSec=1**: Wait 1 second between restarts
- **User=heesung**: Run as specific user (change as needed)
- **Environment**: Set Node.js environment
- **WorkingDirectory**: Application root directory
- **ExecStart**: Command to start application
- **StandardOutput/Error**: Send logs to syslog
- **SyslogIdentifier**: Identifier for log entries

#### [Install] Section
- **WantedBy=multi-user.target**: Enable in multi-user mode

### Customizing Service Configuration

#### Change User and Paths

If your setup differs from the default, modify these values:

```ini
[Service]
User=your-username
WorkingDirectory=/path/to/your/nas-app
ExecStart=/usr/bin/npm start
```

#### Add Environment Variables

```ini
[Service]
Environment=NODE_ENV=production
Environment=PORT=7777  
Environment=DATA_PATH=/mnt/nas-storage
```

#### Resource Limits

Add resource constraints:

```ini
[Service]
# Memory limit (2GB)
MemoryLimit=2G

# CPU limit (50% of one core)
CPUQuota=50%

# File descriptor limit
LimitNOFILE=65536

# Maximum processes
LimitNPROC=1024
```

#### Security Enhancements

```ini
[Service]
# Run in private network namespace
PrivateNetwork=false

# Restrict file system access
ProtectSystem=strict
ProtectHome=read-only

# Allow write access to specific directories
ReadWritePaths=/home/heesung/NAS /mnt/nas-storage /tmp

# Prevent privilege escalation
NoNewPrivileges=true

# Restrict system calls
SystemCallFilter=@system-service
```

## Service Management

### Basic Commands

```bash
# Start service
sudo systemctl start nas-app.service

# Stop service
sudo systemctl stop nas-app.service

# Restart service
sudo systemctl restart nas-app.service

# Reload service configuration (after editing .service file)
sudo systemctl daemon-reload
sudo systemctl reload-or-restart nas-app.service

# Check service status
sudo systemctl status nas-app.service

# Enable auto-start on boot
sudo systemctl enable nas-app.service

# Disable auto-start on boot  
sudo systemctl disable nas-app.service

# Check if service is enabled
sudo systemctl is-enabled nas-app.service

# Check if service is active
sudo systemctl is-active nas-app.service
```

### Service Status Information

```bash
# Detailed status with recent log entries
sudo systemctl status nas-app.service -l

# Service properties
sudo systemctl show nas-app.service

# Service dependencies
sudo systemctl list-dependencies nas-app.service
```

### Managing Service on Boot

```bash
# List all enabled services
sudo systemctl list-unit-files --state=enabled | grep nas

# Check boot time
sudo systemd-analyze blame | grep nas-app

# Check service start order
sudo systemd-analyze critical-chain nas-app.service
```

## Logging and Monitoring

### Viewing Logs

#### Real-time Logs
```bash
# Follow service logs in real-time
sudo journalctl -u nas-app.service -f

# Follow logs with timestamps
sudo journalctl -u nas-app.service -f --no-hostname
```

#### Historical Logs
```bash
# View all logs for service
sudo journalctl -u nas-app.service

# View logs from today
sudo journalctl -u nas-app.service --since today

# View logs from last hour
sudo journalctl -u nas-app.service --since "1 hour ago"

# View logs from specific date range
sudo journalctl -u nas-app.service --since "2025-08-01" --until "2025-08-28"

# View last N lines
sudo journalctl -u nas-app.service -n 50
```

#### Log Filtering
```bash
# Only errors and warnings
sudo journalctl -u nas-app.service -p warning

# Only errors
sudo journalctl -u nas-app.service -p err

# Search for specific text
sudo journalctl -u nas-app.service | grep "ERROR"

# JSON output format
sudo journalctl -u nas-app.service -o json-pretty
```

### Log Configuration

#### Configure Log Retention
```bash
# Create journald configuration
sudo mkdir -p /etc/systemd/journald.conf.d

# Set retention policy
cat << EOF | sudo tee /etc/systemd/journald.conf.d/nas-app.conf
[Journal]
# Keep logs for 1 month
MaxRetentionSec=1month

# Limit total log size to 1GB
SystemMaxUse=1G

# Limit individual log file size
SystemMaxFileSize=100M
EOF

# Restart journald to apply changes
sudo systemctl restart systemd-journald
```

#### Export Logs to File
```bash
# Export logs to file
sudo journalctl -u nas-app.service > nas-app.log

# Export logs with specific format
sudo journalctl -u nas-app.service -o short-iso > nas-app-$(date +%Y%m%d).log

# Create daily log export cron job
echo "0 0 * * * root journalctl -u nas-app.service --since 'yesterday' --until 'today' > /var/log/nas-app/daily-$(date +%Y%m%d).log" | sudo tee /etc/cron.d/nas-app-logs
```

### Monitoring Service Health

#### Create Health Check Script
```bash
#!/bin/bash
# /opt/scripts/nas-app-health.sh

SERVICE="nas-app.service"
HEALTH_URL="http://localhost:7777/"
EMAIL="admin@yourdomain.com"

# Check if service is running
if ! systemctl is-active --quiet $SERVICE; then
    echo "$(date): $SERVICE is not running!"
    # Attempt restart
    sudo systemctl start $SERVICE
    sleep 10
    
    # Check if restart successful
    if systemctl is-active --quiet $SERVICE; then
        echo "$(date): $SERVICE restarted successfully"
    else
        echo "$(date): Failed to restart $SERVICE - sending alert"
        # Send email alert (requires mail setup)
        echo "NAS service failed and could not be restarted" | mail -s "NAS Service Alert" $EMAIL
    fi
    exit 1
fi

# Check HTTP endpoint
if ! curl -f -s --max-time 10 "$HEALTH_URL" > /dev/null; then
    echo "$(date): $SERVICE is running but not responding to HTTP requests"
    # Log application logs for debugging
    sudo journalctl -u $SERVICE -n 20
    exit 1
fi

echo "$(date): $SERVICE is healthy"
exit 0
```

#### Add Health Check to Cron
```bash
# Make script executable
chmod +x /opt/scripts/nas-app-health.sh

# Add to cron (every 5 minutes)
echo "*/5 * * * * /opt/scripts/nas-app-health.sh >> /var/log/nas-app-health.log 2>&1" | crontab -
```

## Troubleshooting

### Common Issues

#### Service Fails to Start

**Check service status:**
```bash
sudo systemctl status nas-app.service -l
```

**Common causes and solutions:**

1. **User doesn't exist:**
   ```bash
   # Check if user exists
   id heesung
   
   # If not, create user or change service file
   sudo useradd -r -s /bin/bash heesung
   ```

2. **Working directory doesn't exist:**
   ```bash
   # Check path in service file
   ls -la /home/heesung/NAS
   
   # Fix path in service file if needed
   sudo systemctl edit nas-app.service
   ```

3. **npm not found:**
   ```bash
   # Check npm location
   which npm
   
   # Update ExecStart path in service file
   ExecStart=/usr/local/bin/npm start
   ```

4. **Permission errors:**
   ```bash
   # Fix ownership
   sudo chown -R heesung:heesung /home/heesung/NAS
   
   # Fix permissions
   sudo chmod -R 755 /home/heesung/NAS
   ```

#### Service Starts But Application Fails

**Check application logs:**
```bash
sudo journalctl -u nas-app.service -n 100
```

**Common issues:**

1. **Environment variables missing:**
   ```bash
   # Check .env file exists
   ls -la /home/heesung/NAS/.env
   
   # Add environment variables to service file
   sudo systemctl edit nas-app.service
   ```

2. **Dependencies not installed:**
   ```bash
   # Install dependencies
   cd /home/heesung/NAS
   sudo -u heesung npm install
   ```

3. **Database/file permissions:**
   ```bash
   # Check data directory permissions
   ls -la /mnt/nas-storage/
   
   # Fix permissions
   sudo chown -R heesung:heesung /mnt/nas-storage/
   ```

#### Service Stops Unexpectedly

**Check for crashes:**
```bash
# Look for crash logs
sudo journalctl -u nas-app.service | grep -i "crash\|segfault\|killed"

# Check system logs
sudo dmesg | grep -i "killed\|oom"
```

**Memory issues:**
```bash
# Check memory usage
free -h

# Add memory limit to service
sudo systemctl edit nas-app.service
# Add: MemoryLimit=2G
```

#### Port Already in Use

```bash
# Find process using port 7777
sudo lsof -i :7777
sudo netstat -tulpn | grep 7777

# Kill conflicting process
sudo kill -9 <PID>

# Or change port in configuration
# Edit .env file to use different port
```

### Performance Issues

#### High CPU Usage
```bash
# Monitor service resource usage
sudo systemctl status nas-app.service

# Add CPU limit
sudo systemctl edit nas-app.service
# Add: CPUQuota=50%
```

#### High Memory Usage
```bash
# Check memory usage
sudo systemctl show nas-app.service | grep Memory

# Set memory limit
sudo systemctl edit nas-app.service
# Add: MemoryLimit=2G
```

#### Service Takes Long to Start
```bash
# Check startup time
sudo systemd-analyze blame | grep nas-app

# Increase timeout if needed
sudo systemctl edit nas-app.service
# Add: TimeoutStartSec=300
```

## Advanced Configuration

### Override Default Configuration

Create service override without modifying the original file:

```bash
# Create override directory
sudo systemctl edit nas-app.service

# This opens an editor where you can add overrides
[Service]
User=different-user
Environment=PORT=8888
MemoryLimit=4G
```

### Multiple Environment Configuration

Create different service variants for different environments:

```bash
# Copy service file for staging
sudo cp /etc/systemd/system/nas-app.service /etc/systemd/system/nas-app-staging.service

# Edit for staging configuration
sudo systemctl edit nas-app-staging.service
```

### Integration with Reverse Proxy

Ensure service starts before web server:

```ini
# In nginx.service or apache2.service
[Unit]
After=nas-app.service
Requires=nas-app.service
```

### Dependency Management

Create service dependencies:

```ini
[Unit]
Description=NAS File Management System
After=network.target mysql.service redis.service
Wants=mysql.service redis.service
```

### Custom Start/Stop Scripts

Use custom scripts for complex startup:

```bash
#!/bin/bash
# /opt/scripts/nas-start.sh

# Pre-startup tasks
echo "Starting NAS application..."

# Check dependencies
if ! systemctl is-active --quiet mysql; then
    echo "Starting MySQL..."
    systemctl start mysql
fi

# Start application
cd /home/heesung/NAS
exec npm start
```

Update service file:
```ini
[Service]
ExecStart=/opt/scripts/nas-start.sh
```

## Best Practices

### Security
- Run service as non-root user
- Use minimal file system permissions
- Enable security restrictions in service file
- Regularly update service configuration

### Reliability
- Configure appropriate restart policies
- Set resource limits to prevent system issues
- Implement health checks
- Monitor service performance

### Maintenance
- Regularly review service logs
- Test service restarts
- Keep service configuration in version control
- Document any customizations

### Monitoring
- Set up log rotation
- Implement alerting for service failures
- Monitor resource usage
- Track service uptime metrics

---

*For additional deployment options, see the [Deployment Guide](deployment-guide.md). For configuration details, see [Environment Setup](../configuration/environment-setup.md).*