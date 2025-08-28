# 📊 Monitoring Guide

Comprehensive monitoring and logging setup for the NAS File Manager system.

## 📋 Table of Contents

- [Monitoring Overview](#monitoring-overview)
- [System Monitoring](#system-monitoring)
- [Application Monitoring](#application-monitoring)
- [Log Management](#log-management)
- [Performance Metrics](#performance-metrics)
- [Alerting Setup](#alerting-setup)
- [Dashboard Creation](#dashboard-creation)
- [Automated Monitoring](#automated-monitoring)

## Monitoring Overview

### Monitoring Stack

| Component | Purpose | Tools |
|-----------|---------|-------|
| **System Metrics** | CPU, Memory, Disk, Network | htop, iostat, netstat |
| **Application Logs** | Service logs and errors | journalctl, logrotate |
| **Health Checks** | Service availability | curl, systemctl |
| **Performance** | Response times, throughput | custom scripts |
| **Storage** | Disk usage, RAID status | df, mdadm |
| **Security** | Failed logins, file access | auth logs, custom logs |

### Monitoring Levels

1. **Real-time**: Immediate alerts for critical issues
2. **Hourly**: Performance trends and anomalies  
3. **Daily**: Usage summaries and health reports
4. **Weekly**: Capacity planning and trend analysis

## System Monitoring

### System Health Script

```bash
#!/bin/bash
# system-monitor.sh - Comprehensive system monitoring

LOG_FILE="/var/log/nas-monitoring.log"
ALERT_EMAIL="admin@example.com"
HOSTNAME=$(hostname)

# Thresholds
CPU_THRESHOLD=80
MEMORY_THRESHOLD=85
DISK_THRESHOLD=90

log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') [$HOSTNAME] $1" >> "$LOG_FILE"
}

send_alert() {
    local subject="$1"
    local message="$2"
    echo "$message" | mail -s "$subject" "$ALERT_EMAIL"
    log_message "ALERT: $subject"
}

check_cpu_usage() {
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//')
    local cpu_int=${cpu_usage%.*}
    
    if [ "$cpu_int" -gt "$CPU_THRESHOLD" ]; then
        send_alert "High CPU Usage" "CPU usage is at ${cpu_usage}% (threshold: ${CPU_THRESHOLD}%)"
    fi
    
    log_message "CPU Usage: ${cpu_usage}%"
}

check_memory_usage() {
    local memory_usage=$(free | grep Mem | awk '{printf "%.0f", ($3/$2) * 100.0}')
    
    if [ "$memory_usage" -gt "$MEMORY_THRESHOLD" ]; then
        send_alert "High Memory Usage" "Memory usage is at ${memory_usage}% (threshold: ${MEMORY_THRESHOLD}%)"
    fi
    
    log_message "Memory Usage: ${memory_usage}%"
}

check_disk_usage() {
    local disk_usage=$(df /mnt/nas-storage | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [ "$disk_usage" -gt "$DISK_THRESHOLD" ]; then
        send_alert "High Disk Usage" "Disk usage is at ${disk_usage}% (threshold: ${DISK_THRESHOLD}%)"
    fi
    
    log_message "Disk Usage: ${disk_usage}%"
}

check_service_status() {
    if ! systemctl is-active --quiet nas-app.service; then
        send_alert "NAS Service Down" "NAS application service is not running"
        # Attempt restart
        systemctl start nas-app.service
        sleep 10
        if systemctl is-active --quiet nas-app.service; then
            log_message "Service restarted successfully"
        else
            log_message "Failed to restart service"
        fi
    else
        log_message "Service Status: Active"
    fi
}

check_application_response() {
    if ! curl -f -s --max-time 10 http://localhost:7777/ > /dev/null; then
        send_alert "NAS Application Not Responding" "Application is not responding to HTTP requests"
        log_message "Application Response: Failed"
    else
        log_message "Application Response: OK"
    fi
}

# Run all checks
log_message "=== System Monitor Check Started ==="
check_cpu_usage
check_memory_usage  
check_disk_usage
check_service_status
check_application_response
log_message "=== System Monitor Check Completed ==="
```

### Storage Monitoring

```bash
#!/bin/bash
# storage-monitor.sh

monitor_storage() {
    local storage_path="/mnt/nas-storage"
    local log_file="/var/log/storage-monitoring.log"
    
    echo "$(date): Storage monitoring started" >> "$log_file"
    
    # Disk usage breakdown
    echo "=== Disk Usage Report ===" >> "$log_file"
    df -h "$storage_path" >> "$log_file"
    
    # Directory sizes
    echo "=== Directory Sizes ===" >> "$log_file"
    du -h "$storage_path"/* 2>/dev/null | sort -hr | head -20 >> "$log_file"
    
    # File counts
    echo "=== File Statistics ===" >> "$log_file"
    echo "Total files: $(find "$storage_path/data" -type f | wc -l)" >> "$log_file"
    echo "Total directories: $(find "$storage_path/data" -type d | wc -l)" >> "$log_file"
    
    # RAID status (if applicable)
    if [ -f /proc/mdstat ]; then
        echo "=== RAID Status ===" >> "$log_file"
        cat /proc/mdstat >> "$log_file"
    fi
}

monitor_storage
```

## Application Monitoring

### Application Performance Monitor

```bash
#!/bin/bash
# app-monitor.sh

APP_URL="http://localhost:7777"
LOG_FILE="/var/log/app-performance.log"

measure_response_time() {
    local endpoint="$1"
    local response_time=$(curl -o /dev/null -s -w "%{time_total}" "$APP_URL$endpoint")
    echo "$response_time"
}

check_api_endpoints() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Health check endpoint
    local health_time=$(measure_response_time "/")
    echo "$timestamp,health,$health_time" >> "$LOG_FILE"
    
    # Auth config endpoint
    local auth_time=$(measure_response_time "/auth/config")
    echo "$timestamp,auth_config,$auth_time" >> "$LOG_FILE"
    
    # System info endpoint (requires token - skip for now)
    # local sysinfo_time=$(measure_response_time "/getSystemInfo?token=dummy")
    # echo "$timestamp,system_info,$sysinfo_time" >> "$LOG_FILE"
    
    # Alert if response time is too high
    if (( $(echo "$health_time > 5.0" | bc -l) )); then
        echo "$(date): WARNING - High response time: ${health_time}s" >> "$LOG_FILE"
    fi
}

# Create CSV header if file doesn't exist
if [ ! -f "$LOG_FILE" ]; then
    echo "timestamp,endpoint,response_time" > "$LOG_FILE"
fi

check_api_endpoints
```

### Database Monitoring

```bash
#!/bin/bash
# database-monitor.sh

DB_PATH="/mnt/nas-storage/database/nas.sqlite"
LOG_FILE="/var/log/database-monitoring.log"

monitor_database() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    if [ -f "$DB_PATH" ]; then
        # Database size
        local db_size=$(du -h "$DB_PATH" | cut -f1)
        
        # Record counts
        local user_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users;" 2>/dev/null || echo "0")
        local log_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM log;" 2>/dev/null || echo "0")
        
        # Database integrity
        local integrity=$(sqlite3 "$DB_PATH" "PRAGMA integrity_check;" 2>/dev/null || echo "error")
        
        # Log statistics
        echo "$timestamp - DB Size: $db_size, Users: $user_count, Logs: $log_count, Integrity: $integrity" >> "$LOG_FILE"
        
        # Check for issues
        if [ "$integrity" != "ok" ]; then
            echo "$timestamp - ERROR: Database integrity check failed" >> "$LOG_FILE"
        fi
        
        # Check database locks
        local lock_count=$(lsof "$DB_PATH" 2>/dev/null | wc -l)
        if [ "$lock_count" -gt 1 ]; then
            echo "$timestamp - WARNING: Multiple database connections detected" >> "$LOG_FILE"
        fi
    else
        echo "$timestamp - ERROR: Database file not found" >> "$LOG_FILE"
    fi
}

monitor_database
```

## Log Management

### Centralized Logging Setup

```bash
#!/bin/bash
# setup-logging.sh

LOG_DIR="/var/log/nas-app"
RSYSLOG_CONF="/etc/rsyslog.d/nas-app.conf"

# Create log directory
mkdir -p "$LOG_DIR"
chown syslog:syslog "$LOG_DIR"

# Configure rsyslog for NAS application
cat > "$RSYSLOG_CONF" << 'EOF'
# NAS Application Logging
if $programname == 'nas-app' then {
    /var/log/nas-app/application.log
    stop
}

# Separate error logs
if $programname == 'nas-app' and $syslogseverity <= 3 then {
    /var/log/nas-app/error.log
    stop
}
EOF

# Restart rsyslog
systemctl restart rsyslog

# Setup log rotation
cat > /etc/logrotate.d/nas-app << 'EOF'
/var/log/nas-app/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0644 syslog syslog
    postrotate
        systemctl reload rsyslog
    endscript
}
EOF

echo "✅ Logging setup complete"
```

### Log Analysis Scripts

```bash
#!/bin/bash
# analyze-logs.sh

LOG_DIR="/var/log/nas-app"
REPORT_FILE="/tmp/log-analysis-$(date +%Y%m%d).txt"

analyze_logs() {
    echo "NAS Log Analysis Report - $(date)" > "$REPORT_FILE"
    echo "=======================================" >> "$REPORT_FILE"
    
    # Error summary
    echo -e "\n🚨 Error Summary (Last 24 hours):" >> "$REPORT_FILE"
    journalctl -u nas-app.service --since "24 hours ago" | grep -i error | wc -l >> "$REPORT_FILE"
    
    # Top error messages
    echo -e "\n🔍 Top Error Messages:" >> "$REPORT_FILE"
    journalctl -u nas-app.service --since "24 hours ago" | grep -i error | sort | uniq -c | sort -nr | head -10 >> "$REPORT_FILE"
    
    # Authentication failures
    echo -e "\n🔒 Authentication Failures:" >> "$REPORT_FILE"
    journalctl -u nas-app.service --since "24 hours ago" | grep -i "auth.*fail\|login.*fail" | wc -l >> "$REPORT_FILE"
    
    # File operations
    echo -e "\n📁 File Operations:" >> "$REPORT_FILE"
    journalctl -u nas-app.service --since "24 hours ago" | grep -E "(upload|download|delete)" | wc -l >> "$REPORT_FILE"
    
    # Service restarts
    echo -e "\n🔄 Service Restarts:" >> "$REPORT_FILE"
    journalctl -u nas-app.service --since "24 hours ago" | grep -i "started\|stopped" | wc -l >> "$REPORT_FILE"
    
    echo -e "\nReport saved to: $REPORT_FILE"
}

analyze_logs
```

## Performance Metrics

### Performance Data Collection

```bash
#!/bin/bash
# collect-metrics.sh

METRICS_DIR="/var/log/nas-metrics"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
CSV_FILE="$METRICS_DIR/metrics-$(date +%Y%m%d).csv"

# Create metrics directory
mkdir -p "$METRICS_DIR"

# Initialize CSV file with headers if it doesn't exist
if [ ! -f "$CSV_FILE" ]; then
    echo "timestamp,cpu_usage,memory_usage,disk_usage,load_avg,active_connections,response_time" > "$CSV_FILE"
fi

collect_system_metrics() {
    # CPU usage
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//')
    
    # Memory usage
    local memory_usage=$(free | grep Mem | awk '{printf "%.1f", ($3/$2) * 100.0}')
    
    # Disk usage  
    local disk_usage=$(df /mnt/nas-storage | awk 'NR==2 {print $5}' | sed 's/%//')
    
    # Load average
    local load_avg=$(uptime | awk -F'load average:' '{ print $2 }' | cut -d, -f1 | sed 's/^ *//')
    
    # Active connections
    local active_connections=$(ss -t | grep :7777 | wc -l)
    
    # Response time
    local response_time=$(curl -o /dev/null -s -w "%{time_total}" http://localhost:7777/)
    
    # Write to CSV
    echo "$TIMESTAMP,$cpu_usage,$memory_usage,$disk_usage,$load_avg,$active_connections,$response_time" >> "$CSV_FILE"
}

collect_system_metrics
```

### Performance Report Generator

```bash
#!/bin/bash
# generate-performance-report.sh

METRICS_DIR="/var/log/nas-metrics"
REPORT_DIR="/var/log/nas-reports"
DATE=$(date +%Y%m%d)

mkdir -p "$REPORT_DIR"

generate_report() {
    local report_file="$REPORT_DIR/performance-report-$DATE.html"
    
    cat > "$report_file" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>NAS Performance Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .metric { margin: 20px 0; }
    </style>
</head>
<body>
    <h1>NAS Performance Report</h1>
    <p>Generated: $(date)</p>
EOF

    # Add system metrics summary
    local csv_file="$METRICS_DIR/metrics-$DATE.csv"
    if [ -f "$csv_file" ]; then
        echo "<div class='metric'>" >> "$report_file"
        echo "<h2>System Metrics Summary</h2>" >> "$report_file"
        echo "<table>" >> "$report_file"
        echo "<tr><th>Metric</th><th>Average</th><th>Maximum</th><th>Minimum</th></tr>" >> "$report_file"
        
        # Calculate averages using awk
        awk -F, 'NR>1 {
            cpu_sum += $2; cpu_max = ($2 > cpu_max) ? $2 : cpu_max; cpu_min = (cpu_min == "" || $2 < cpu_min) ? $2 : cpu_min;
            mem_sum += $3; mem_max = ($3 > mem_max) ? $3 : mem_max; mem_min = (mem_min == "" || $3 < mem_min) ? $3 : mem_min;
            disk_sum += $4; disk_max = ($4 > disk_max) ? $4 : disk_max; disk_min = (disk_min == "" || $4 < disk_min) ? $4 : disk_min;
            count++
        }
        END {
            printf "<tr><td>CPU Usage (%)</td><td>%.1f</td><td>%.1f</td><td>%.1f</td></tr>\n", cpu_sum/count, cpu_max, cpu_min;
            printf "<tr><td>Memory Usage (%)</td><td>%.1f</td><td>%.1f</td><td>%.1f</td></tr>\n", mem_sum/count, mem_max, mem_min;
            printf "<tr><td>Disk Usage (%)</td><td>%.1f</td><td>%.1f</td><td>%.1f</td></tr>\n", disk_sum/count, disk_max, disk_min;
        }' "$csv_file" >> "$report_file"
        
        echo "</table>" >> "$report_file"
        echo "</div>" >> "$report_file"
    fi
    
    echo "</body></html>" >> "$report_file"
    echo "Performance report generated: $report_file"
}

generate_report
```

## Alerting Setup

### Email Alerting Configuration

```bash
#!/bin/bash
# setup-email-alerts.sh

# Install mail utilities
apt-get install -y mailutils postfix

# Configure postfix for local delivery or SMTP relay
# This is a basic configuration - adjust for your email setup

# Create alert configuration
cat > /etc/nas-alerts.conf << 'EOF'
# NAS Alerting Configuration
ALERT_EMAIL=admin@example.com
SMTP_SERVER=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
EOF

chmod 600 /etc/nas-alerts.conf

# Alert function library
cat > /opt/nas-scripts/alert-functions.sh << 'EOF'
#!/bin/bash
# Alert function library

source /etc/nas-alerts.conf

send_email_alert() {
    local subject="$1"
    local message="$2"
    local priority="${3:-normal}"  # normal, high, low
    
    # Prepare email
    {
        echo "Subject: [NAS Alert] $subject"
        echo "Priority: $priority"
        echo "Date: $(date)"
        echo ""
        echo "$message"
        echo ""
        echo "---"
        echo "NAS System: $(hostname)"
        echo "Timestamp: $(date '+%Y-%m-%d %H:%M:%S')"
    } | sendmail "$ALERT_EMAIL"
}

send_slack_alert() {
    local webhook_url="$SLACK_WEBHOOK_URL"  # Set this in config
    local message="$1"
    
    if [ -n "$webhook_url" ]; then
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"NAS Alert: $message\"}" \
            "$webhook_url"
    fi
}

send_combined_alert() {
    local subject="$1"
    local message="$2"
    
    send_email_alert "$subject" "$message"
    send_slack_alert "$subject: $message"
}
EOF

chmod +x /opt/nas-scripts/alert-functions.sh
```

### Monitoring Dashboard

```bash
#!/bin/bash
# create-dashboard.sh

DASHBOARD_DIR="/var/www/nas-dashboard"
WEB_ROOT="/var/www/html"

# Create dashboard directory
mkdir -p "$DASHBOARD_DIR"

# Simple HTML dashboard
cat > "$DASHBOARD_DIR/index.html" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>NAS Monitoring Dashboard</title>
    <meta http-equiv="refresh" content="30">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { text-align: center; color: #333; margin-bottom: 30px; }
        .status-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .status-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .status-title { font-size: 18px; font-weight: bold; margin-bottom: 15px; }
        .metric { display: flex; justify-content: space-between; margin-bottom: 10px; }
        .status-ok { color: #4CAF50; }
        .status-warning { color: #FF9800; }
        .status-error { color: #F44336; }
        .last-update { text-align: center; margin-top: 30px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🖥️ NAS Monitoring Dashboard</h1>
        </div>
        
        <div class="status-grid">
            <div class="status-card">
                <div class="status-title">🔧 System Status</div>
                <div class="metric">
                    <span>Service Status:</span>
                    <span id="service-status" class="status-ok">Active</span>
                </div>
                <div class="metric">
                    <span>Application Response:</span>
                    <span id="app-response" class="status-ok">OK</span>
                </div>
                <div class="metric">
                    <span>Uptime:</span>
                    <span id="uptime">Loading...</span>
                </div>
            </div>
            
            <div class="status-card">
                <div class="status-title">💾 Storage</div>
                <div class="metric">
                    <span>Disk Usage:</span>
                    <span id="disk-usage">Loading...</span>
                </div>
                <div class="metric">
                    <span>Available Space:</span>
                    <span id="available-space">Loading...</span>
                </div>
                <div class="metric">
                    <span>Total Files:</span>
                    <span id="file-count">Loading...</span>
                </div>
            </div>
            
            <div class="status-card">
                <div class="status-title">⚡ Performance</div>
                <div class="metric">
                    <span>CPU Usage:</span>
                    <span id="cpu-usage">Loading...</span>
                </div>
                <div class="metric">
                    <span>Memory Usage:</span>
                    <span id="memory-usage">Loading...</span>
                </div>
                <div class="metric">
                    <span>Load Average:</span>
                    <span id="load-average">Loading...</span>
                </div>
            </div>
            
            <div class="status-card">
                <div class="status-title">📊 Statistics</div>
                <div class="metric">
                    <span>Total Users:</span>
                    <span id="user-count">Loading...</span>
                </div>
                <div class="metric">
                    <span>Active Sessions:</span>
                    <span id="active-sessions">Loading...</span>
                </div>
                <div class="metric">
                    <span>Recent Errors:</span>
                    <span id="error-count">Loading...</span>
                </div>
            </div>
        </div>
        
        <div class="last-update">
            Last updated: <span id="last-update">Loading...</span>
        </div>
    </div>
    
    <script>
        // Auto-refresh dashboard data
        function updateDashboard() {
            document.getElementById('last-update').textContent = new Date().toLocaleString();
            // Add AJAX calls to fetch real data from monitoring endpoints
        }
        
        // Update every 30 seconds
        setInterval(updateDashboard, 30000);
        updateDashboard();
    </script>
</body>
</html>
EOF

# Create symlink to web root
ln -sf "$DASHBOARD_DIR/index.html" "$WEB_ROOT/dashboard.html"

echo "Dashboard created at: http://your-server/dashboard.html"
```

## Automated Monitoring

### Comprehensive Monitoring Setup

```bash
#!/bin/bash
# setup-monitoring.sh

echo "🔧 Setting up comprehensive NAS monitoring..."

# Create monitoring directories
mkdir -p /opt/nas-monitoring/{scripts,logs,reports}
mkdir -p /var/log/nas-monitoring

# Install monitoring scripts
SCRIPT_DIR="/opt/nas-monitoring/scripts"

# Copy all monitoring scripts to the directory
# (Assume scripts are already created above)

# Make all scripts executable
chmod +x "$SCRIPT_DIR"/*.sh

# Setup monitoring cron jobs
(crontab -l 2>/dev/null; cat << 'EOF'
# NAS Monitoring Cron Jobs

# System monitoring every 5 minutes
*/5 * * * * /opt/nas-monitoring/scripts/system-monitor.sh

# Application monitoring every minute  
* * * * * /opt/nas-monitoring/scripts/app-monitor.sh

# Storage monitoring every 15 minutes
*/15 * * * * /opt/nas-monitoring/scripts/storage-monitor.sh

# Database monitoring every hour
0 * * * * /opt/nas-monitoring/scripts/database-monitor.sh

# Performance metrics collection every minute
* * * * * /opt/nas-monitoring/scripts/collect-metrics.sh

# Daily log analysis at 6 AM
0 6 * * * /opt/nas-monitoring/scripts/analyze-logs.sh

# Weekly performance report on Monday at 8 AM
0 8 * * 1 /opt/nas-monitoring/scripts/generate-performance-report.sh
EOF
) | crontab -

echo "✅ Monitoring setup complete"
echo "📊 Dashboard available at: http://your-server/dashboard.html"
echo "📋 Logs available in: /var/log/nas-monitoring/"
echo "📈 Reports available in: /var/log/nas-reports/"
```

---

*For troubleshooting monitoring issues, see [Troubleshooting Guide](troubleshooting.md). For maintenance of monitoring systems, see [Maintenance Guide](maintenance.md).*