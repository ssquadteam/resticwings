import React, { useEffect, useState } from 'react';
import { Button, Card } from '@/components/elements/button/index';
import { Field } from '@/components/elements/field/index';
import { Form } from '@/components/elements/form/index';
import useFlash from '@/plugins/useFlash';
import { ServerContext } from '@/state/server';
import { httpErrorToHuman } from '@/api/http';

interface MountPoint {
    id: string;
    path: string;
    mounted: boolean;
}

export default () => {
    const [loading, setLoading] = useState(false);
    const [mountPoints, setMountPoints] = useState<MountPoint[]>([]);
    const [selectedSnapshot, setSelectedSnapshot] = useState('');
    const [mountPath, setMountPath] = useState('');
    const [snapshots, setSnapshots] = useState<string[]>([]);

    const { addError, clearFlashes } = useFlash();
    const server = ServerContext.useStoreState(state => state.server.data!);

    useEffect(() => {
        clearFlashes('restic:mount');
        loadSnapshots();
        loadMountPoints();
    }, []);

    const loadSnapshots = async () => {
        try {
            // Load snapshots from the API
            // This would need to be implemented in the panel's API
        } catch (error) {
            console.error(error);
            addError({ message: httpErrorToHuman(error), key: 'restic:mount' });
        }
    };

    const loadMountPoints = async () => {
        try {
            // Load mount points from the API
            // This would need to be implemented in the panel's API
        } catch (error) {
            console.error(error);
            addError({ message: httpErrorToHuman(error), key: 'restic:mount' });
        }
    };

    const handleMount = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        setLoading(true);

        try {
            // Call API to mount the snapshot
            // This would need to be implemented in the panel's API
            
            await loadMountPoints();
        } catch (error) {
            console.error(error);
            addError({ message: httpErrorToHuman(error), key: 'restic:mount' });
        } finally {
            setLoading(false);
        }
    };

    const handleUnmount = async (mountId: string) => {
        setLoading(true);

        try {
            // Call API to unmount
            // This would need to be implemented in the panel's API
            
            await loadMountPoints();
        } catch (error) {
            console.error(error);
            addError({ message: httpErrorToHuman(error), key: 'restic:mount' });
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="mb-6">
            <Card>
                <div className="mb-6">
                    <h2 className="text-2xl mb-2">Mount Restic Backup</h2>
                    <p className="text-sm text-neutral-300">
                        Mount a restic backup snapshot to access its contents.
                    </p>
                </div>

                <Form onSubmit={handleMount}>
                    <div className="mb-6">
                        <Field
                            name="snapshot"
                            label="Snapshot"
                            description="Select the snapshot to mount."
                        >
                            <select
                                value={selectedSnapshot}
                                onChange={e => setSelectedSnapshot(e.target.value)}
                                className="w-full"
                            >
                                <option value="">Select a snapshot...</option>
                                {snapshots.map(snapshot => (
                                    <option key={snapshot} value={snapshot}>
                                        {snapshot}
                                    </option>
                                ))}
                            </select>
                        </Field>
                    </div>

                    <div className="mb-6">
                        <Field
                            name="mountPath"
                            label="Mount Path"
                            description="The path where the backup will be mounted."
                        >
                            <input
                                type="text"
                                value={mountPath}
                                onChange={e => setMountPath(e.target.value)}
                                className="w-full"
                            />
                        </Field>
                    </div>

                    <div className="flex justify-end">
                        <Button type="submit" loading={loading}>
                            Mount Backup
                        </Button>
                    </div>
                </Form>

                {mountPoints.length > 0 && (
                    <div className="mt-6">
                        <h3 className="text-xl mb-2">Mounted Backups</h3>
                        <div className="grid gap-2">
                            {mountPoints.map(point => (
                                <div
                                    key={point.id}
                                    className="flex items-center justify-between p-2 bg-gray-700 rounded"
                                >
                                    <span>{point.path}</span>
                                    <Button
                                        color="red"
                                        size="small"
                                        onClick={() => handleUnmount(point.id)}
                                        disabled={loading}
                                    >
                                        Unmount
                                    </Button>
                                </div>
                            ))}
                        </div>
                    </div>
                )}
            </Card>
        </div>
    );
}; 