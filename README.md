[![Logo Image](https://cdn.pterodactyl.io/logos/new/pterodactyl_logo.png)](https://pterodactyl.io)

![Discord](https://img.shields.io/discord/122900397965705216?label=Discord&logo=Discord&logoColor=white)
![GitHub Releases](https://img.shields.io/github/downloads/pterodactyl/wings/latest/total)
[![Go Report Card](https://goreportcard.com/badge/github.com/pterodactyl/wings)](https://goreportcard.com/report/github.com/pterodactyl/wings)

# Pterodactyl Wings with Restic Backup Support

This is a modified version of Pterodactyl Wings that adds support for restic-based backups, including backup mounting capabilities. This modification allows for efficient, encrypted, and deduplicated backups of game servers with the ability to mount and browse backups directly.

## Features

- Full restic backup integration
- Backup mounting support via FUSE
- Deduplication and encryption of backups
- Support for existing S3 and local backup methods
- Per-server restic repositories
- Modern frontend interface for backup management

## Prerequisites

- Linux x86_64 system
- Go 1.22.2 or higher
- Restic installed on the system
- FUSE support for backup mounting
- Pterodactyl Panel installation
- Root access to the server

## Installation

### 1. Install System Dependencies

```bash
# For Ubuntu/Debian
apt update
apt install -y fuse restic

# For CentOS/RHEL
dnf install -y fuse restic

# For Alpine Linux
apk add fuse restic
```

### 2. Build Modified Wings

```bash
# Clone the repository
git clone https://github.com/your-repo/restic-wings.git
cd restic-wings

# Build the binary
go build -o wings

# Stop existing wings service
systemctl stop wings

# Backup existing wings binary
mv /usr/local/bin/wings /usr/local/bin/wings.backup

# Install new wings binary
mv wings /usr/local/bin/wings
chmod +x /usr/local/bin/wings
```

### 3. Configure Wings

Edit your wings configuration file (typically at `/etc/pterodactyl/config.yml`):

```yaml
# Add this section to your existing config.yml
restic:
  enabled: true
  repo_path: /var/lib/pterodactyl/restic-repos
  password: your-secure-repository-password
```

Create required directories:
```bash
mkdir -p /var/lib/pterodactyl/restic-repos
chown -R pterodactyl:pterodactyl /var/lib/pterodactyl/restic-repos
chmod 700 /var/lib/pterodactyl/restic-repos
```

### 4. Panel Modifications

#### 4.1. Install Backend Components

```bash
# Navigate to your panel directory
cd /var/www/pterodactyl

# Copy the controller
cp path/to/ResticBackupController.php app/Http/Controllers/Api/Client/Servers/

# Copy the service
cp path/to/ResticBackupService.php app/Services/Servers/

# Update composer autoloader
composer dump-autoload
```

#### 4.2. Database Modifications

```sql
-- Run this in your panel's database
ALTER TABLE servers 
ADD COLUMN backup_config JSON NULL;
```

#### 4.3. Install Frontend Components

```bash
# Navigate to panel resources
cd /var/www/pterodactyl/resources/scripts/components/server/backups/

# Copy frontend components
cp path/to/ResticBackupContainer.tsx .
cp path/to/ResticMountContainer.tsx .

# Install dependencies and rebuild assets
yarn install
yarn build:production

# Clear panel cache
php artisan view:clear
php artisan cache:clear
```

### 5. Configure Permissions

```bash
# Set proper ownership
chown -R www-data:www-data /var/www/pterodactyl/*

# Set proper permissions for storage
chmod -R 755 /var/www/pterodactyl/storage/* 
chmod -R 755 /var/www/pterodactyl/bootstrap/cache
```

### 6. Restart Services

```bash
# Restart wings
systemctl restart wings

# Restart panel workers
php artisan queue:restart

# Restart web server
systemctl restart nginx # or apache2
```

## Usage

### Creating a Backup

1. Navigate to the server's backup page in the panel
2. Click "Create Backup"
3. Select "Restic" as the backup adapter
4. Configure backup settings:
   - Set backup name
   - Choose files to ignore (optional)
   - Set compression level
5. Click "Start Backup"

### Mounting a Backup

1. Go to the backup management page
2. Select a backup to mount
3. Choose a mount point
4. Click "Mount Backup"
5. Access files through the mounted path

### Managing Mounted Backups

1. View all mounted backups in the "Mounted Backups" section
2. Click "Unmount" when finished accessing the backup
3. The mount point will be automatically cleaned up

## Configuration Options

### Wings Configuration

```yaml
restic:
  enabled: true                                  # Enable/disable restic support
  repo_path: /var/lib/pterodactyl/restic-repos  # Base path for repositories
  password: your-secure-password                 # Default repository password
```

### Server-Specific Settings

Each server can have its own restic configuration:

```json
{
  "enabled": true,
  "repo_path": "/var/lib/pterodactyl/restic-repos/server-uuid",
  "password": "server-specific-password"
}
```

## Troubleshooting

### Common Issues

1. **Mount Failed**: Ensure FUSE is properly installed and the user has permissions
   ```bash
   # Check FUSE installation
   fusermount -V
   
   # Add user to fuse group
   usermod -aG fuse pterodactyl
   ```

2. **Permission Denied**: Check directory permissions
   ```bash
   # Fix repository permissions
   chown -R pterodactyl:pterodactyl /var/lib/pterodactyl/restic-repos
   chmod 700 /var/lib/pterodactyl/restic-repos
   ```

3. **Backup Failed**: Check restic installation
   ```bash
   # Verify restic installation
   restic version
   
   # Test repository access
   sudo -u pterodactyl restic -r /path/to/repo check
   ```

### Logs

Check these locations for troubleshooting:
- Wings logs: `/var/log/pterodactyl/wings.log`
- Panel logs: `/var/www/pterodactyl/storage/logs/laravel.log`
- System logs: `journalctl -u wings`

## Security Considerations

1. Always use strong passwords for restic repositories
2. Regularly rotate repository passwords
3. Keep the restic binary updated
4. Monitor mounted backups and unmount when not in use
5. Implement proper backup rotation policies

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This modification is released under the same license as Pterodactyl Wings.

## Support

For support:
1. Check the troubleshooting guide above
2. Review existing GitHub issues
3. Create a new issue with detailed information about your problem

## Acknowledgments

- Pterodactyl Team for the amazing Wings daemon
- Restic Team for their backup tool
- Contributors to this modification
