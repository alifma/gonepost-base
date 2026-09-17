'use client';

import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { api } from '@/lib/api-client';
import type { components } from '@gonepost/api-client';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';

type Role = components['schemas']['Role'];
const permissions = ['users:read', 'users:write', 'roles:read', 'roles:write', 'audit:read'];

export default function ApiRolesPanel() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [rolePermissions, setRolePermissions] = useState<string[]>([]);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isAccessOpen, setIsAccessOpen] = useState(false);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  async function loadRoles() {
    setIsLoading(true);
    const response = await api.GET('/api/v1/roles', { credentials: 'include' });
    if (response.error) setError('Roles belum bisa dimuat. Pastikan session dan permission tersedia.');
    else { setRoles(response.data); setError(''); }
    setIsLoading(false);
  }

  useEffect(() => { void loadRoles(); }, []);

  async function createRole(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSaving(true);
    setError('');
    const response = await api.POST('/api/v1/roles', {
      credentials: 'include', body: { name, description: description || undefined }
    });
    if (response.error) {
      setError('Role gagal dibuat. Pastikan akun memiliki roles:write.');
      toast.error('Role gagal dibuat');
    } else {
      setName(''); setDescription(''); setIsCreateOpen(false);
      toast.success('Role berhasil dibuat');
      await loadRoles();
    }
    setIsSaving(false);
  }

  async function openAccess(role: Role) {
    setSelectedRole(role);
    setIsAccessOpen(true);
    const response = await api.GET('/api/v1/roles/{id}/permissions', {
      credentials: 'include', params: { path: { id: role.id } }
    });
    if (response.error) { setError('Permission role gagal dimuat.'); setRolePermissions([]); }
    else { setRolePermissions(Array.isArray(response.data) ? response.data : []); setError(''); }
  }

  async function togglePermission(permission: string) {
    if (!selectedRole) return;
    setIsSaving(true);
    const assigned = rolePermissions.includes(permission);
    if (assigned && !window.confirm(`Revoke ${permission} from this role?`)) {
      setIsSaving(false);
      return;
    }
    const response = assigned
      ? await api.DELETE('/api/v1/roles/{id}/permissions/{code}', {
          credentials: 'include', params: { path: { id: selectedRole.id, code: permission } }
        })
      : await api.POST('/api/v1/roles/{id}/permissions', {
          credentials: 'include', params: { path: { id: selectedRole.id } }, body: { code: permission }
        });
    if (response.error) {
      setError(`Permission ${permission} gagal diubah.`);
      toast.error('Permission gagal diubah');
    }
    else {
      setRolePermissions((current) => current.includes(permission)
        ? current.filter((currentPermission) => currentPermission !== permission)
        : [...current, permission]);
      setError('');
      toast.success(assigned ? 'Permission dicabut' : 'Permission diberikan');
    }
    setIsSaving(false);
  }

  return (
    <>
      <section className='border-border/70 bg-card overflow-hidden rounded-xl border'>
        <div className='border-border/70 flex items-center justify-between border-b px-5 py-4'>
          <div><h2 className='font-semibold'>Roles</h2><p className='text-muted-foreground text-sm'>Create roles and manage their API access.</p></div>
          <div className='flex gap-2'><Button variant='outline' size='sm' onClick={() => void loadRoles()} disabled={isLoading}>Refresh</Button><Button size='sm' onClick={() => setIsCreateOpen(true)}>Create role</Button></div>
        </div>
        {error && <p className='text-destructive px-5 pt-4 text-sm'>{error}</p>}
        <div className='grid gap-3 p-5 md:grid-cols-2 xl:grid-cols-3'>
          {roles.map((role) => <article key={role.id} className='border-border/70 rounded-lg border p-4'><div className='flex items-start justify-between gap-3'><h3 className='font-medium'>{role.name}</h3>{role.is_system && <span className='text-muted-foreground text-xs'>System</span>}</div><p className='text-muted-foreground mt-2 text-sm'>{role.description || 'No description'}</p><Button variant='outline' size='sm' className='mt-4' onClick={() => void openAccess(role)}>Manage access</Button></article>)}
          {!isLoading && roles.length === 0 && <p className='text-muted-foreground text-sm'>No roles found.</p>}
        </div>
      </section>

      <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}><DialogContent><DialogHeader><DialogTitle>Create role</DialogTitle><DialogDescription>Add a role that can receive API permissions.</DialogDescription></DialogHeader><form className='grid gap-4' onSubmit={createRole}><div className='grid gap-2'><Label htmlFor='role-name'>Name</Label><Input id='role-name' value={name} onChange={(event) => setName(event.target.value)} required /></div><div className='grid gap-2'><Label htmlFor='role-description'>Description</Label><Input id='role-description' value={description} onChange={(event) => setDescription(event.target.value)} /></div><DialogFooter><Button type='submit' disabled={isSaving}>{isSaving ? 'Creating...' : 'Create role'}</Button></DialogFooter></form></DialogContent></Dialog>
  <Dialog open={isAccessOpen} onOpenChange={setIsAccessOpen}><DialogContent><DialogHeader><DialogTitle>Manage access: {selectedRole?.name}</DialogTitle><DialogDescription>Grant or revoke permissions for this role.</DialogDescription></DialogHeader><div className='grid gap-3'>{permissions.map((permission) => { const granted = rolePermissions.includes(permission); return <label key={permission} className='border-border/70 flex items-center gap-3 rounded-lg border p-3 text-sm'><input type='checkbox' checked={granted} disabled={isSaving} onChange={() => void togglePermission(permission)} /><span>{permission}</span></label>; })}</div><DialogFooter showCloseButton /></DialogContent></Dialog>
    </>
  );
}
