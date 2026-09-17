'use client';

import { Button } from '@/components/ui/button';
import { api } from '@/lib/api-client';
import type { components } from '@gonepost/api-client';
import { useEffect, useState } from 'react';

type AuditLog = components['schemas']['AuditLogEvent'];

export default function ApiAuditLogPanel() {
  const [events, setEvents] = useState<AuditLog[]>([]);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(true);

  async function loadEvents() {
    setIsLoading(true);
    const response = await api.GET('/api/v1/audit-logs', {
      credentials: 'include',
      params: { query: { limit: 50, offset: 0 } }
    });
    if (response.error) {
      setError('Audit logs belum bisa dimuat. Pastikan permission audit:read tersedia.');
    } else {
      setEvents(response.data.data);
      setError('');
    }
    setIsLoading(false);
  }

  useEffect(() => {
    void loadEvents();
  }, []);

  return (
    <section className='border-border/70 bg-card overflow-hidden rounded-xl border'>
      <div className='border-border/70 flex items-center justify-between border-b px-5 py-4'>
        <div>
          <h2 className='font-semibold'>Audit trail</h2>
          <p className='text-muted-foreground text-sm'>Recent events recorded by the Go API.</p>
        </div>
        <Button variant='outline' size='sm' onClick={() => void loadEvents()} disabled={isLoading}>Refresh</Button>
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
