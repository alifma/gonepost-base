'use client';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog';
import { api } from '@/lib/api-client';
import { useEffect, useState } from 'react';
import type { components } from '@gonepost/api-client';
import { toast } from 'sonner';

type User = components['schemas']['User'];

type UserList = components['schemas']['UserListResponse'];
type Role = components['schemas']['Role'];
type UserStatus = components['schemas']['UserStatus'];

export default function ApiUsersPanel() {
  const [users, setUsers] = useState<User[]>([]);
  const [total, setTotal] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [fullName, setFullName] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [isRolesOpen, setIsRolesOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [roles, setRoles] = useState<Role[]>([]);
  const [userRoles, setUserRoles] = useState<Role[]>([]);
  const [editName, setEditName] = useState('');
  const [editUsername, setEditUsername] = useState('');
  const [editStatus, setEditStatus] = useState<UserStatus>('active');
  const [isSaving, setIsSaving] = useState(false);
  const [search, setSearch] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 10;

  async function loadUsers(searchTerm = search, page = currentPage) {
    setIsLoading(true);
    const response = await api.GET('/api/v1/users', {
      credentials: 'include',
      params: { query: { search: searchTerm || undefined, limit: pageSize, offset: (page - 1) * pageSize } }
    });
    if (response.error) {
      setError('Users belum bisa dimuat. Pastikan login dan API aktif.');
    } else {
      const result = response.data as UserList;
      setUsers(result.data);
      setTotal(result.total);
      setError('');
    }
    setIsLoading(false);
  }

  useEffect(() => {
    void loadUsers();
    void loadRoles();
  }, []);

  async function loadRoles() {
    const response = await api.GET('/api/v1/roles', { credentials: 'include' });
    if (!response.error) setRoles(response.data);
  }

  function openEdit(user: User) {
    setSelectedUser(user);
    setEditName(user.full_name ?? '');
    setEditUsername(user.username ?? '');
    setEditStatus(user.status);
    setIsEditOpen(true);
  }

  async function saveUser(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedUser) return;
    setIsSaving(true);
    setError('');
    const response = await api.PATCH('/api/v1/users/{id}', {
      credentials: 'include',
      params: { path: { id: selectedUser.id } },
      body: { full_name: editName, username: editUsername, status: editStatus }
    });
    if (response.error) {
      setError('User gagal diubah. Periksa permission users:write.');
      toast.error('User gagal diubah');
    } else {
      setIsEditOpen(false);
      toast.success('User berhasil diubah');
      await loadUsers();
    }
    setIsSaving(false);
  }

  async function openRoles(user: User) {
    setSelectedUser(user);
    setIsRolesOpen(true);
    const response = await api.GET('/api/v1/users/{id}/roles', {
      credentials: 'include', params: { path: { id: user.id } }
    });
    if (response.error) { setUserRoles([]); setError('Role user gagal dimuat.'); }
    else setUserRoles(Array.isArray(response.data) ? response.data : []);
  }

  async function toggleUserRole(role: Role) {
    if (!selectedUser) return;
    setIsSaving(true);
    const assigned = userRoles.some((currentRole) => currentRole.id === role.id);
    if (assigned && !window.confirm(`Remove ${role.name} from this user?`)) {
      setIsSaving(false);
      return;
    }
    const response = assigned
      ? await api.DELETE('/api/v1/users/{id}/roles/{roleId}', {
          credentials: 'include', params: { path: { id: selectedUser.id, roleId: role.id } }
        })
      : await api.POST('/api/v1/users/{id}/roles', {
          credentials: 'include', params: { path: { id: selectedUser.id } }, body: { role_id: role.id }
        });
    if (response.error) {
      setError('Role user gagal diubah. Backend menolak perubahan ini.');
      toast.error('Role user gagal diubah');
    } else {
      setUserRoles((currentRoles) => assigned
        ? currentRoles.filter((currentRole) => currentRole.id !== role.id)
        : [...currentRoles, role]);
      toast.success(assigned ? 'Role dicabut' : 'Role diberikan');
    }
    setIsSaving(false);
  }

  async function createUser(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsCreating(true);
    setError('');
    const response = await api.POST('/api/v1/users', {
      credentials: 'include',
      body: { email, password, full_name: fullName || undefined }
    });
    if (response.error) {
      setError('User gagal dibuat. Periksa permission atau data input.');
      toast.error('User gagal dibuat');
    } else {
      setEmail('');
      setPassword('');
      setFullName('');
      setIsCreateOpen(false);
      toast.success('User berhasil dibuat');
      await loadUsers();
    }
    setIsCreating(false);
  }

  function submitSearch(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setCurrentPage(1);
    void loadUsers(search, 1);
  }

  function changePage(page: number) {
    setCurrentPage(page);
    void loadUsers(search, page);
  }

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  return (
    <div className='w-full'>
      <section className='border-border/70 bg-card w-full overflow-hidden rounded-xl border'>
        <div className='border-border/70 flex items-center justify-between border-b px-5 py-4'>
          <div>
            <h2 className='font-semibold'>Directory</h2>
            <p className='text-muted-foreground text-sm'>{total} total users</p>
          </div>
          <div className='flex gap-2'>
            <form className='hidden gap-2 md:flex' onSubmit={submitSearch}>
              <Input className='h-9 w-64' placeholder='Search name or email' value={search} onChange={(event) => setSearch(event.target.value)} />
              <Button variant='outline' size='sm' type='submit'>Search</Button>
            </form>
            <Button variant='outline' size='sm' onClick={() => void loadUsers()} disabled={isLoading}>Refresh</Button>
            <Button size='sm' onClick={() => setIsCreateOpen(true)}>Create user</Button>
          </div>
        </div>
        <form className='flex gap-2 border-b px-5 py-4 md:hidden' onSubmit={submitSearch}>
          <Input placeholder='Search name or email' value={search} onChange={(event) => setSearch(event.target.value)} />
          <Button variant='outline' type='submit'>Search</Button>
        </form>
        {error && <p className='text-destructive px-5 pt-4 text-sm'>{error}</p>}
        <div className='overflow-x-auto'>
          <table className='w-full text-sm'>
            <thead className='bg-muted/40 text-muted-foreground text-left'>
              <tr>
                <th className='px-5 py-3 font-medium'>User</th>
                <th className='px-5 py-3 font-medium'>Status</th>
                <th className='px-5 py-3 font-medium'>Created</th>
                <th className='px-5 py-3 font-medium'>Actions</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id} className='border-border/60 border-t'>
                  <td className='px-5 py-3'>
                    <div className='font-medium'>{user.full_name || user.username || 'Unnamed user'}</div>
                    <div className='text-muted-foreground'>{user.email}</div>
                  </td>
                  <td className='px-5 py-3 capitalize'>{user.status}</td>
                  <td className='text-muted-foreground px-5 py-3'>{new Date(user.created_at).toLocaleDateString()}</td>
                  <td className='px-5 py-3'>
                    <div className='flex gap-2'>
                      <Button variant='outline' size='sm' onClick={() => openEdit(user)}>Edit</Button>
                      <Button variant='outline' size='sm' onClick={() => void openRoles(user)}>Roles</Button>
                    </div>
                  </td>
                </tr>
              ))}
              {!isLoading && users.length === 0 && (
                <tr>
                  <td colSpan={4} className='text-muted-foreground px-5 py-8 text-center'>
                    No users found.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        <div className='border-border/70 flex items-center justify-between border-t px-5 py-4'>
          <p className='text-muted-foreground text-sm'>Page {currentPage} of {totalPages}</p>
          <div className='flex gap-2'>
            <Button variant='outline' size='sm' disabled={currentPage <= 1 || isLoading} onClick={() => changePage(currentPage - 1)}>Previous</Button>
            <Button variant='outline' size='sm' disabled={currentPage >= totalPages || isLoading} onClick={() => changePage(currentPage + 1)}>Next</Button>
          </div>
        </div>
      </section>

      <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create user</DialogTitle>
            <DialogDescription>Create an account through the Go API.</DialogDescription>
          </DialogHeader>
          <form className='grid gap-4' onSubmit={createUser}>
          <div className='grid gap-2'>
            <Label htmlFor='full-name'>Full name</Label>
            <Input id='full-name' value={fullName} onChange={(event) => setFullName(event.target.value)} />
          </div>
          <div className='grid gap-2'>
            <Label htmlFor='user-email'>Email</Label>
            <Input id='user-email' type='email' value={email} onChange={(event) => setEmail(event.target.value)} required />
          </div>
          <div className='grid gap-2'>
            <Label htmlFor='user-password'>Temporary password</Label>
            <Input id='user-password' type='password' value={password} onChange={(event) => setPassword(event.target.value)} minLength={8} required />
          </div>
            <DialogFooter>
              <Button type='submit' disabled={isCreating}>{isCreating ? 'Creating...' : 'Create user'}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={isEditOpen} onOpenChange={setIsEditOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit user</DialogTitle>
            <DialogDescription>Update profile fields and account status.</DialogDescription>
          </DialogHeader>
          <form className='grid gap-4' onSubmit={saveUser}>
            <div className='grid gap-2'><Label htmlFor='edit-full-name'>Full name</Label><Input id='edit-full-name' value={editName} onChange={(event) => setEditName(event.target.value)} /></div>
            <div className='grid gap-2'><Label htmlFor='edit-username'>Username</Label><Input id='edit-username' value={editUsername} onChange={(event) => setEditUsername(event.target.value)} /></div>
            <div className='grid gap-2'><Label htmlFor='edit-status'>Status</Label><select id='edit-status' className='border-input bg-background h-9 rounded-lg border px-3 text-sm' value={editStatus} onChange={(event) => setEditStatus(event.target.value as UserStatus)}><option value='active'>Active</option><option value='inactive'>Inactive</option><option value='suspended'>Suspended</option></select></div>
            <DialogFooter><Button type='submit' disabled={isSaving}>{isSaving ? 'Saving...' : 'Save changes'}</Button></DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={isRolesOpen} onOpenChange={setIsRolesOpen}>
        <DialogContent>
          <DialogHeader><DialogTitle>Roles: {selectedUser?.email}</DialogTitle><DialogDescription>Assign or remove roles for this user.</DialogDescription></DialogHeader>
          <div className='grid gap-3'>{roles.map((role) => { const assigned = userRoles.some((currentRole) => currentRole.id === role.id); return <label key={role.id} className='border-border/70 flex items-center gap-3 rounded-lg border p-3 text-sm'><input type='checkbox' checked={assigned} disabled={isSaving} onChange={() => void toggleUserRole(role)} /><span className='flex-1'>{role.name}</span>{role.is_system && <span className='text-muted-foreground text-xs'>System</span>}</label>; })}</div>
          <DialogFooter showCloseButton />
        </DialogContent>
      </Dialog>
    </div>
  );
}
