'use client';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { api } from '@/lib/api-client';
import type { components } from '@gonepost/api-client';
import { useEffect, useState } from 'react';

type AuditLog = components['schemas']['AuditLogEvent'];

export default function ApiAuditLogPanel() {
  const [events, setEvents] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [action, setAction] = useState('');
  const [actorUserId, setActorUserId] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const pageSize = 20;

  async function loadEvents(actionFilter = action, actorFilter = actorUserId, page = currentPage) {
    setIsLoading(true);
    const response = await api.GET('/api/v1/audit-logs', {
      credentials: 'include',
      params: {
        query: {
          action: actionFilter || undefined,
          actor_user_id: actorFilter || undefined,
          limit: pageSize,
          offset: (page - 1) * pageSize
        }
      }
    });
    if (response.error) {
      setError('Audit logs belum bisa dimuat. Pastikan permission audit:read tersedia.');
    } else {
      setEvents(response.data.data);
      setTotal(response.data.total);
      setError('');
    }
    setIsLoading(false);
  }

  useEffect(() => {
    void loadEvents();
  }, []);

  function submitFilters(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setCurrentPage(1);
    void loadEvents(action, actorUserId, 1);
  }

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  return (
    <section className='border-border/70 bg-card overflow-hidden rounded-xl border'>
      <div className='border-border/70 flex items-center justify-between border-b px-5 py-4'>
        <div>
          <h2 className='font-semibold'>Audit trail</h2>
          <p className='text-muted-foreground text-sm'>Recent events recorded by the Go API.</p>
        </div>
        <form className='border-border/70 grid gap-2 border-b px-5 py-4 sm:grid-cols-[1fr_1fr_auto]' onSubmit={submitFilters}>
          <Input placeholder='Filter action, e.g. user.create' value={action} onChange={(event) => setAction(event.target.value)} />
          <Input placeholder='Filter actor user ID' value={actorUserId} onChange={(event) => setActorUserId(event.target.value)} />
          <Button type='submit' variant='outline'>Apply filters</Button>
        </form>
        <Button variant='outline' size='sm' onClick={() => void loadEvents()} disabled={isLoading}>Refresh</Button>
      </div>
      <div className='border-border/70 flex items-center justify-between border-t px-5 py-4'>
        <p className='text-muted-foreground text-sm'>Page {currentPage} of {totalPages} · {total} events</p>
        <div className='flex gap-2'>
          <Button variant='outline' size='sm' disabled={currentPage <= 1 || isLoading} onClick={() => { const page = currentPage - 1; setCurrentPage(page); void loadEvents(action, actorUserId, page); }}>Previous</Button>
          <Button variant='outline' size='sm' disabled={currentPage >= totalPages || isLoading} onClick={() => { const page = currentPage + 1; setCurrentPage(page); void loadEvents(action, actorUserId, page); }}>Next</Button>
        </div>
      </div>
      {error && <p className='text-destructive px-5 pt-4 text-sm'>{error}</p>}
      <div className='overflow-x-auto'>
        <table className='w-full text-sm'>
          <thead className='bg-muted/40 text-muted-foreground text-left'>
            <tr>
              <th className='px-5 py-3 font-medium'>Action</th>
              <th className='px-5 py-3 font-medium'>Resource</th>
              <th className='px-5 py-3 font-medium'>Result</th>
              <th className='px-5 py-3 font-medium'>Time</th>
            </tr>
          </thead>
          <tbody>
            {events.map((event) => (
              <tr key={event.id} className='border-border/60 border-t'>
                <td className='px-5 py-3 font-medium'>{event.action}</td>
                <td className='px-5 py-3'>{event.resource}</td>
                <td className='px-5 py-3 capitalize'>{event.result}</td>
                <td className='text-muted-foreground px-5 py-3'>{new Date(event.created_at).toLocaleString()}</td>
              </tr>
            ))}
            {!isLoading && events.length === 0 && (
              <tr><td colSpan={4} className='text-muted-foreground px-5 py-8 text-center'>No audit events found.</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </section>
  );
}
