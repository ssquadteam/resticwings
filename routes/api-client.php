<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\ResticBackupController;

/*
|--------------------------------------------------------------------------
| Web Routes
|--------------------------------------------------------------------------
|
| Here is where you can register web routes for your application. These
| routes are loaded by the RouteServiceProvider and all of them will
| be assigned to the "web" middleware group. Make something great!
|
*/

Route::middleware([
    'auth:sanctum',
    config('jetstream.auth_middleware'),
    'verified'
])->group(function () {
    Route::get('/dashboard', function () {
        return view('dashboard');
    })->name('dashboard');
});

Route::group(['prefix' => '/servers/{server}/backups/restic'], function () {
    Route::get('/', [ResticBackupController::class, 'index']);
    Route::post('/', [ResticBackupController::class, 'update']);
    Route::get('/snapshots', [ResticBackupController::class, 'listSnapshots']);
    Route::get('/mounts', [ResticBackupController::class, 'listMountPoints']);
    Route::post('/mount', [ResticBackupController::class, 'mount']);
    Route::delete('/mount/{mount}', [ResticBackupController::class, 'unmount']);
}); 