'use client';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Icons } from '@/components/icons';
import { api } from '@/lib/api-client';
import { useRouter } from 'next/navigation';
import { type FormEvent, useState } from 'react';

export default function SignInPage() {
  const router = useRouter();
  const [email, setEmail] = useState('admin@example.com');
  const [password, setPassword] = useState('changeme123');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError('');
    setIsSubmitting(true);

    try {
      const response = await api.POST('/api/v1/auth/login', {
        body: { email, password },
        credentials: 'include'
      });

      if (response.error) {
        setError('Email atau password tidak valid.');
        setIsSubmitting(false);
        return;
      }

      router.push('/dashboard/overview');
      router.refresh();
    } catch {
      setError('API belum tersedia. Jalankan backend di port 8080.');
      setIsSubmitting(false);
    }
  }

  return (
    <main className='bg-background grid min-h-svh lg:grid-cols-[minmax(0,1.05fr)_minmax(420px,0.95fr)]'>
      <section className='bg-foreground text-background relative hidden overflow-hidden p-10 lg:flex lg:flex-col lg:justify-between'>
        <div className='absolute inset-0 opacity-20 [background-image:linear-gradient(to_right,transparent_49%,currentColor_50%,transparent_51%),linear-gradient(to_bottom,transparent_49%,currentColor_50%,transparent_51%)] [background-size:64px_64px]' />
        <div className='relative'>
          <div className='flex items-center gap-3'>
            <span className='bg-background text-foreground flex size-10 items-center justify-center rounded-xl'>
              <Icons.logo className='size-5' />
            </span>
            <span className='text-sm font-semibold tracking-[0.2em] uppercase'>Gonepost</span>
          </div>
        </div>
        <div className='relative max-w-xl'>
          <p className='text-background/60 mb-5 text-xs font-semibold tracking-[0.24em] uppercase'>Operations control plane</p>
          <h1 className='max-w-lg text-5xl leading-[0.98] font-semibold tracking-tight xl:text-7xl'>
            Clarity for the work behind the work.
          </h1>
          <p className='text-background/65 mt-7 max-w-md text-base leading-7'>
            Manage people, permissions, and the audit trail from one focused workspace.
          </p>
        </div>
        <div className='text-background/55 relative flex items-center gap-2 text-xs'>
          <span className='bg-emerald-400 size-2 rounded-full' />
          Local workspace / API session authentication
        </div>
      </section>

      <section className='flex min-h-svh items-center justify-center px-6 py-12 sm:px-12'>
        <div className='w-full max-w-sm'>
          <div className='mb-10 lg:hidden'>
            <div className='flex items-center gap-3'>
              <span className='bg-foreground text-background flex size-10 items-center justify-center rounded-xl'>
                <Icons.logo className='size-5' />
              </span>
              <span className='text-sm font-semibold tracking-[0.2em] uppercase'>Gonepost</span>
            </div>
          </div>
          <div className='mb-8'>
            <div className='text-muted-foreground mb-4 flex items-center gap-2 text-xs font-semibold tracking-[0.18em] uppercase'>
              <Icons.lock className='size-3.5' /> Secure workspace
            </div>
            <h2 className='text-3xl font-semibold tracking-tight'>Welcome back.</h2>
            <p className='text-muted-foreground mt-3 text-sm leading-6'>Sign in to continue to your operations dashboard.</p>
          </div>
          <form className='grid gap-5' onSubmit={submit}>
            <div className='grid gap-2'>
              <Label htmlFor='email'>Email address</Label>
              <Input id='email' type='email' value={email} onChange={(event) => setEmail(event.target.value)} required />
            </div>
            <div className='grid gap-2'>
              <div className='flex items-center justify-between'>
                <Label htmlFor='password'>Password</Label>
                <span className='text-muted-foreground text-xs'>Local account</span>
              </div>
              <Input id='password' type='password' value={password} onChange={(event) => setPassword(event.target.value)} required />
            </div>
            {error && <p className='bg-destructive/10 text-destructive rounded-lg px-3 py-2 text-sm'>{error}</p>}
            <Button type='submit' disabled={isSubmitting} className='mt-2 h-11 w-full'>
              {isSubmitting ? 'Signing in...' : 'Enter workspace'}
            </Button>
          </form>
          <p className='text-muted-foreground mt-8 text-center text-xs leading-5'>
            Session access is secured by the Gonepost API.
          </p>
        </div>
      </section>
    </main>
  );
}
