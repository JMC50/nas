#!/bin/bash

# RAID Setup Script for NAS Storage
# This script configures RAID for NAS data storage

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}===== NAS RAID Setup Script =====${NC}"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root (use sudo)${NC}"
    exit 1
fi

# Function to print colored messages
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check current disk status
print_info "Current disk configuration:"
lsblk -o NAME,SIZE,TYPE,MOUNTPOINT,FSTYPE

echo ""
echo -e "${YELLOW}Available disks for RAID:${NC}"
echo "  - /dev/sda: 3.6TB (currently empty)"
echo "  - /dev/sdb: 1.8TB (has NTFS partition with data)"
echo ""

# RAID level selection
echo "Select RAID level:"
echo "1) RAID 0 - Striping (Max capacity: ~5.4TB, High performance, No redundancy)"
echo "2) RAID 1 - Mirroring (Max capacity: 1.8TB, Data redundancy, Lower write performance)"
echo ""
read -p "Enter your choice (1 or 2): " raid_choice

case $raid_choice in
    1)
        RAID_LEVEL=0
        RAID_NAME="raid0"
        print_info "Selected RAID 0 (Striping)"
        print_warn "WARNING: RAID 0 provides no redundancy. If one drive fails, all data is lost!"
        ;;
    2)
        RAID_LEVEL=1
        RAID_NAME="raid1"
        print_info "Selected RAID 1 (Mirroring)"
        print_info "Data will be mirrored across both drives for redundancy"
        ;;
    *)
        print_error "Invalid selection"
        exit 1
        ;;
esac

echo ""
read -p "Do you want to proceed? This will ERASE ALL DATA on /dev/sda and /dev/sdb (y/N): " confirm

if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    print_info "Operation cancelled"
    exit 0
fi

# Backup warning for sdb
if mountpoint -q /media/heesung/5E52680E5267E96B; then
    print_warn "Data exists on /dev/sdb. Please backup important data first!"
    read -p "Have you backed up all important data from /dev/sdb? (y/N): " backup_confirm
    if [ "$backup_confirm" != "y" ] && [ "$backup_confirm" != "Y" ]; then
        print_error "Please backup your data first"
        exit 1
    fi
fi

print_info "Installing mdadm if not already installed..."
apt-get update
apt-get install -y mdadm

# Unmount drives if mounted
print_info "Unmounting drives..."
umount /dev/sda 2>/dev/null || true
umount /dev/sdb1 2>/dev/null || true
umount /media/heesung/DATA 2>/dev/null || true
umount /media/heesung/5E52680E5267E96B 2>/dev/null || true

# Stop any existing RAID arrays on these devices
print_info "Stopping any existing RAID arrays..."
mdadm --stop /dev/md0 2>/dev/null || true
mdadm --stop /dev/md127 2>/dev/null || true

# Clear existing RAID superblocks
print_info "Clearing existing RAID metadata..."
mdadm --zero-superblock /dev/sda 2>/dev/null || true
mdadm --zero-superblock /dev/sdb 2>/dev/null || true

# Wipe partition tables
print_info "Wiping partition tables..."
wipefs -a /dev/sda
wipefs -a /dev/sdb

# Create new partition tables
print_info "Creating new partition tables..."
parted -s /dev/sda mklabel gpt
parted -s /dev/sdb mklabel gpt

# Create partitions for RAID
print_info "Creating partitions..."
parted -s /dev/sda mkpart primary 0% 100%
parted -s /dev/sdb mkpart primary 0% 100%

# Set partition type to Linux RAID
parted -s /dev/sda set 1 raid on
parted -s /dev/sdb set 1 raid on

# Wait for kernel to recognize new partitions
sleep 2
partprobe

# Create RAID array
print_info "Creating RAID $RAID_LEVEL array..."
if [ $RAID_LEVEL -eq 0 ]; then
    mdadm --create --verbose /dev/md0 --level=0 --raid-devices=2 /dev/sda1 /dev/sdb1 --assume-clean
else
    mdadm --create --verbose /dev/md0 --level=1 --raid-devices=2 /dev/sda1 /dev/sdb1
fi

# Wait for RAID to initialize
print_info "Waiting for RAID array to initialize..."
sleep 5

# Check RAID status
print_info "RAID array status:"
cat /proc/mdstat

# Create filesystem on RAID array
print_info "Creating ext4 filesystem on RAID array..."
mkfs.ext4 -F /dev/md0

# Create mount point for NAS storage
NAS_MOUNT="/mnt/nas-storage"
print_info "Creating mount point at $NAS_MOUNT..."
mkdir -p $NAS_MOUNT

# Mount the RAID array
print_info "Mounting RAID array..."
mount /dev/md0 $NAS_MOUNT

# Create NAS data directories
print_info "Creating NAS data directories..."
mkdir -p $NAS_MOUNT/nas-data
mkdir -p $NAS_MOUNT/nas-data-admin

# Set proper permissions
print_info "Setting permissions..."
chown -R $(logname):$(logname) $NAS_MOUNT
chmod 755 $NAS_MOUNT

# Save RAID configuration
print_info "Saving RAID configuration..."
mdadm --detail --scan >> /etc/mdadm/mdadm.conf

# Update initramfs
print_info "Updating initramfs..."
update-initramfs -u

# Add to fstab for automatic mounting
print_info "Adding to /etc/fstab for automatic mounting..."
echo "" >> /etc/fstab
echo "# NAS RAID Storage" >> /etc/fstab
echo "/dev/md0 $NAS_MOUNT ext4 defaults 0 2" >> /etc/fstab

# Create systemd service for RAID monitoring
print_info "Setting up RAID monitoring..."
cat > /etc/systemd/system/mdadm-monitor.service << EOF
[Unit]
Description=mdadm RAID monitor
After=mdadm.service

[Service]
Type=forking
ExecStart=/sbin/mdadm --monitor --scan --daemonise --pid-file /var/run/mdadm/monitor.pid
PIDFile=/var/run/mdadm/monitor.pid

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable mdadm-monitor.service
systemctl start mdadm-monitor.service

# Display final status
echo ""
echo -e "${GREEN}===== RAID Setup Complete =====${NC}"
echo ""
print_info "RAID Configuration:"
mdadm --detail /dev/md0
echo ""
print_info "Filesystem Usage:"
df -h $NAS_MOUNT
echo ""
print_info "Next Steps:"
echo "  1. Migrate existing NAS data to $NAS_MOUNT/nas-data"
echo "  2. Update NAS application configuration to use new paths"
echo "  3. Test the setup thoroughly"
echo ""
print_warn "Important: Remember to set up regular backups. RAID is not a backup solution!"