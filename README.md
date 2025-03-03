<!-- ResticWings Logo -->
<p align="center">
  <img src="https://pterodactyl.io/logos/pterry.svg" width="150px" />
</p>

# ResticWings: Pterodactyl with Restic Backup Integration

ResticWings is a fork of Pterodactyl's Wings daemon that integrates advanced backup functionality using Restic. This provides a robust, efficient, and secure backup solution for your game servers.

[![Discord](https://img.shields.io/discord/122900397965705216?label=Discord&logo=discord&logoColor=white)](https://pterodactyl.io/discord)
[![License](https://img.shields.io/github/license/ssquadteam/resticwings)](https://github.com/ssquadteam/resticwings/blob/develop/LICENSE)

## Features

- All the features of Pterodactyl Wings
- Restic backup integration for efficient, encrypted, and deduplicating backups
- Support for various storage backends (local, SFTP, S3, etc.)
- Ability to mount backups for easy file access
- Automatic encryption of backups

# Comprehensive Installation Guide for Pterodactyl with Restic Backup Support

This guide will walk you through installing a customized version of Pterodactyl that includes Restic backup integration. This is a fork of the official Pterodactyl project with additional backup capabilities.

## Prerequisites

- A server running Ubuntu 20.04 or newer
- At least 2GB of RAM
- A valid domain name (recommended)
- Root access to your server

## Part 1: System Preparation

### Update your system
```bash
apt update && apt upgrade -y
```

### Install required dependencies
```bash
apt -y install software-properties-common curl apt-transport-https ca-certificates gnupg
```

### Add required repositories
```bash
# Add PHP repository
LC_ALL=C.UTF-8 add-apt-repository -y ppa:ondrej/php

# Add MariaDB repository
curl -sS https://downloads.mariadb.com/MariaDB/mariadb_repo_setup | sudo bash

# Add Redis repository
curl -fsSL https://packages.redis.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/redis.list

# Update apt again
apt update
```

### Install PHP and its extensions
```bash
apt -y install php8.3 php8.3-{cli,gd,mysql,pdo,mbstring,tokenizer,bcmath,xml,fpm,curl,zip}
```

### Install MariaDB (MySQL)
```bash
apt -y install mariadb-server
```

### Install Redis
```bash
apt -y install redis-server
```

### Install additional dependencies
```bash
apt -y install nginx tar unzip git
```

### Install Restic
```bash
apt -y install restic
```

## Part 2: Configure MariaDB

### Secure your MariaDB installation
```bash
mysql_secure_installation
```

Follow the prompts to set a root password and secure your installation.

### Create a database for Pterodactyl
```bash
mysql -u root -p
```

Once logged in, execute these SQL commands:
```sql
CREATE DATABASE panel;
CREATE USER 'pterodactyl'@'127.0.0.1' IDENTIFIED BY 'YourSecurePassword';
GRANT ALL PRIVILEGES ON panel.* TO 'pterodactyl'@'127.0.0.1' WITH GRANT OPTION;
FLUSH PRIVILEGES;
EXIT;
```

Replace `YourSecurePassword` with a secure password.

## Part 3: Install Composer

```bash
curl -sS https://getcomposer.org/installer | sudo php -- --install-dir=/usr/local/bin --filename=composer
```

## Part 4: Install Panel

### Clone the custom panel repository
```bash
mkdir -p /var/www/pterodactyl
cd /var/www/pterodactyl
git clone https://github.com/ssquadteam/panel.git ./
git checkout restic-integration
```

### Install panel dependencies
```bash
composer install --no-dev --optimize-autoloader
```

### Configure environment
```bash
cp .env.example .env
```

### Set application key
```bash
php artisan key:generate --force
```

### Setup database
```bash
php artisan p:environment:setup
php artisan p:environment:database
```

### Set up the mail service
```bash
php artisan p:environment:mail
```

### Run migrations and create the administrator user
```bash
php artisan migrate --seed --force
php artisan p:user:make
```

### Set permissions
```bash
chown -R www-data:www-data /var/www/pterodactyl/*
```

## Part 5: Set up Nginx

Create a new Nginx configuration file:
```bash
cat > /etc/nginx/sites-available/pterodactyl.conf <<EOL
server {
    listen 80;
    server_name your-domain.com;
    
    root /var/www/pterodactyl/public;
    index index.php;
    
    access_log /var/log/nginx/pterodactyl.app-access.log;
    error_log  /var/log/nginx/pterodactyl.app-error.log error;
    
    # allow larger file uploads and longer script runtimes
    client_max_body_size 100m;
    client_body_timeout 120s;
    
    sendfile off;
    
    # SSL Configuration
    # listen 443 ssl http2;
    # ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    # ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    # ssl_session_cache shared:SSL:10m;
    # ssl_protocols TLSv1.2 TLSv1.3;
    # ssl_ciphers 'ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256';
    # ssl_prefer_server_ciphers on;
    
    location / {
        try_files \$uri \$uri/ /index.php?\$query_string;
    }
    
    location ~ \.php$ {
        fastcgi_split_path_info ^(.+\.php)(/.+)$;
        fastcgi_pass unix:/run/php/php8.3-fpm.sock;
        fastcgi_index index.php;
        include fastcgi_params;
        fastcgi_param PHP_VALUE "upload_max_filesize = 100M \n post_max_size=100M";
        fastcgi_param SCRIPT_FILENAME \$document_root\$fastcgi_script_name;
        fastcgi_param HTTP_PROXY "";
        fastcgi_intercept_errors off;
        fastcgi_buffer_size 16k;
        fastcgi_buffers 4 16k;
        fastcgi_connect_timeout 300;
        fastcgi_send_timeout 300;
        fastcgi_read_timeout 300;
    }
    
    location ~ /\.ht {
        deny all;
    }
}
EOL
```

Replace `your-domain.com` with your actual domain.

Enable the site:
```bash
ln -s /etc/nginx/sites-available/pterodactyl.conf /etc/nginx/sites-enabled/pterodactyl.conf
nginx -t && systemctl restart nginx
```

## Part 6: Set up SSL (Optional but Recommended)

Install Certbot:
```bash
apt install certbot python3-certbot-nginx -y
```

Obtain and configure an SSL certificate:
```bash
certbot --nginx -d your-domain.com
```

Replace `your-domain.com` with your actual domain.

## Part 7: Panel Cron Job

Set up the cron job for the panel:
```bash
crontab -e
```

Add the following line:
```
* * * * * php /var/www/pterodactyl/artisan schedule:run >> /dev/null 2>&1
```

## Part 8: Install Wings (Daemon)

### Install required dependencies
```bash
apt -y install curl tar unzip
```

### Install Docker
```bash
curl -sSL https://get.docker.com/ | CHANNEL=stable bash
```

### Add your user to the docker group
```bash
usermod -aG docker $USER
```

### Enable and start Docker
```bash
systemctl enable --now docker
```

### Download the custom Wings binary

First, create the necessary directories:
```bash
mkdir -p /etc/pterodactyl /var/lib/pterodactyl/restic /var/lib/pterodactyl/mounts
```

Download and install the custom Wings binary with Restic support:
```bash
wget https://github.com/ssquadteam/resticwings/releases/latest/download/wings_linux_amd64
chmod +x wings_linux_amd64
mv wings_linux_amd64 /usr/local/bin/wings
```

If there's no release available, you can build it from source:

```bash
# Install Go
wget https://golang.org/dl/go1.22.2.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone and build Wings
git clone https://github.com/ssquadteam/resticwings.git
cd resticwings
git checkout restic-integration
go build -o wings
mv wings /usr/local/bin/wings
chmod +x /usr/local/bin/wings
```

### Configure Wings

Go to your Pterodactyl Panel and navigate to:
1. Admin CP > Nodes
2. Add a New Node
3. Fill in the required information, including the Restic backup options
4. Once created, click on the node and go to the Configuration tab
5. Copy the configuration

Create a configuration file:
```bash
nano /etc/pterodactyl/config.yml
```

Paste the configuration from the panel and save.

### Create a service file for Wings
```bash
cat > /etc/systemd/system/wings.service <<EOL
[Unit]
Description=Pterodactyl Wings Daemon
After=docker.service
Requires=docker.service
PartOf=docker.service

[Service]
User=root
WorkingDirectory=/etc/pterodactyl
LimitNOFILE=4096
PIDFile=/var/run/wings/daemon.pid
ExecStart=/usr/local/bin/wings
Restart=on-failure
StartLimitInterval=180
StartLimitBurst=30
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOL
```

### Start the Wings service
```bash
systemctl enable --now wings
```

## Part 9: Configure and Use Restic Backups

### Configure Restic in the Panel

1. Go to Admin CP > Nodes
2. Edit your node
3. In the Restic Backup Configuration section:
   - Enable Restic Backup: Yes
   - Restic Repository URL: Enter your Restic repository URL (e.g., sftp:user@host:/path/to/repo)
   - Restic Repository Password: Set a secure password for your repository
   - Restic Mount Path: /var/lib/pterodactyl/mounts
   - SSH Private Key: If using SFTP, enter your SSH private key

### Creating Backups with Restic

1. Go to your server's control panel
2. Click on Backups
3. Click Create Backup
4. Select Restic as the Backup Type
5. Fill in the required details:
   - Name for your backup
   - Ignored files (optional)
   - Password (optional, will use the node's password if not specified)
   - SSH Key Path (optional, will use the node's key if not specified)
   - Mount Point (optional, will use the node's mount point if not specified)
6. Click Start Backup

### Mounting Restic Backups

1. Go to your server's control panel
2. Click on Backups
3. Find the Restic backup you want to mount
4. Click the dropdown menu (three dots)
5. Click "Mount Backup"
6. Specify the mount point (or use the default)
7. Click Mount

## Troubleshooting

### Check Wings Logs
```bash
tail -f /var/log/pterodactyl/wings.log
```

### Check Panel Logs
```bash
tail -f /var/www/pterodactyl/storage/logs/laravel-$(date +%Y-%m-%d).log
```

### Common Issues

1. **Restic Repository Not Initialized**
   - Manually initialize the repository:
   ```bash
   RESTIC_PASSWORD="your-password" restic -r your-repo-url init
   ```

2. **SSH Key Permissions**
   - Ensure SSH keys have the correct permissions:
   ```bash
   chmod 600 /path/to/ssh/key
   ```

3. **Mount Issues**
   - Make sure FUSE is installed:
   ```bash
   apt install fuse -y
   ```

4. **Database Migration Issues**
   - Run migrations manually:
   ```bash
   cd /var/www/pterodactyl
   php artisan migrate --force
   ```

## Conclusion

You've successfully installed Pterodactyl with Restic backup integration! This provides a robust, efficient, and secure backup solution for your game servers.

For more information on Restic commands and features, refer to the [official Restic documentation](https://restic.readthedocs.io/).

## License

Pterodactyl® Copyright © 2015 - 2022 Dane Everitt and contributors.

ResticWings modifications Copyright © 2023 SSQuad Team.

Code released under the MIT License.
