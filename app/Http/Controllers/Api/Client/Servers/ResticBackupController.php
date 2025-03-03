<?php

namespace Pterodactyl\Http\Controllers\Api\Client\Servers;

use Pterodactyl\Http\Controllers\Api\Client\ClientApiController;
use Pterodactyl\Models\Server;
use Pterodactyl\Services\Servers\ResticBackupService;
use Illuminate\Http\Request;
use Illuminate\Http\Response;

class ResticBackupController extends ClientApiController
{
    private ResticBackupService $backupService;

    public function __construct(ResticBackupService $backupService)
    {
        parent::__construct();
        $this->backupService = $backupService;
    }

    public function index(Request $request, Server $server)
    {
        return response()->json([
            'data' => [
                'enabled' => $server->backup_config['restic']['enabled'] ?? false,
                'repo_path' => $server->backup_config['restic']['repo_path'] ?? '',
                'password' => $server->backup_config['restic']['password'] ?? '',
            ],
        ]);
    }

    public function update(Request $request, Server $server)
    {
        $validated = $request->validate([
            'enabled' => 'required|boolean',
            'repo_path' => 'required_if:enabled,true|string',
            'password' => 'required_if:enabled,true|string',
        ]);

        $server->update([
            'backup_config' => array_merge($server->backup_config ?? [], [
                'restic' => $validated,
            ]),
        ]);

        return $this->index($request, $server);
    }

    public function listSnapshots(Server $server)
    {
        $snapshots = $this->backupService->listSnapshots($server);

        return response()->json([
            'data' => $snapshots,
        ]);
    }

    public function mount(Request $request, Server $server)
    {
        $validated = $request->validate([
            'snapshot_id' => 'required|string',
            'mount_path' => 'required|string',
        ]);

        $mountPoint = $this->backupService->mountSnapshot(
            $server,
            $validated['snapshot_id'],
            $validated['mount_path']
        );

        return response()->json([
            'data' => $mountPoint,
        ]);
    }

    public function unmount(Request $request, Server $server)
    {
        $validated = $request->validate([
            'mount_id' => 'required|string',
        ]);

        $this->backupService->unmountSnapshot($server, $validated['mount_id']);

        return response()->noContent();
    }

    public function listMountPoints(Server $server)
    {
        $mountPoints = $this->backupService->listMountPoints($server);

        return response()->json([
            'data' => $mountPoints,
        ]);
    }
} 