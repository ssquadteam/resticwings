import React, { useEffect, useState } from 'react';
import { Button, Card } from '@/components/elements/button/index';
import { Field } from '@/components/elements/field/index';
import { Form } from '@/components/elements/form/index';
import useFlash from '@/plugins/useFlash';
import { ServerContext } from '@/state/server';
import { httpErrorToHuman } from '@/api/http';

interface ResticBackupSettings {
    enabled: boolean;
    repoPath: string;
    password: string;
}

export default () => {
    const [loading, setLoading] = useState(false);
    const [settings, setSettings] = useState<ResticBackupSettings>({
        enabled: false,
        repoPath: '',
        password: '',
    });

    const { addError, clearFlashes } = useFlash();
    const server = ServerContext.useStoreState(state => state.server.data!);

    useEffect(() => {
        clearFlashes('restic:settings');
        
        // Load settings from the API
        // This would need to be implemented in the panel's API
    }, []);

    const submit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        setLoading(true);

        try {
            // Submit settings to the API
            // This would need to be implemented in the panel's API
            
            // Show success message
        } catch (error) {
            console.error(error);
            addError({ message: httpErrorToHuman(error), key: 'restic:settings' });
        } finally {
            setLoading(false);
        }
    };

    return (
        <Form onSubmit={submit}>
            <Card>
                <div className="mb-6">
                    <h2 className="text-2xl mb-2">Restic Backup Settings</h2>
                    <p className="text-sm text-neutral-300">
                        Configure restic backup settings for this server.
                    </p>
                </div>

                <div className="mb-6">
                    <Field
                        name="enabled"
                        label="Enable Restic Backups"
                        description="Enable or disable restic backup support for this server."
                    >
                        <input
                            type="checkbox"
                            checked={settings.enabled}
                            onChange={e => setSettings(s => ({ ...s, enabled: e.target.checked }))}
                        />
                    </Field>
                </div>

                {settings.enabled && (
                    <>
                        <div className="mb-6">
                            <Field
                                name="repoPath"
                                label="Repository Path"
                                description="The path where restic repositories will be stored."
                            >
                                <input
                                    type="text"
                                    value={settings.repoPath}
                                    onChange={e => setSettings(s => ({ ...s, repoPath: e.target.value }))}
                                    className="w-full"
                                />
                            </Field>
                        </div>

                        <div className="mb-6">
                            <Field
                                name="password"
                                label="Repository Password"
                                description="The password used to encrypt the restic repository."
                            >
                                <input
                                    type="password"
                                    value={settings.password}
                                    onChange={e => setSettings(s => ({ ...s, password: e.target.value }))}
                                    className="w-full"
                                />
                            </Field>
                        </div>
                    </>
                )}

                <div className="flex justify-end">
                    <Button type="submit" loading={loading}>
                        Save Changes
                    </Button>
                </div>
            </Card>
        </Form>
    );
}; 