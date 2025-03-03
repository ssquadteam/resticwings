<?php

namespace Pterodactyl\Services\Servers;

use Pterodactyl\Models\Server;
use Pterodactyl\Exceptions\Service\Backup\BackupNotFoundException;
use Illuminate\Support\Facades\Http;

class ResticBackupService
{
    public function listSnapshots(Server $server): array
    {
        $response = Http::get($this->getWingsUrl($server) . '/api/servers/' . $server->uuid . '/backups/restic/snapshots');

        if (!$response->successful()) {
            throw new BackupNotFoundException;
        }

        return $response->json('data', []);
    }

    public function mountSnapshot(Server $server, string $snapshotId, string $mountPath): array
    {
        $response = Http::post($this->getWingsUrl($server) . '/api/servers/' . $server->uuid . '/backups/restic/mount', [
            'snapshot_id' => $snapshotId,
            'mount_path' => $mountPath,
        ]);

        if (!$response->successful()) {
            throw new BackupNotFoundException;
        }

        return $response->json('data');
    }

    public function unmountSnapshot(Server $server, string $mountId): void
    {
        $response = Http::delete($this->getWingsUrl($server) . '/api/servers/' . $server->uuid . '/backups/restic/mount/' . $mountId);

        if (!$response->successful()) {
            throw new BackupNotFoundException;
        }
    }

    public function listMountPoints(Server $server): array
    {
        $response = Http::get($this->getWingsUrl($server) . '/api/servers/' . $server->uuid . '/backups/restic/mounts');

        if (!$response->successful()) {
            throw new BackupNotFoundException;
        }

        return $response->json('data', []);
    }

    private function getWingsUrl(Server $server): string
    {
        return $server->node->getConnectionAddress();
    }
} 